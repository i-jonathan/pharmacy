package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"pharmacy/httperror"
	"pharmacy/internal/constant"
	"pharmacy/internal/types"
	"pharmacy/repository"
	"strings"
	"unicode/utf8"

	"github.com/jmoiron/sqlx"
)

// Health readings are sensitive: no message here repeats a value, so none ends up in a log.

const (
	healthChecksPerPage = 20
	maxHealthCheckPage  = 1_000_000
	maxHealthNote       = 500
	maxHealthReason     = 200
	maxHealthRetryKey   = 100
)

// healthRange is an allowed span of a reading. The spans catch typos such as 1200/80; they are not
// clinical limits, and nothing here judges a number.
type healthRange struct{ min, max float64 }

var (
	systolicRange  = healthRange{50, 300}
	diastolicRange = healthRange{30, 200}
	pulseRange     = healthRange{20, 250}
	glucoseRanges  = map[string]healthRange{"mmol/L": {1.0, 50.0}, "mg/dL": {20, 900}}
	// glucoseUnits maps a unit, lowercased, to how it is stored.
	glucoseUnits    = map[string]string{"mmol/l": "mmol/L", "mg/dl": "mg/dL"}
	glucoseContexts = map[string]bool{"fasting": true, "after_meal": true, "random": true}
)

func (r healthRange) holds(v float64) bool { return v >= r.min && v <= r.max }

type healthCheckService struct {
	repo repository.PharmacyRepository
}

func NewHealthCheckService(repo repository.PharmacyRepository) *healthCheckService {
	return &healthCheckService{repo: repo}
}

func requirePermissions(ctx context.Context, msg string, keys ...string) error {
	for _, key := range keys {
		if !HasPermission(ctx, key) {
			return httperror.Forbidden(msg, nil)
		}
	}
	return nil
}

func healthCheckNotFound(err error) error {
	return httperror.NotFound("this health check was not found", err)
}

// conflictWithCode is a 409 whose code lets the page act on it without reading the message.
func conflictWithCode(code, msg string) error {
	err := httperror.Conflict(msg, nil)
	err.Reason = code
	return err
}

// normalizeHealthReadings trims and checks a reading. Blood pressure is both numbers (and an
// optional pulse); glucose is a value with its unit and when it was taken; at least one is needed.
func normalizeHealthReadings(h types.HealthReadings) (types.HealthReadings, error) {
	h.Note = strings.TrimSpace(h.Note)
	h.GlucoseUnit = strings.TrimSpace(h.GlucoseUnit)
	h.GlucoseContext = strings.ToLower(strings.TrimSpace(h.GlucoseContext))
	if strings.ContainsRune(h.Note, 0) {
		return h, badRequest("the note can't contain control characters")
	}
	if utf8.RuneCountInString(h.Note) > maxHealthNote {
		return h, badRequest(fmt.Sprintf("the note can be at most %d characters", maxHealthNote))
	}
	if err := checkBloodPressure(h); err != nil {
		return h, err
	}
	h, err := checkGlucose(h)
	if err != nil {
		return h, err
	}
	if h.Systolic == nil && h.Glucose == nil {
		return h, badRequest("enter a blood pressure or a blood glucose reading")
	}
	return h, nil
}

func checkBloodPressure(h types.HealthReadings) error {
	if (h.Systolic == nil) != (h.Diastolic == nil) {
		return badRequest("enter both blood pressure numbers")
	}
	if h.Systolic == nil {
		if h.Pulse != nil {
			return badRequest("a pulse is recorded with a blood pressure reading")
		}
		return nil
	}
	switch {
	case !systolicRange.holds(float64(*h.Systolic)):
		return badRequest("the top blood pressure number must be between 50 and 300; check it")
	case !diastolicRange.holds(float64(*h.Diastolic)):
		return badRequest("the bottom blood pressure number must be between 30 and 200; check it")
	case *h.Systolic <= *h.Diastolic:
		return badRequest("the top blood pressure number must be higher than the bottom one")
	case h.Pulse != nil && !pulseRange.holds(float64(*h.Pulse)):
		return badRequest("the pulse must be between 20 and 250; check it")
	}
	return nil
}

func checkGlucose(h types.HealthReadings) (types.HealthReadings, error) {
	if h.Glucose == nil && h.GlucoseUnit == "" && h.GlucoseContext == "" {
		return h, nil
	}
	if h.Glucose == nil || h.GlucoseUnit == "" || h.GlucoseContext == "" {
		return h, badRequest("enter the glucose value, its unit and when it was taken")
	}
	unit, ok := glucoseUnits[strings.ToLower(h.GlucoseUnit)]
	if !ok {
		return h, badRequest("the glucose unit must be mmol/L or mg/dL")
	}
	h.GlucoseUnit = unit
	if !glucoseContexts[h.GlucoseContext] {
		return h, badRequest("say whether the glucose was fasting, after a meal or random")
	}
	v := *h.Glucose
	// The column keeps one decimal place; more is refused rather than rounded away.
	if tenths := v * 10; math.Abs(tenths-math.Round(tenths)) > 1e-9 {
		return h, badRequest("the glucose value can have at most one decimal place")
	}
	if r := glucoseRanges[unit]; !r.holds(v) {
		return h, badRequest(fmt.Sprintf(
			"glucose in %s must be between %g and %g; check the value and the unit",
			unit, r.min, r.max,
		))
	}
	return h, nil
}

// healthCheckFingerprint identifies what a normalized request asks for, so a reused retry key can
// be told apart from a retry of the same save.
func healthCheckFingerprint(req types.HealthCheckRequest) string {
	body, _ := json.Marshal(struct {
		CustomerID   int
		Readings     types.HealthReadings
		ConsentGiven bool
	}{req.CustomerID, req.HealthReadings, req.ConsentGiven})
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

// RecordHealthCheck saves a reading for a customer and returns it for the slip. A customer whose
// health consent is not on file needs ConsentGiven, which records it in the same transaction. A
// retry under the same key returns the first save, even if the customer was deactivated or
// withdrew consent since.
func (s *healthCheckService) RecordHealthCheck(
	ctx context.Context, req types.HealthCheckRequest, userID int,
) (types.HealthCheck, error) {
	err := requirePermissions(ctx, "you are not allowed to record health checks",
		constant.RecordHealthCheckPermissionKey, constant.ViewCustomersPermissionKey)
	if err != nil {
		return types.HealthCheck{}, err
	}
	req.IdempotencyKey = strings.TrimSpace(req.IdempotencyKey)
	if req.IdempotencyKey == "" || len(req.IdempotencyKey) > maxHealthRetryKey {
		return types.HealthCheck{}, badRequest("invalid idempotency key")
	}
	if req.CustomerID <= 0 {
		return types.HealthCheck{}, badRequest("choose a customer")
	}
	if req.HealthReadings, err = normalizeHealthReadings(req.HealthReadings); err != nil {
		return types.HealthCheck{}, err
	}
	fingerprint := healthCheckFingerprint(req)

	var saved types.HealthCheck
	err = runInTx(ctx, s.repo, "failed to save the health check", func(tx *sqlx.Tx) error {
		customer, err := s.repo.LockCustomerForHealthCheckTx(ctx, tx, req.CustomerID)
		if errors.Is(err, sql.ErrNoRows) {
			return badRequest("this customer was not found")
		}
		if err != nil {
			return httperror.ServerError("failed to fetch customer", err)
		}
		// After the lock, so a copy that committed while this one waited is found.
		id, found, err := s.earlierHealthCheck(ctx, tx, req.IdempotencyKey, userID, fingerprint)
		if !found {
			id, err = s.recordHealthCheckTx(ctx, tx, req, customer, fingerprint, userID)
		}
		if err != nil {
			return err
		}
		saved, err = s.readHealthCheck(ctx, tx, id)
		return err
	})
	return saved, err
}

func (s *healthCheckService) recordHealthCheckTx(
	ctx context.Context, tx *sqlx.Tx, req types.HealthCheckRequest, customer types.CustomerRef,
	fingerprint string, userID int,
) (int, error) {
	if !customer.Active {
		return 0, conflictWithCode("customer_inactive",
			"this customer is deactivated; reactivate them before recording a health check")
	}
	if !customer.HealthConsent {
		if !req.ConsentGiven {
			return 0, conflictWithCode("health_consent_needed",
				"ask the customer whether the pharmacy may keep their health readings")
		}
		if err := s.repo.SetCustomerHealthConsentTx(ctx, tx, customer.ID, true); err != nil {
			return 0, httperror.ServerError("failed to save the customer's consent", err)
		}
		err := auditHealthConsent(ctx, s.repo, tx, userID, customer.ID, true, "health_check")
		if err != nil {
			return 0, err
		}
	}
	id, created, err := s.repo.CreateHealthCheckTx(ctx, tx, types.NewHealthCheck{
		CustomerID: req.CustomerID, TakenByID: userID, Readings: req.HealthReadings,
		IdempotencyKey: req.IdempotencyKey, RequestHash: fingerprint,
	})
	if err != nil {
		return 0, httperror.ServerError("failed to save the health check", err)
	}
	if !created {
		// Another save took the key after the lookup; answer as that save would be answered.
		id, _, err = s.earlierHealthCheck(ctx, tx, req.IdempotencyKey, userID, fingerprint)
		return id, err
	}
	return id, appendAudit(ctx, s.repo, tx, userID, "health_check", id, "recorded", "",
		map[string]any{"customer_id": req.CustomerID})
}

// earlierHealthCheck answers a request whose key was already used: found is false when nothing
// was saved under it; otherwise it returns that save when it is this exact request from this
// user, and a 409 already_saved for anything else.
func (s *healthCheckService) earlierHealthCheck(
	ctx context.Context, tx *sqlx.Tx, key string, userID int, fingerprint string,
) (int, bool, error) {
	saved, err := s.repo.HealthCheckByKeyTx(ctx, tx, key)
	if err != nil {
		return 0, true, httperror.ServerError("failed to check for a repeated request", err)
	}
	if saved == nil {
		return 0, false, nil
	}
	if saved.TakenByID != userID || saved.RequestHash != fingerprint {
		return 0, true, alreadySaved(
			"this health check was already saved; check the readings before saving again",
		)
	}
	return saved.ID, true, nil
}

func (s *healthCheckService) readHealthCheck(
	ctx context.Context, tx *sqlx.Tx, id int,
) (types.HealthCheck, error) {
	h, err := s.repo.GetHealthCheckTx(ctx, tx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return h, healthCheckNotFound(err)
	}
	if err != nil {
		return h, httperror.ServerError("failed to fetch the health check", err)
	}
	return h, nil
}

// CustomerHealthChecks returns one page of a customer's readings, newest first. Every load is
// audited because the readings are sensitive.
func (s *healthCheckService) CustomerHealthChecks(
	ctx context.Context, customerID, page, userID int,
) (types.HealthCheckPage, error) {
	err := requirePermissions(ctx, "you are not allowed to see health readings",
		constant.ViewHealthChecksPermissionKey)
	if err != nil {
		return types.HealthCheckPage{}, err
	}
	if page < 1 || page > maxHealthCheckPage {
		page = 1
	}
	result := types.HealthCheckPage{Page: page, PerPage: healthChecksPerPage}
	err = runInTx(ctx, s.repo, "failed to fetch health readings", func(tx *sqlx.Tx) error {
		_, err := s.repo.GetCustomerRefTx(ctx, tx, customerID)
		if errors.Is(err, sql.ErrNoRows) {
			return customerNotFound(err)
		}
		if err != nil {
			return httperror.ServerError("failed to fetch customer", err)
		}
		result.Data, result.TotalCount, err = s.repo.ListCustomerHealthChecksTx(
			ctx, tx, customerID, healthChecksPerPage, (page-1)*healthChecksPerPage,
		)
		if err != nil {
			return httperror.ServerError("failed to fetch health readings", err)
		}
		return appendAudit(ctx, s.repo, tx, userID, "customer", customerID,
			"health_history_viewed", "", map[string]any{"page": page})
	})
	if err != nil {
		return types.HealthCheckPage{}, err
	}
	return result, nil
}

// CorrectHealthCheck replaces a saved reading's numbers. The caller sends the version it loaded
// and why; the audit entry keeps the old and new value of every field that changed. A correction
// that changes nothing writes nothing.
func (s *healthCheckService) CorrectHealthCheck(
	ctx context.Context, req types.HealthCheckCorrection, userID int,
) (types.HealthCheck, error) {
	err := requirePermissions(ctx, "you are not allowed to correct health readings",
		constant.CorrectHealthCheckPermissionKey, constant.ViewHealthChecksPermissionKey)
	if err != nil {
		return types.HealthCheck{}, err
	}
	req.Reason = strings.TrimSpace(req.Reason)
	switch {
	case req.Version < 1:
		return types.HealthCheck{}, badRequest("reload the reading before correcting it")
	case req.Reason == "":
		return types.HealthCheck{}, badRequest("say why the reading is being corrected")
	case strings.ContainsRune(req.Reason, 0):
		return types.HealthCheck{}, badRequest("the reason can't contain control characters")
	case utf8.RuneCountInString(req.Reason) > maxHealthReason:
		return types.HealthCheck{}, badRequest(
			fmt.Sprintf("the reason can be at most %d characters", maxHealthReason),
		)
	}
	if req.HealthReadings, err = normalizeHealthReadings(req.HealthReadings); err != nil {
		return types.HealthCheck{}, err
	}

	var result types.HealthCheck
	err = runInTx(ctx, s.repo, "failed to correct the health check", func(tx *sqlx.Tx) error {
		stored, err := s.repo.GetHealthCheckForUpdateTx(ctx, tx, req.ID)
		if errors.Is(err, sql.ErrNoRows) {
			return healthCheckNotFound(err)
		}
		if err != nil {
			return httperror.ServerError("failed to fetch the health check", err)
		}
		if req.Version != stored.Version {
			return conflictWithCode("changed",
				"this reading was changed by someone else; reload and try again")
		}
		changes := changedHealthFields(stored, req.HealthReadings)
		if len(changes) == 0 {
			result = stored
			return nil
		}
		if err = s.repo.UpdateHealthCheckTx(ctx, tx, req.ID, req.HealthReadings); err != nil {
			return httperror.ServerError("failed to correct the health check", err)
		}
		err = appendAudit(ctx, s.repo, tx, userID, "health_check", req.ID, "corrected",
			req.Reason, map[string]any{"changed": changes})
		if err != nil {
			return err
		}
		result, err = s.readHealthCheck(ctx, tx, req.ID)
		return err
	})
	return result, err
}

// healthChange is one corrected field's value before and after; nil is "not measured".
type healthChange struct {
	From any `json:"from"`
	To   any `json:"to"`
}

// changedHealthFields lists the fields a correction changes; a new field is added here once.
func changedHealthFields(
	stored types.HealthCheck, next types.HealthReadings,
) map[string]healthChange {
	changes := map[string]healthChange{}
	add := func(name string, from, to any) {
		if from != to {
			changes[name] = healthChange{from, to}
		}
	}
	add("systolic", intValue(stored.Systolic), intValue(next.Systolic))
	add("diastolic", intValue(stored.Diastolic), intValue(next.Diastolic))
	add("pulse", intValue(stored.Pulse), intValue(next.Pulse))
	add("glucose", tenthsValue(stored.Glucose), tenthsValue(next.Glucose))
	add("glucose_unit", textValue(stored.GlucoseUnit), textValue(&next.GlucoseUnit))
	add("glucose_context", textValue(stored.GlucoseContext), textValue(&next.GlucoseContext))
	add("note", stored.Note, next.Note)
	if c, ok := changes["glucose"]; ok {
		// Compared in tenths so 5.6 read back from the database equals 5.6 sent; shown as is.
		changes["glucose"] = healthChange{fromTenths(c.From), fromTenths(c.To)}
	}
	return changes
}

func intValue(v *int) any {
	if v == nil {
		return nil
	}
	return *v
}

func tenthsValue(v *float64) any {
	if v == nil {
		return nil
	}
	return int64(math.Round(*v * 10))
}

func fromTenths(v any) any {
	if t, ok := v.(int64); ok {
		return float64(t) / 10
	}
	return nil
}

func textValue(v *string) any {
	if v == nil || *v == "" {
		return nil
	}
	return *v
}

// ReprintHealthCheck returns a saved reading for a slip marked as a copy, and records the reprint.
func (s *healthCheckService) ReprintHealthCheck(
	ctx context.Context, id, userID int,
) (types.HealthCheck, error) {
	err := requirePermissions(ctx, "you are not allowed to reprint health checks",
		constant.ViewHealthChecksPermissionKey)
	if err != nil {
		return types.HealthCheck{}, err
	}
	var result types.HealthCheck
	err = runInTx(ctx, s.repo, "failed to reprint the health check", func(tx *sqlx.Tx) error {
		var err error
		if result, err = s.readHealthCheck(ctx, tx, id); err != nil {
			return err
		}
		return appendAudit(ctx, s.repo, tx, userID, "health_check", id, "reprinted", "",
			map[string]any{})
	})
	return result, err
}
