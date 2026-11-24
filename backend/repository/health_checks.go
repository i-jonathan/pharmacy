package repository

import (
	"context"
	"database/sql"
	"errors"
	"pharmacy/internal/types"

	"github.com/jmoiron/sqlx"
)

// healthCheckSelect reads readings with the customer's name and who took them. An empty unit or
// context is stored as NULL.
const healthCheckSelect = `
	SELECT h.id, h.customer_id, c.name AS customer_name, h.taken_at, h.taken_by_id,
		u.username AS taken_by, h.systolic, h.diastolic, h.pulse, h.glucose::float8 AS glucose,
		h.glucose_unit, h.glucose_context, h.note, h.version, h.version > 1 AS corrected,
		h.updated_at
	FROM health_check h
	JOIN customer c ON c.id = h.customer_id
	JOIN users u ON u.id = h.taken_by_id`

const createHealthCheckQuery = `
	INSERT INTO health_check (
		customer_id, taken_by_id, systolic, diastolic, pulse, glucose, glucose_unit,
		glucose_context, note, idempotency_key, request_hash
	)
	VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''), NULLIF($8, ''), $9, NULLIF($10, ''), $11)
	ON CONFLICT (idempotency_key) WHERE idempotency_key IS NOT NULL DO NOTHING
	RETURNING id`

func (r *repo) CreateHealthCheckTx(
	ctx context.Context, tx *sqlx.Tx, check types.NewHealthCheck,
) (int, bool, error) {
	h := check.Readings
	var id int
	err := tx.GetContext(ctx, &id, createHealthCheckQuery,
		check.CustomerID, check.TakenByID, h.Systolic, h.Diastolic, h.Pulse, h.Glucose,
		h.GlucoseUnit, h.GlucoseContext, h.Note, check.IdempotencyKey, check.RequestHash)
	if errors.Is(err, sql.ErrNoRows) && check.IdempotencyKey != "" {
		return 0, false, nil
	}
	return id, err == nil, err
}

func (r *repo) HealthCheckByKeyTx(
	ctx context.Context, tx *sqlx.Tx, key string,
) (*types.SavedHealthCheck, error) {
	var saved types.SavedHealthCheck
	err := tx.GetContext(ctx, &saved, `
		SELECT id, taken_by_id, COALESCE(request_hash, '') AS request_hash
		FROM health_check WHERE idempotency_key = $1`, key)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &saved, nil
}

func (r *repo) GetHealthCheckTx(
	ctx context.Context, tx *sqlx.Tx, id int,
) (types.HealthCheck, error) {
	var h types.HealthCheck
	err := tx.GetContext(ctx, &h, healthCheckSelect+` WHERE h.id = $1`, id)
	return h, err
}

func (r *repo) GetHealthCheckForUpdateTx(
	ctx context.Context, tx *sqlx.Tx, id int,
) (types.HealthCheck, error) {
	var h types.HealthCheck
	err := tx.GetContext(ctx, &h, healthCheckSelect+` WHERE h.id = $1 FOR UPDATE OF h`, id)
	return h, err
}

func (r *repo) UpdateHealthCheckTx(
	ctx context.Context, tx *sqlx.Tx, id int, h types.HealthReadings,
) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE health_check
		SET systolic = $2, diastolic = $3, pulse = $4, glucose = $5,
			glucose_unit = NULLIF($6, ''), glucose_context = NULLIF($7, ''), note = $8,
			version = version + 1, updated_at = NOW()
		WHERE id = $1`,
		id, h.Systolic, h.Diastolic, h.Pulse, h.Glucose, h.GlucoseUnit, h.GlucoseContext, h.Note)
	return err
}

func (r *repo) ListCustomerHealthChecksTx(
	ctx context.Context, tx *sqlx.Tx, customerID, limit, offset int,
) ([]types.HealthCheck, int, error) {
	var total int
	err := tx.GetContext(ctx, &total,
		`SELECT count(*) FROM health_check WHERE customer_id = $1`, customerID)
	if err != nil {
		return nil, 0, err
	}
	checks := []types.HealthCheck{}
	err = tx.SelectContext(ctx, &checks, healthCheckSelect+`
		WHERE h.customer_id = $1
		ORDER BY h.taken_at DESC, h.id DESC
		LIMIT $2 OFFSET $3`, customerID, limit, offset)
	return checks, total, err
}
