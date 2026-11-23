package types

import "time"

// HealthReadings are the numbers of one health check: a blood pressure (systolic and diastolic,
// with an optional pulse) and/or a blood glucose with its unit and when it was taken. Absent parts
// are nil or empty.
type HealthReadings struct {
	Systolic       *int     `json:"systolic"`
	Diastolic      *int     `json:"diastolic"`
	Pulse          *int     `json:"pulse"`
	Glucose        *float64 `json:"glucose"`
	GlucoseUnit    string   `json:"glucose_unit"`
	GlucoseContext string   `json:"glucose_context"`
	Note           string   `json:"note"`
}

// HealthCheckRequest records a reading for a customer. ConsentGiven says the customer agreed, at
// the counter, to the pharmacy keeping their readings; it is needed only until their consent is on
// file.
type HealthCheckRequest struct {
	CustomerID int `json:"customer_id"`
	HealthReadings
	ConsentGiven   bool   `json:"consent_given"`
	IdempotencyKey string `json:"idempotency_key"`
}

// HealthCheckCorrection corrects a saved reading. Version is the one the caller loaded; Reason
// says why the numbers changed.
type HealthCheckCorrection struct {
	ID      int    `json:"-"`
	Version int    `json:"version"`
	Reason  string `json:"reason"`
	HealthReadings
}

// HealthCheck is a saved reading with the names a slip or history line shows.
type HealthCheck struct {
	ID             int       `json:"id"              db:"id"`
	CustomerID     int       `json:"customer_id"     db:"customer_id"`
	CustomerName   string    `json:"customer_name"   db:"customer_name"`
	TakenAt        time.Time `json:"taken_at"        db:"taken_at"`
	TakenByID      int       `json:"-"               db:"taken_by_id"`
	TakenBy        string    `json:"taken_by"        db:"taken_by"`
	Systolic       *int      `json:"systolic"        db:"systolic"`
	Diastolic      *int      `json:"diastolic"       db:"diastolic"`
	Pulse          *int      `json:"pulse"           db:"pulse"`
	Glucose        *float64  `json:"glucose"         db:"glucose"`
	GlucoseUnit    *string   `json:"glucose_unit"    db:"glucose_unit"`
	GlucoseContext *string   `json:"glucose_context" db:"glucose_context"`
	Note           string    `json:"note"            db:"note"`
	Version        int       `json:"version"         db:"version"`
	// Corrected is true once the reading has been corrected at least once.
	Corrected bool      `json:"corrected"  db:"corrected"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// NewHealthCheck is what the repository stores for a new reading; the readings are validated.
type NewHealthCheck struct {
	CustomerID     int
	TakenByID      int
	Readings       HealthReadings
	IdempotencyKey string
	RequestHash    string
}

// SavedHealthCheck is what a retry needs of the reading saved under its key.
type SavedHealthCheck struct {
	ID          int    `db:"id"`
	TakenByID   int    `db:"taken_by_id"`
	RequestHash string `db:"request_hash"`
}

// HealthCheckPage is one page of a customer's readings, newest first.
type HealthCheckPage struct {
	Data       []HealthCheck `json:"data"`
	Page       int           `json:"page"`
	PerPage    int           `json:"per_page"`
	TotalCount int           `json:"total_count"`
}
