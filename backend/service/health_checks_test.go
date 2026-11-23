package service

import (
	"context"
	"net/http"
	"pharmacy/internal/constant"
	"pharmacy/internal/types"
	"strings"
	"testing"
)

func intp(v int) *int           { return &v }
func floatp(v float64) *float64 { return &v }
func healthCtx(keys ...string) context.Context {
	perms := map[string]bool{}
	for _, k := range keys {
		perms[k] = true
	}
	return context.WithValue(context.Background(), constant.PermissionsSessionKey, perms)
}

func bp(sys, dia int) types.HealthReadings {
	return types.HealthReadings{Systolic: intp(sys), Diastolic: intp(dia)}
}

func glucose(v float64, unit, when string) types.HealthReadings {
	return types.HealthReadings{Glucose: floatp(v), GlucoseUnit: unit, GlucoseContext: when}
}

func TestNormalizeHealthReadingsAcceptsAndTidies(t *testing.T) {
	in := types.HealthReadings{
		Systolic: intp(120), Diastolic: intp(80), Pulse: intp(72),
		Glucose: floatp(5.6), GlucoseUnit: " MMOL/L ", GlucoseContext: " After_Meal ",
		Note: "  rested 5 minutes  ",
	}
	got, err := normalizeHealthReadings(in)
	if err != nil {
		t.Fatal(err)
	}
	if got.GlucoseUnit != "mmol/L" || got.GlucoseContext != "after_meal" ||
		got.Note != "rested 5 minutes" {
		t.Fatalf("not tidied: %+v", got)
	}
	for _, ok := range []types.HealthReadings{
		bp(300, 200), bp(50, 30),
		glucose(1.0, "mmol/L", "fasting"), glucose(50, "mmol/L", "random"),
		glucose(20, "mg/dL", "fasting"), glucose(900, "mg/dL", "fasting"),
		glucose(99.5, "mg/dl", "random"),
	} {
		if _, err := normalizeHealthReadings(ok); err != nil {
			t.Errorf("%+v refused: %v", ok, err)
		}
	}
}

func TestNormalizeHealthReadingsRefusesWhatCanNeverBeRight(t *testing.T) {
	withPulse := bp(120, 80)
	withPulse.Pulse = intp(251)
	pulseOnly := types.HealthReadings{Pulse: intp(72)}
	longNote := bp(120, 80)
	longNote.Note = strings.Repeat("é", 501)
	nul := bp(120, 80)
	nul.Note = "a\x00b"
	cases := map[string]types.HealthReadings{
		"nothing":            {},
		"systolic alone":     {Systolic: intp(120)},
		"diastolic alone":    {Diastolic: intp(80)},
		"pulse alone":        pulseOnly,
		"pulse out of range": withPulse,
		"top not above":      bp(80, 80),
		"systolic typo":      bp(1200, 80),
		"systolic low":       bp(49, 30),
		"diastolic high":     bp(290, 201),
		"diastolic low":      bp(100, 29),
		"glucose no unit":    glucose(5.6, "", "fasting"),
		"glucose no context": glucose(5.6, "mmol/L", ""),
		"unit no glucose":    {GlucoseUnit: "mmol/L", GlucoseContext: "fasting"},
		"unknown unit":       glucose(5.6, "mg", "fasting"),
		"unknown context":    glucose(5.6, "mmol/L", "lunch"),
		"two decimals":       glucose(5.65, "mmol/L", "fasting"),
		"mmol high":          glucose(50.1, "mmol/L", "fasting"),
		"mmol low":           glucose(0.9, "mmol/L", "fasting"),
		"mg high":            glucose(900.1, "mg/dL", "fasting"),
		"mg low":             glucose(19.9, "mg/dL", "fasting"),
		"negative glucose":   glucose(-5, "mmol/L", "fasting"),
		"note too long":      longNote,
		"control char note":  nul,
	}
	for name, in := range cases {
		_, err := normalizeHealthReadings(in)
		if err == nil {
			t.Errorf("%s: expected a refusal", name)
			continue
		}
		requireHTTPStatus(t, err, http.StatusBadRequest)
		// A refusal never repeats a health value.
		if strings.Contains(err.Error(), "1200") || strings.Contains(err.Error(), "5.65") {
			t.Errorf("%s: message echoes the value: %v", name, err)
		}
	}
}

func TestHealthCheckMethodsNeedTheirPermissions(t *testing.T) {
	svc := NewHealthCheckService(nil) // a nil repository: none of these may reach the database
	record := types.HealthCheckRequest{
		CustomerID: 1, HealthReadings: bp(120, 80), IdempotencyKey: "k",
	}
	correction := types.HealthCheckCorrection{
		ID: 1, Version: 1, Reason: "typo", HealthReadings: bp(120, 80),
	}
	cashier := healthCtx(
		constant.RecordHealthCheckPermissionKey,
		constant.ViewCustomersPermissionKey,
	)

	_, err := svc.RecordHealthCheck(healthCtx(constant.ViewCustomersPermissionKey), record, 1)
	requireHTTPStatus(t, err, http.StatusForbidden)
	_, err = svc.RecordHealthCheck(healthCtx(constant.RecordHealthCheckPermissionKey), record, 1)
	requireHTTPStatus(t, err, http.StatusForbidden)

	_, err = svc.CustomerHealthChecks(cashier, 1, 1, 1)
	requireHTTPStatus(t, err, http.StatusForbidden)
	_, err = svc.ReprintHealthCheck(cashier, 1, 1)
	requireHTTPStatus(t, err, http.StatusForbidden)
	_, err = svc.CorrectHealthCheck(cashier, correction, 1)
	requireHTTPStatus(t, err, http.StatusForbidden)
	// Correcting also needs to be allowed to see readings.
	_, err = svc.CorrectHealthCheck(
		healthCtx(constant.CorrectHealthCheckPermissionKey),
		correction,
		1,
	)
	requireHTTPStatus(t, err, http.StatusForbidden)
}

func TestRecordHealthCheckNeedsARetryKeyAndACustomer(t *testing.T) {
	svc := NewHealthCheckService(nil) // a nil repository: none of these may reach the database
	ctx := healthCtx(constant.RecordHealthCheckPermissionKey, constant.ViewCustomersPermissionKey)
	_, err := svc.RecordHealthCheck(ctx, types.HealthCheckRequest{
		CustomerID: 1, HealthReadings: bp(120, 80),
	}, 1)
	requireHTTPStatus(t, err, http.StatusBadRequest)
	_, err = svc.RecordHealthCheck(ctx, types.HealthCheckRequest{
		HealthReadings: bp(120, 80), IdempotencyKey: "k",
	}, 1)
	requireHTTPStatus(t, err, http.StatusBadRequest)
}

func TestCorrectHealthCheckNeedsAReason(t *testing.T) {
	svc := NewHealthCheckService(nil) // a nil repository: none of these may reach the database
	ctx := healthCtx(
		constant.CorrectHealthCheckPermissionKey,
		constant.ViewHealthChecksPermissionKey,
	)
	for _, reason := range []string{"", "   ", strings.Repeat("x", 201), "a\x00"} {
		_, err := svc.CorrectHealthCheck(ctx, types.HealthCheckCorrection{
			ID: 1, Version: 1, Reason: reason, HealthReadings: bp(120, 80),
		}, 1)
		requireHTTPStatus(t, err, http.StatusBadRequest)
	}
}
