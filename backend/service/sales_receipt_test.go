package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/jmoiron/sqlx"

	"pharmacy/httperror"
	"pharmacy/internal/constant"
	"pharmacy/internal/types"
	"pharmacy/model"
	"pharmacy/repository"
)

type reprintRepo struct {
	repository.PharmacyRepository

	receiptNumber string
	missing       bool
	audit         []types.AuditEntry
	commits       int
	rollbacks     int
}

func (r *reprintRepo) BeginTx(context.Context) (*sqlx.Tx, error) { return nil, nil }
func (r *reprintRepo) CommitTx(*sqlx.Tx) error                   { r.commits++; return nil }
func (r *reprintRepo) RollbackTx(*sqlx.Tx)                       { r.rollbacks++ }

func (r *reprintRepo) FetchSaleReceiptNumberByID(
	context.Context, *sqlx.Tx, int,
) (string, error) {
	if r.missing {
		return "", sql.ErrNoRows
	}
	return r.receiptNumber, nil
}

func (r *reprintRepo) InsertAuditLogTx(_ context.Context, _ *sqlx.Tx, e types.AuditEntry) error {
	r.audit = append(r.audit, e)
	return nil
}

func TestReprintReceiptRecordsWhoReprintedWhat(t *testing.T) {
	repo := &reprintRepo{receiptNumber: "2490-1234"}
	svc := NewSaleService(repo)

	got, err := svc.ReprintReceipt(context.Background(), 42, 7)
	if err != nil {
		t.Fatalf("ReprintReceipt returned error: %v", err)
	}
	if got.ReceiptNumber != "2490-1234" || got.ReprintedAt == "" {
		t.Fatalf("unexpected reply: %#v", got)
	}
	if len(repo.audit) != 1 {
		t.Fatalf("expected one audit entry, got %d", len(repo.audit))
	}
	e := repo.audit[0]
	if e.ActorID != 7 || e.EntityType != constant.AuditEntitySale || e.EntityID != 42 ||
		e.Action != constant.AuditActionReceiptReprinted {
		t.Fatalf("unexpected audit entry: %#v", e)
	}
	var details map[string]any
	if err := json.Unmarshal(e.Details, &details); err != nil {
		t.Fatalf("details are not JSON: %v", err)
	}
	if details["receipt_number"] != "2490-1234" {
		t.Fatalf("details should carry the receipt number, got %#v", details)
	}
	if repo.commits != 1 {
		t.Fatalf("expected one commit, got %d", repo.commits)
	}
}

func TestReprintReceiptUnknownSaleIsNotFoundAndWritesNothing(t *testing.T) {
	repo := &reprintRepo{missing: true}
	svc := NewSaleService(repo)

	_, err := svc.ReprintReceipt(context.Background(), 99, 7)

	var httpErr *httperror.HTTPError
	if !errors.As(err, &httpErr) || httpErr.Code != http.StatusNotFound {
		t.Fatalf("expected a 404, got %v", err)
	}
	if len(repo.audit) != 0 || repo.commits != 0 {
		t.Fatalf("nothing should be written: audit=%d commits=%d", len(repo.audit), repo.commits)
	}
}

func TestManufacturerNameToleratesProductsWithoutOne(t *testing.T) {
	if got := manufacturerName(model.Product{}); got != "" {
		t.Fatalf("expected empty name, got %q", got)
	}
	name := "Emzor"
	if got := manufacturerName(model.Product{Manufacturer: &name}); got != "Emzor" {
		t.Fatalf("expected Emzor, got %q", got)
	}
}
