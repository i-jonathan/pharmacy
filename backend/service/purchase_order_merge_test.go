package service

import (
	"context"
	"net/http"
	"pharmacy/internal/constant"
	"pharmacy/internal/types"
	"strings"
	"testing"
)

func mergeLine(
	id int, product *int, name, manufacturer, note string, quantity int,
) types.PurchaseOrderItem {
	return types.PurchaseOrderItem{
		ID: id, ProductID: product, Name: name, Manufacturer: manufacturer, Note: note,
		Quantity: quantity,
	}
}

func TestPlanPurchaseOrderMerge(t *testing.T) {
	alpha := intPtr(1)
	order := func(note string, items ...types.PurchaseOrderItem) types.PurchaseOrder {
		return types.PurchaseOrder{Note: note, Items: items}
	}

	t.Run("same product keeps the higher quantity", func(t *testing.T) {
		plan, err := planPurchaseOrderMerge(
			order("", mergeLine(5, alpha, "Alpha", "", "", 10)),
			order("", mergeLine(9, alpha, "Alpha", "", "", 15)),
		)
		if err != nil {
			t.Fatal(err)
		}
		if len(plan.updates) != 1 || len(plan.inserts) != 0 {
			t.Fatalf("expected one update and no inserts, got %+v", plan)
		}
		got := plan.updates[0]
		if got.item.ID != 5 || got.item.Quantity != 15 || got.quantityBefore != 10 {
			t.Fatalf("expected line 5 raised from 10 to 15, got %+v", got)
		}
	})

	t.Run("a lower or equal draft quantity changes nothing", func(t *testing.T) {
		for _, draftQuantity := range []int{4, 10} {
			plan, err := planPurchaseOrderMerge(
				order("", mergeLine(5, alpha, "Alpha", "", "", 10)),
				order("", mergeLine(9, alpha, "Alpha", "", "", draftQuantity)),
			)
			if err != nil {
				t.Fatal(err)
			}
			if len(plan.updates) != 0 || len(plan.inserts) != 0 {
				t.Fatalf("draft quantity %d: expected no changes, got %+v", draftQuantity, plan)
			}
		}
	})

	t.Run(
		"lines without a product match on name and manufacturer ignoring case",
		func(t *testing.T) {
			plan, err := planPurchaseOrderMerge(
				order("", mergeLine(5, nil, "Amoxil", "GSK", "", 5)),
				order("", mergeLine(9, nil, "AMOXIL", "gsk", "", 8)),
			)
			if err != nil {
				t.Fatal(err)
			}
			if len(plan.updates) != 1 || plan.updates[0].item.Quantity != 8 ||
				len(plan.inserts) != 0 {
				t.Fatalf("expected the typed line raised to 8, got %+v", plan)
			}

			plan, err = planPurchaseOrderMerge(
				order("", mergeLine(5, nil, "Amoxil", "GSK", "", 5)),
				order("", mergeLine(9, nil, "Amoxil", "Emzor", "", 8)),
			)
			if err != nil {
				t.Fatal(err)
			}
			if len(plan.updates) != 0 || len(plan.inserts) != 1 {
				t.Fatalf("a different manufacturer is a different line, got %+v", plan)
			}
		},
	)

	t.Run("a typed draft line does not match a product line", func(t *testing.T) {
		plan, err := planPurchaseOrderMerge(
			order("", mergeLine(5, alpha, "Alpha", "", "", 10)),
			order("", mergeLine(9, nil, "Alpha", "", "", 15)),
		)
		if err != nil {
			t.Fatal(err)
		}
		if len(plan.updates) != 0 || len(plan.inserts) != 1 {
			t.Fatalf("expected the typed line to be added, got %+v", plan)
		}
	})

	t.Run("unmatched draft lines are added as new lines", func(t *testing.T) {
		beta := intPtr(2)
		plan, err := planPurchaseOrderMerge(
			order("", mergeLine(5, alpha, "Alpha", "", "", 10)),
			order("", mergeLine(9, beta, "Beta", "M", "urgent", 3)),
		)
		if err != nil {
			t.Fatal(err)
		}
		if len(plan.inserts) != 1 {
			t.Fatalf("expected one insert, got %+v", plan)
		}
		got := plan.inserts[0]
		if got.ID != 0 || got.ReceivedQuantity != 0 || got.Quantity != 3 || got.Note != "urgent" ||
			got.ProductID == nil || *got.ProductID != 2 || got.Manufacturer != "M" {
			t.Fatalf("expected a fresh copy of the draft line, got %+v", got)
		}
	})

	t.Run("order notes are joined", func(t *testing.T) {
		cases := []struct{ target, draft, want string }{
			{"A", "B", "A\nB"},
			{"A", "A", "A"},
			{"A", "", "A"},
			{"", "B", "B"},
			{"", "", ""},
		}
		for _, c := range cases {
			plan, err := planPurchaseOrderMerge(
				order(c.target, mergeLine(5, alpha, "Alpha", "", "", 10)),
				order(c.draft, mergeLine(9, alpha, "Alpha", "", "", 10)),
			)
			if err != nil {
				t.Fatal(err)
			}
			if plan.note != c.want {
				t.Fatalf("notes %q + %q: expected %q, got %q", c.target, c.draft, c.want, plan.note)
			}
		}
	})

	t.Run("matched line notes are joined and count as a change", func(t *testing.T) {
		plan, err := planPurchaseOrderMerge(
			order("", mergeLine(5, alpha, "Alpha", "", "box", 10)),
			order("", mergeLine(9, alpha, "Alpha", "", "cold chain", 10)),
		)
		if err != nil {
			t.Fatal(err)
		}
		if len(plan.updates) != 1 {
			t.Fatalf("expected the note change to be an update, got %+v", plan)
		}
		got := plan.updates[0]
		if got.item.Note != "box; cold chain" || got.item.Quantity != 10 ||
			got.quantityBefore != 10 {
			t.Fatalf("expected the notes joined and the quantity unchanged, got %+v", got)
		}
	})

	t.Run("notes that would be too long are refused", func(t *testing.T) {
		longOrderNote := strings.Repeat("x", maxPurchaseOrderNote)
		_, err := planPurchaseOrderMerge(
			order(longOrderNote, mergeLine(5, alpha, "Alpha", "", "", 10)),
			order("y", mergeLine(9, alpha, "Alpha", "", "", 10)),
		)
		requireHTTPStatus(t, err, http.StatusBadRequest)

		longLineNote := strings.Repeat("x", maxPurchaseOrderItemNote)
		_, err = planPurchaseOrderMerge(
			order("", mergeLine(5, alpha, "Alpha", "", longLineNote, 10)),
			order("", mergeLine(9, alpha, "Alpha", "", "y", 10)),
		)
		requireHTTPStatus(t, err, http.StatusBadRequest)
		if !strings.Contains(err.Error(), "Alpha") {
			t.Fatalf("the line error should name the item, got %v", err)
		}
	})
}

func mergeFixture() *fakePurchaseOrderRepo {
	target := types.PurchaseOrder{
		ID: 1, Status: constant.PurchaseOrderOpen, Version: 3, Supplier: "Emzor",
		SupplierID: intPtr(7), Note: "weekly",
		Items: []types.PurchaseOrderItem{mergeLine(11, intPtr(1), "Alpha", "", "", 10)},
	}
	draft := types.PurchaseOrder{
		ID: 2, Status: constant.PurchaseOrderDraft, Version: 2, Supplier: "Other",
		SupplierID: intPtr(8),
		Items: []types.PurchaseOrderItem{
			mergeLine(21, intPtr(1), "Alpha", "", "", 15),
			mergeLine(22, intPtr(2), "Beta", "", "", 4),
		},
	}
	return &fakePurchaseOrderRepo{orders: map[int]types.PurchaseOrder{1: target, 2: draft}}
}

func mergeRequest() types.MergePurchaseOrderRequest {
	return types.MergePurchaseOrderRequest{Version: 3, DraftID: 2, DraftVersion: 2}
}

func TestMergePurchaseOrder(t *testing.T) {
	merge := func(ctx context.Context, repo *fakePurchaseOrderRepo, req types.MergePurchaseOrderRequest) error {
		return NewInventoryService(repo).MergePurchaseOrder(ctx, 1, req, 9)
	}
	requireNothingWritten := func(t *testing.T, repo *fakePurchaseOrderRepo) {
		t.Helper()
		if repo.commits != 0 || len(repo.reverted) != 0 || len(repo.closedIDs) != 0 ||
			len(repo.updatedLines) != 0 || len(repo.insertedLines) != 0 || len(repo.audit) != 0 {
			t.Fatalf("a refused merge must write nothing, got %+v", repo)
		}
	}

	t.Run("a manager merges a draft into an open order", func(t *testing.T) {
		repo := mergeFixture()
		if err := merge(approverContext(), repo, mergeRequest()); err != nil {
			t.Fatal(err)
		}
		if len(repo.reverted) != 1 || repo.reverted[0] != 1 {
			t.Fatalf("expected the open order reverted to draft, got %v", repo.reverted)
		}
		if len(repo.closedIDs) != 1 || repo.closedIDs[0] != 2 {
			t.Fatalf("expected the old draft closed, got %v", repo.closedIDs)
		}
		if len(repo.updatedLines) != 1 || repo.updatedLines[0].ID != 11 ||
			repo.updatedLines[0].Quantity != 15 {
			t.Fatalf("expected Alpha raised to 15, got %+v", repo.updatedLines)
		}
		if len(repo.insertedLines) != 1 || repo.insertedLines[0].Name != "Beta" {
			t.Fatalf("expected Beta added, got %+v", repo.insertedLines)
		}
		if repo.commits != 1 || repo.rollbacks != 0 {
			t.Fatalf(
				"expected one commit, got %d commits %d rollbacks",
				repo.commits,
				repo.rollbacks,
			)
		}
		if len(repo.audit) != 2 {
			t.Fatalf("expected an audit entry on each order, got %d", len(repo.audit))
		}
		for i, wantID := range []int{1, 2} {
			entry := repo.audit[i]
			if entry.Action != constant.AuditMerged || entry.EntityID != wantID {
				t.Fatalf("audit entry %d: expected merged on order %d, got %+v", i, wantID, entry)
			}
		}
	})

	t.Run("the open order keeps its supplier and the orders lock in id order", func(t *testing.T) {
		repo := mergeFixture()
		if err := merge(approverContext(), repo, mergeRequest()); err != nil {
			t.Fatal(err)
		}
		if repo.supplierSaved == nil || *repo.supplierSaved != 7 || repo.supplierSavedN != "Emzor" {
			t.Fatalf("expected the open order's supplier kept, got %v %q",
				repo.supplierSaved, repo.supplierSavedN)
		}
		if len(repo.locked) != 2 || repo.locked[0] != 1 || repo.locked[1] != 2 {
			t.Fatalf("expected locks on 1 then 2, got %v", repo.locked)
		}
		req := types.MergePurchaseOrderRequest{Version: 2, DraftID: 1, DraftVersion: 3}
		swapped := mergeFixture()
		// Order 2 is the open one here and order 1 the draft: still locked low id first.
		o1, o2 := swapped.orders[1], swapped.orders[2]
		o1.Status, o1.Version = constant.PurchaseOrderDraft, 3
		o2.Status, o2.Version = constant.PurchaseOrderOpen, 2
		swapped.orders = map[int]types.PurchaseOrder{1: o1, 2: o2}
		if err := NewInventoryService(
			swapped,
		).MergePurchaseOrder(approverContext(), 2, req, 9); err != nil {
			t.Fatal(err)
		}
		if swapped.locked[0] != 1 || swapped.locked[1] != 2 {
			t.Fatalf("expected locks on 1 then 2, got %v", swapped.locked)
		}
	})

	t.Run("a drafter can merge", func(t *testing.T) {
		if err := merge(drafterContext(), mergeFixture(), mergeRequest()); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("no purchase order permission is refused", func(t *testing.T) {
		repo := mergeFixture()
		requireHTTPStatus(t, merge(bareContext(), repo, mergeRequest()), http.StatusForbidden)
		requireNothingWritten(t, repo)
	})

	t.Run("states that cannot be merged are refused", func(t *testing.T) {
		cases := map[string]func(repo *fakePurchaseOrderRepo){
			"target partial": func(r *fakePurchaseOrderRepo) { setStatus(r, 1, constant.PurchaseOrderPartial) },
			"target with received stock": func(r *fakePurchaseOrderRepo) {
				o := r.orders[1]
				o.Items = []types.PurchaseOrderItem{mergeLine(11, intPtr(1), "Alpha", "", "", 10)}
				o.Items[0].ReceivedQuantity = 1
				r.orders[1] = o
			},
			"target draft":     func(r *fakePurchaseOrderRepo) { setStatus(r, 1, constant.PurchaseOrderDraft) },
			"target submitted": func(r *fakePurchaseOrderRepo) { setStatus(r, 1, constant.PurchaseOrderPending) },
			"target closed":    func(r *fakePurchaseOrderRepo) { setStatus(r, 1, constant.PurchaseOrderClosed) },
			"source open":      func(r *fakePurchaseOrderRepo) { setStatus(r, 2, constant.PurchaseOrderOpen) },
			"source submitted": func(r *fakePurchaseOrderRepo) { setStatus(r, 2, constant.PurchaseOrderPending) },
			"source closed":    func(r *fakePurchaseOrderRepo) { setStatus(r, 2, constant.PurchaseOrderClosed) },
		}
		for name, arrange := range cases {
			repo := mergeFixture()
			arrange(repo)
			requireHTTPStatus(
				t,
				merge(approverContext(), repo, mergeRequest()),
				http.StatusConflict,
			)
			if repo.commits != 0 || len(repo.reverted) != 0 || len(repo.closedIDs) != 0 {
				t.Fatalf("%s: a refused merge must write nothing", name)
			}
		}
	})

	t.Run("stale versions are refused", func(t *testing.T) {
		for name, req := range map[string]types.MergePurchaseOrderRequest{
			"open order": {Version: 2, DraftID: 2, DraftVersion: 2},
			"draft":      {Version: 3, DraftID: 2, DraftVersion: 1},
		} {
			repo := mergeFixture()
			requireHTTPStatus(t, merge(approverContext(), repo, req), http.StatusConflict)
			if repo.commits != 0 || len(repo.reverted) != 0 {
				t.Fatalf("%s: a stale merge must write nothing", name)
			}
		}
	})

	t.Run("the same order twice or no draft is a bad request", func(t *testing.T) {
		for _, draftID := range []int{1, 0, -4} {
			repo := mergeFixture()
			req := types.MergePurchaseOrderRequest{Version: 3, DraftID: draftID, DraftVersion: 2}
			requireHTTPStatus(t, merge(approverContext(), repo, req), http.StatusBadRequest)
			requireNothingWritten(t, repo)
		}
	})

	t.Run("an unknown order is not found", func(t *testing.T) {
		repo := mergeFixture()
		req := types.MergePurchaseOrderRequest{Version: 3, DraftID: 99, DraftVersion: 2}
		requireHTTPStatus(t, merge(approverContext(), repo, req), http.StatusNotFound)
		requireNothingWritten(t, repo)
	})

	t.Run("notes that pass the limit write nothing", func(t *testing.T) {
		repo := mergeFixture()
		d := repo.orders[2]
		d.Note = strings.Repeat("x", maxPurchaseOrderNote)
		repo.orders[2] = d
		requireHTTPStatus(t, merge(approverContext(), repo, mergeRequest()), http.StatusBadRequest)
		requireNothingWritten(t, repo)
	})
}

func setStatus(repo *fakePurchaseOrderRepo, id int, status string) {
	order := repo.orders[id]
	order.Status = status
	repo.orders[id] = order
}
