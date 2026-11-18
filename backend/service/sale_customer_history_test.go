package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"pharmacy/internal/constant"
	"pharmacy/internal/types"
	"pharmacy/model"

	"github.com/jmoiron/sqlx"
)

// changesRepo is a sale history with two sales and a stored list of customer changes.
type changesRepo struct {
	*customerSaleRepo
	rows         []types.SaleCustomerChangeRow
	changesError error
	changesAsked [][]int
}

func (r *changesRepo) FetchSaleCustomerChanges(
	_ context.Context, _ *sqlx.Tx, ids []int,
) ([]types.SaleCustomerChangeRow, error) {
	r.changesAsked = append(r.changesAsked, ids)
	return r.rows, r.changesError
}

func newChangesRepo(rows ...types.SaleCustomerChangeRow) *changesRepo {
	base := &customerSaleRepo{
		saleRows: []model.Sale{
			{ReceiptNumber: "A-1"},
			{ReceiptNumber: "A-2"},
		},
		namesByID: map[int]string{5: "Ada Obi", 6: "Bola Ade"},
	}
	base.saleRows[0].ID = 1
	base.saleRows[1].ID = 2
	return &changesRepo{customerSaleRepo: base, rows: rows}
}

func changeRow(sale int, actor, reason string, from, to int) types.SaleCustomerChangeRow {
	details, _ := json.Marshal(map[string]any{
		"receipt_number": "A", "from_customer_id": from, "to_customer_id": to,
	})
	return types.SaleCustomerChangeRow{
		SaleID: sale, OccurredAt: time.Date(2026, 11, 18, 9, 30, 0, 0, time.UTC),
		Actor: actor, Reason: reason, Details: details,
	}
}

func historyFor(t *testing.T, repo *changesRepo, keys ...string) types.SaleHistory {
	t.Helper()
	history, err := NewSaleService(repo).FetchSalesHistory(
		contextWith(keys...), types.SaleFilter{Page: 1, PerPage: 10},
	)
	if err != nil {
		t.Fatal(err)
	}
	return history
}

func TestSalesHistoryListsWhoChangedTheCustomerAndWhen(t *testing.T) {
	repo := newChangesRepo(
		changeRow(1, "jide", "other: she gave a new number", 6, 0),
		changeRow(1, "tola", "picked_by_mistake", 5, 6),
	)
	history := historyFor(t, repo, constant.ViewCustomersPermissionKey)

	got := history.Data[0].CustomerChanges
	if len(got) != 2 {
		t.Fatalf("sale 1 should list 2 changes, got %+v", got)
	}
	first := got[0]
	if first.By != "jide" || first.At != "2026-11-18T09:30:00Z" ||
		first.From == nil || first.From.ID != 6 || first.From.Name != "Bola Ade" ||
		first.To != nil || first.Reason != "other" || first.Note != "she gave a new number" {
		t.Fatalf("a removal to walk-in with a note: %+v", first)
	}
	second := got[1]
	if second.From == nil || second.From.Name != "Ada Obi" || second.To == nil ||
		second.To.Name != "Bola Ade" || second.Reason != "picked_by_mistake" || second.Note != "" {
		t.Fatalf("a change with no note: %+v", second)
	}
	if history.Data[1].CustomerChanges == nil || len(history.Data[1].CustomerChanges) != 0 {
		t.Fatalf("a sale never changed has an empty list, not null: %+v",
			history.Data[1].CustomerChanges)
	}
}

func TestSalesHistoryAsksForChangesAndNamesInOneLookupEach(t *testing.T) {
	repo := newChangesRepo(changeRow(1, "jide", "customer_asked", 5, 6))
	historyFor(t, repo, constant.ViewCustomersPermissionKey)

	if len(repo.changesAsked) != 1 || len(repo.changesAsked[0]) != 2 {
		t.Fatalf("one query for the 2 sales on the page, got %v", repo.changesAsked)
	}
	if repo.namesFetched != 1 || len(repo.namesAsked[0]) != 2 {
		t.Fatalf("one name lookup for customers 5 and 6, got %v", repo.namesAsked)
	}
}

func TestSalesHistoryHidesCustomerChangesFromUsersWhoCannotSeeCustomers(t *testing.T) {
	repo := newChangesRepo(changeRow(1, "jide", "customer_asked", 5, 6))
	history := historyFor(t, repo)

	if len(repo.changesAsked) != 0 || repo.namesFetched != 0 {
		t.Fatalf("nothing is read for them: changes %v, lookups %d",
			repo.changesAsked, repo.namesFetched)
	}
	for _, sale := range history.Data {
		if sale.CustomerChanges == nil || len(sale.CustomerChanges) != 0 {
			t.Fatalf("the list must be present and empty: %+v", sale.CustomerChanges)
		}
	}
}

func TestStoredReasonIsSplitIntoCodeAndNote(t *testing.T) {
	for name, tc := range map[string]struct{ stored, reason, note string }{
		"code only":              {"customer_asked", "customer_asked", ""},
		"code and note":          {"other: wrong phone: typo", "other", "wrong phone: typo"},
		"unknown start":          {"typo: x", "", "typo: x"},
		"free text, no code":     {"fixing it", "", "fixing it"},
		"code with empty note":   {"other: ", "other", ""},
		"code name not at start": {"note picked_by_mistake", "", "note picked_by_mistake"},
	} {
		repo := newChangesRepo(changeRow(1, "jide", tc.stored, 5, 6))
		got := historyFor(t, repo, constant.ViewCustomersPermissionKey).Data[0].CustomerChanges
		if len(got) != 1 || got[0].Reason != tc.reason || got[0].Note != tc.note {
			t.Errorf("%s: %q -> %+v, want %q / %q", name, tc.stored, got, tc.reason, tc.note)
		}
	}
}

func TestAnEntryWithUnreadableDetailsIsSkippedNotFatal(t *testing.T) {
	bad := changeRow(1, "jide", "customer_asked", 5, 6)
	bad.Details = json.RawMessage(`not json`)
	repo := newChangesRepo(bad, changeRow(1, "tola", "customer_asked", 5, 6))

	got := historyFor(t, repo, constant.ViewCustomersPermissionKey).Data[0].CustomerChanges
	if len(got) != 1 || got[0].By != "tola" {
		t.Fatalf("only the readable entry is listed: %+v", got)
	}
}

func TestSalesHistoryRollsBackWhenTheChangesCannotBeRead(t *testing.T) {
	repo := newChangesRepo()
	repo.changesError = errors.New("connection reset")
	_, err := NewSaleService(repo).FetchSalesHistory(
		contextWith(constant.ViewCustomersPermissionKey), types.SaleFilter{Page: 1, PerPage: 10},
	)
	requireHTTPStatus(t, err, http.StatusInternalServerError)
	if repo.rollbacks != 1 || repo.commits != 0 {
		t.Fatalf("rollbacks %d, commits %d", repo.rollbacks, repo.commits)
	}
}
