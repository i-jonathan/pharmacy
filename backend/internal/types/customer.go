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
	Active           bool       `json:"active"             db:"active"`
	Version          int        `json:"version"            db:"version"`
	CreatedAt        time.Time  `json:"created_at"         db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"         db:"updated_at"`
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
}
