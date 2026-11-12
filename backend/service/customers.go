package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/mail"
	"pharmacy/httperror"
	"pharmacy/internal/constant"
	"pharmacy/internal/types"
	"pharmacy/model"
	"pharmacy/repository"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/jmoiron/sqlx"
)

const (
	customersPerPage = 25
	// maxCustomerPage keeps the list offset from overflowing.
	maxCustomerPage  = 1_000_000
	maxCustomerName  = 120
	maxCustomerEmail = 254
	maxCustomerNotes = 1000
	maxCustomerQuery = 100
)

type customerService struct {
	repo repository.PharmacyRepository
}

func NewCustomerService(repo repository.PharmacyRepository) *customerService {
	return &customerService{repo: repo}
}

func requireCustomerViewer(ctx context.Context) error {
	if !HasPermission(ctx, constant.ViewCustomersPermissionKey) {
		return httperror.Forbidden("you are not allowed to see customers", nil)
	}
	return nil
}

func requireCustomerManager(ctx context.Context) error {
	if !HasPermission(ctx, constant.ManageCustomersPermissionKey) {
		return httperror.Forbidden("you are not allowed to change customers", nil)
	}
	return nil
}

func customerNotFound(err error) error {
	return httperror.NotFound("customer was not found", err)
}

func customerChangedElsewhere() error {
	return conflict("this customer was changed by someone else; reload and try again")
}

// phoneTaken is the refusal for a number another customer has, naming them so the person typing
// can find the existing record instead of making a second one.
func phoneTaken(owner *types.Customer) error {
	msg := fmt.Sprintf("this phone number is already saved for %s", owner.Name)
	if !owner.Active {
		msg += " (deactivated; reactivate them instead)"
	}
	return phoneTakenError(msg)
}

func phoneTakenRace() error {
	return phoneTakenError("this phone number was just saved for another customer")
}

// phoneTakenError is a 409 whose code lets the page tell it from a stale-edit conflict.
func phoneTakenError(msg string) error {
	err := httperror.Conflict(msg, errors.New(msg))
	err.Reason = "phone_taken"
	return err
}

// tidyCustomerRequest trims the text, normalizes the phone and rejects what can never be valid.
func tidyCustomerRequest(req types.CustomerRequest) (types.CustomerRequest, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)
	req.Notes = strings.TrimSpace(req.Notes)
	switch {
	case req.Name == "":
		return req, badRequest("the customer's name is required")
	case utf8.RuneCountInString(req.Name) > maxCustomerName:
		return req, badRequest(
			fmt.Sprintf("the name can be at most %d characters", maxCustomerName),
		)
	case utf8.RuneCountInString(req.Notes) > maxCustomerNotes:
		return req, badRequest(fmt.Sprintf("notes can be at most %d characters", maxCustomerNotes))
	}
	if req.Email != "" {
		addr, err := mail.ParseAddress(req.Email)
		if err != nil || addr.Address != req.Email || len(req.Email) > maxCustomerEmail {
			return req, badRequest("enter a valid email address, or leave it empty")
		}
	}
	phone, err := model.NormalizePhone(req.Phone)
	if err != nil {
		return req, badRequest(err.Error())
	}
	req.Phone = phone
	return req, nil
}

func (s *customerService) CreateCustomer(
	ctx context.Context, req types.CustomerRequest, userID int,
) (int, error) {
	if err := requireCustomerManager(ctx); err != nil {
		return 0, err
	}
	req, err := tidyCustomerRequest(req)
	if err != nil {
		return 0, err
	}

	var id int
	err = runInTx(ctx, s.repo, "failed to save customer", func(tx *sqlx.Tx) error {
		if err := s.requirePhoneFree(ctx, tx, req.Phone, 0); err != nil {
			return err
		}
		var err error
		id, err = s.repo.CreateCustomerTx(ctx, tx, types.NewCustomer{
			Name: req.Name, Phone: req.Phone, Email: req.Email, Notes: req.Notes,
			ContactConsent: req.ContactConsent, CreatedByID: userID,
		})
		if errors.Is(err, repository.ErrCustomerPhoneTaken) {
			return phoneTakenRace()
		}
		if err != nil {
			return httperror.ServerError("failed to save customer", err)
		}
		return appendAudit(ctx, s.repo, tx, userID, "customer", id, "created", "",
			map[string]any{"contact_consent": req.ContactConsent})
	})
	return id, err
}

// requirePhoneFree refuses a phone number that a customer other than exceptID already has.
func (s *customerService) requirePhoneFree(
	ctx context.Context, tx *sqlx.Tx, phone string, exceptID int,
) error {
	if phone == "" {
		return nil
	}
	owner, err := s.repo.CustomerByPhoneTx(ctx, tx, phone)
	if err != nil {
		return httperror.ServerError("failed to check the phone number", err)
	}
	if owner != nil && owner.ID != exceptID {
		return phoneTaken(owner)
	}
	return nil
}

func (s *customerService) UpdateCustomer(
	ctx context.Context, req types.CustomerRequest, userID int,
) error {
	if err := requireCustomerManager(ctx); err != nil {
		return err
	}
	req, err := tidyCustomerRequest(req)
	if err != nil {
		return err
	}
	return runInTx(ctx, s.repo, "failed to save customer", func(tx *sqlx.Tx) error {
		stored, err := s.lockCustomer(ctx, tx, req.ID, req.Version)
		if err != nil {
			return err
		}

		changed := changedCustomerFields(stored, req)
		if len(changed) == 0 {
			return nil
		}
		details := map[string]any{"changed": changed}
		if stored.ContactConsent != req.ContactConsent {
			details["contact_consent"] = map[string]bool{
				"from": stored.ContactConsent, "to": req.ContactConsent,
			}
		}
		edited := stored
		edited.Name, edited.Phone, edited.Email = req.Name, req.Phone, req.Email
		edited.Notes, edited.ContactConsent = req.Notes, req.ContactConsent
		if stored.Phone != edited.Phone {
			if err = s.requirePhoneFree(ctx, tx, edited.Phone, stored.ID); err != nil {
				return err
			}
		}
		err = s.repo.UpdateCustomerTx(ctx, tx, edited)
		if errors.Is(err, repository.ErrCustomerPhoneTaken) {
			return phoneTakenRace()
		}
		if err != nil {
			return httperror.ServerError("failed to save customer", err)
		}
		return appendAudit(ctx, s.repo, tx, userID, "customer", stored.ID, "updated", "", details)
	})
}

// changedCustomerFields names the fields an edit changes; a new field is added here once.
func changedCustomerFields(stored types.Customer, req types.CustomerRequest) []string {
	var changed []string
	for _, f := range []struct {
		name      string
		different bool
	}{
		{"name", stored.Name != req.Name},
		{"phone", stored.Phone != req.Phone},
		{"email", stored.Email != req.Email},
		{"notes", stored.Notes != req.Notes},
		{"contact_consent", stored.ContactConsent != req.ContactConsent},
	} {
		if f.different {
			changed = append(changed, f.name)
		}
	}
	return changed
}

// lockCustomer reads the customer with its row locked and refuses a copy that is out of date.
func (s *customerService) lockCustomer(
	ctx context.Context, tx *sqlx.Tx, id, version int,
) (types.Customer, error) {
	stored, err := s.repo.GetCustomerForUpdateTx(ctx, tx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return stored, customerNotFound(err)
	}
	if err != nil {
		return stored, httperror.ServerError("failed to fetch customer", err)
	}
	if version != stored.Version {
		return stored, customerChangedElsewhere()
	}
	return stored, nil
}

func (s *customerService) SetCustomerActive(
	ctx context.Context, id, version int, active bool, userID int,
) error {
	if err := requireCustomerManager(ctx); err != nil {
		return err
	}
	return runInTx(ctx, s.repo, "failed to save customer", func(tx *sqlx.Tx) error {
		stored, err := s.lockCustomer(ctx, tx, id, version)
		if err != nil {
			return err
		}
		if stored.Active == active {
			return nil
		}
		if err = s.repo.SetCustomerActiveTx(ctx, tx, id, active); err != nil {
			return httperror.ServerError("failed to save customer", err)
		}
		action := "deactivated"
		if active {
			action = "reactivated"
		}
		return appendAudit(ctx, s.repo, tx, userID, "customer", id, action, "", map[string]any{})
	})
}

func (s *customerService) GetCustomer(ctx context.Context, id int) (types.Customer, error) {
	if err := requireCustomerViewer(ctx); err != nil {
		return types.Customer{}, err
	}
	c, err := s.repo.GetCustomer(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return c, customerNotFound(err)
	}
	if err != nil {
		return c, httperror.ServerError("failed to fetch customer", err)
	}
	return c, nil
}

func (s *customerService) ListCustomers(
	ctx context.Context, search string, includeInactive bool, page int,
) (types.CustomerPage, error) {
	if err := requireCustomerViewer(ctx); err != nil {
		return types.CustomerPage{}, err
	}
	if page < 1 || page > maxCustomerPage {
		page = 1
	}
	filter := customerFilter(search)
	filter.IncludeInactive = includeInactive

	customers, total, err := s.repo.ListCustomers(
		ctx, filter, customersPerPage, (page-1)*customersPerPage,
	)
	if err != nil {
		return types.CustomerPage{}, httperror.ServerError("failed to fetch customers", err)
	}
	return types.CustomerPage{
		Data: customers, Page: page, PerPage: customersPerPage, TotalCount: total,
	}, nil
}

// customerFilter reads what was typed in the search box. Text with a letter in it looks for a
// name. Otherwise it is part of a phone number. A number written from its start (0803, +234 803,
// 234803) is matched from the start of the stored number, so "0803" does not find numbers that
// merely contain 803; digits that don't begin a number match anywhere inside it.
func customerFilter(search string) types.CustomerFilter {
	search = strings.TrimSpace(search)
	if r := []rune(search); len(r) > maxCustomerQuery {
		search = string(r[:maxCustomerQuery])
	}
	if strings.IndexFunc(search, unicode.IsLetter) >= 0 {
		return types.CustomerFilter{Name: search}
	}
	var digits strings.Builder
	for _, r := range search {
		if r >= '0' && r <= '9' {
			digits.WriteRune(r)
		}
	}
	number := digits.String()
	if number == "" {
		return types.CustomerFilter{}
	}
	if strings.HasPrefix(search, "+") {
		if national, ok := strings.CutPrefix(number, "234"); ok {
			return types.CustomerFilter{
				PhonePrefix: "+234", PhoneDigits: strings.TrimPrefix(national, "0"),
			}
		}
		return types.CustomerFilter{PhonePrefix: "+", PhoneDigits: number}
	}
	if national, ok := strings.CutPrefix(number, "0"); ok {
		return types.CustomerFilter{PhonePrefix: "+234", PhoneDigits: national}
	}
	if national, ok := strings.CutPrefix(number, "234"); ok {
		return types.CustomerFilter{PhonePrefix: "+234", PhoneDigits: national}
	}
	return types.CustomerFilter{PhoneDigits: number}
}
