package types

import "time"

// Customer is a person the pharmacy keeps on file. Phone is stored normalized (+234...) and is
// empty when the customer gave none.
type Customer struct {
	ID               int        `json:"id"                 db:"id"`
	Name             string     `json:"name"               db:"name"`
	Phone            string     `json:"phone"              db:"phone"`
	Email            string     `json:"email"              db:"email"`
	Notes            string     `json:"notes"              db:"notes"`
	ContactConsent   bool       `json:"contact_consent"    db:"contact_consent"`
	ConsentUpdatedAt *time.Time `json:"consent_updated_at" db:"consent_updated_at"`
	// HealthConsent: the customer agreed to the pharmacy keeping their health readings.
	HealthConsent          bool       `json:"health_consent"            db:"health_consent"`
	HealthConsentUpdatedAt *time.Time `json:"health_consent_updated_at" db:"health_consent_updated_at"`
	Active           bool       `json:"active"             db:"active"`
	Version          int        `json:"version"            db:"version"`
	CreatedAt        time.Time  `json:"created_at"         db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"         db:"updated_at"`
}

// CustomerRef is the part of a customer a sale needs: who they are and whether they may be put
// on a sale. It is read instead of the whole row, which holds phone, email and notes.
type CustomerRef struct {
	ID     int    `db:"id"`
	Name   string `db:"name"`
	Active bool   `db:"active"`
	// HealthConsent is read only by LockCustomerForHealthCheckTx.
	HealthConsent bool `db:"health_consent"`
}

// CustomerRequest adds a customer (ID 0) or edits one; an edit sends the version it loaded.
type CustomerRequest struct {
	ID             int    `json:"id"`
	Version        int    `json:"version"`
	Name           string `json:"name"`
	Phone          string `json:"phone"`
	Email          string `json:"email"`
	Notes          string `json:"notes"`
	ContactConsent bool   `json:"contact_consent"`
	// HealthConsent is left as it is when absent, so a form that does not show it can't clear it.
	HealthConsent *bool `json:"health_consent"`
}

// CustomerActionRequest deactivates or reactivates a customer.
type CustomerActionRequest struct {
	Version int `json:"version"`
}

// NewCustomer is what the repository stores for a new customer; Phone is already normalized.
type NewCustomer struct {
	Name           string
	Phone          string
	Email          string
	Notes          string
	ContactConsent bool
	CreatedByID    int
}

// CustomerPage is one page of the customer list.
type CustomerPage struct {
	Data       []Customer `json:"data"`
	Page       int        `json:"page"`
	PerPage    int        `json:"per_page"`
	TotalCount int        `json:"total_count"`
}

// CustomerFilter narrows the customer list. Name matches part of a name. PhoneDigits matches
// digits of the stored phone number: straight after PhonePrefix ("+234" or "+") when that is set,
// anywhere in the number when it is not. At most one of Name and the phone fields is set.
type CustomerFilter struct {
	Name            string
	PhonePrefix     string
	PhoneDigits     string
	IncludeInactive bool
	// ActiveFirst lists active customers before deactivated ones (each group by name); otherwise
	// the list is one alphabetical run. The till asks for it so a page of matches never fills
	// with deactivated customers ahead of an active one.
	ActiveFirst bool
}
