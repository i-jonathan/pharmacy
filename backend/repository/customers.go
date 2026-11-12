package repository

import (
	"context"
	"database/sql"
	"errors"
	"pharmacy/internal/types"
	"strings"

	"github.com/jmoiron/sqlx"
)

// ErrCustomerPhoneTaken is returned when a customer is saved with a phone number another customer
// already has.
var ErrCustomerPhoneTaken = errors.New("phone number already belongs to another customer")

// likeEscaper makes text safe to put inside a LIKE pattern that uses \ as its escape character.
var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// customerColumns reads a customer, with no phone shown as an empty string.
const customerColumns = `
	id, name, COALESCE(phone, '') AS phone, email, notes, contact_consent, consent_updated_at,
	active, version, created_at, updated_at`

func (r *repo) CreateCustomerTx(
	ctx context.Context, tx *sqlx.Tx, c types.NewCustomer,
) (int, error) {
	var id int
	err := tx.GetContext(ctx, &id, `
		INSERT INTO customer
			(name, phone, email, notes, contact_consent, consent_updated_at, created_by_id)
		VALUES ($1, NULLIF($2, ''), $3, $4, $5, CASE WHEN $5 THEN NOW() END, $6)
		RETURNING id`,
		c.Name, c.Phone, c.Email, c.Notes, c.ContactConsent, c.CreatedByID)
	if isUniqueViolation(err) {
		return 0, ErrCustomerPhoneTaken
	}
	return id, err
}

func (r *repo) GetCustomerForUpdateTx(
	ctx context.Context, tx *sqlx.Tx, id int,
) (types.Customer, error) {
	var c types.Customer
	err := tx.GetContext(ctx, &c,
		`SELECT `+customerColumns+` FROM customer WHERE id = $1 FOR UPDATE`, id)
	return c, err
}

func (r *repo) UpdateCustomerTx(ctx context.Context, tx *sqlx.Tx, c types.Customer) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE customer
		SET name = $2, phone = NULLIF($3, ''), email = $4, notes = $5,
			consent_updated_at = CASE
				WHEN contact_consent IS DISTINCT FROM $6 THEN NOW() ELSE consent_updated_at END,
			contact_consent = $6, version = version + 1, updated_at = NOW()
		WHERE id = $1`,
		c.ID, c.Name, c.Phone, c.Email, c.Notes, c.ContactConsent)
	if isUniqueViolation(err) {
		return ErrCustomerPhoneTaken
	}
	return err
}

func (r *repo) SetCustomerActiveTx(ctx context.Context, tx *sqlx.Tx, id int, active bool) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE customer SET active = $2, version = version + 1, updated_at = NOW()
		WHERE id = $1`, id, active)
	return err
}

func (r *repo) CustomerByPhoneTx(
	ctx context.Context, tx *sqlx.Tx, phone string,
) (*types.Customer, error) {
	var c types.Customer
	err := tx.GetContext(ctx, &c,
		`SELECT `+customerColumns+` FROM customer WHERE phone = $1`, phone)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *repo) GetCustomer(ctx context.Context, id int) (types.Customer, error) {
	var c types.Customer
	err := r.Data.GetContext(ctx, &c, `SELECT `+customerColumns+` FROM customer WHERE id = $1`, id)
	return c, err
}

// ListCustomers returns one page ordered by name, and the number of customers the filter matches
// in all.
func (r *repo) ListCustomers(
	ctx context.Context, filter types.CustomerFilter, limit, offset int,
) ([]types.Customer, int, error) {
	// LIKE wildcards typed by the user are escaped so "50%" means the text "50%".
	const where = `
		WHERE ($1 OR active)
		  AND ($2 = '' OR lower(name) LIKE '%' || lower($2) || '%' ESCAPE '\')
		  AND ($3 = '' AND $4 = '' OR phone LIKE
			CASE WHEN $4 = '' THEN '%' || $3 || '%' ELSE $4 || $3 || '%' END)`
	name := likeEscaper.Replace(filter.Name)

	var total int
	err := r.Data.GetContext(ctx, &total, `SELECT count(*) FROM customer`+where,
		filter.IncludeInactive, name, filter.PhoneDigits, filter.PhonePrefix)
	if err != nil {
		return nil, 0, err
	}
	customers := []types.Customer{}
	err = r.Data.SelectContext(ctx, &customers, `
		SELECT `+customerColumns+` FROM customer`+where+`
		ORDER BY lower(name), id LIMIT $5 OFFSET $6`,
		filter.IncludeInactive, name, filter.PhoneDigits, filter.PhonePrefix, limit, offset)
	return customers, total, err
}
