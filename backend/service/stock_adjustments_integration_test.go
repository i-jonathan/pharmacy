package service_test

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"pharmacy/internal/constant"
	"pharmacy/internal/types"
	"pharmacy/repository"
)

func stockItem(productID int, mode string, quantity int) types.StockAdjustmentItemRequest {
	return types.StockAdjustmentItemRequest{ProductID: productID, Mode: mode, Quantity: quantity}
}

func TestStockAdjustmentAgainstTheDatabase(t *testing.T) {
	f := newFixture(t)
	manager := asUser(constant.AdjustStockPermissionKey, constant.FlagStockPermissionKey)
	cashier := asUser(constant.FlagStockPermissionKey)
	amoxil := f.product(t, "Amoxil")

	// Opening stock of 12.
	id, err := f.svc.CreateStockAdjustment(manager, types.StockAdjustmentRequest{
		Reason: "opening_stock",
		Items:  []types.StockAdjustmentItemRequest{stockItem(amoxil, "add", 12)},
	}, f.uid)
	if err != nil {
		t.Fatal(err)
	}
	if got := f.stock(t, amoxil); got != 12 {
		t.Fatalf("expected 12 in stock after opening stock, got %d", got)
	}

	// A cashier flags it; the manager counts 10 and adjusts, resolving the flag.
	counted := 10
	flagID, err := f.svc.FlagStock(cashier, types.StockFlagRequest{
		ProductID: amoxil, Note: "two packs crushed", CountedQuantity: &counted,
	}, f.uid)
	if err != nil {
		t.Fatal(err)
	}
	if n, _ := f.svc.CountStockFlags(manager); n != 1 {
		t.Fatalf("expected one open flag, got %d", n)
	}
	expiry := "2027-06-30"
	req := types.StockAdjustmentRequest{
		Reason: "damaged", IdempotencyKey: "adj-1", FlagIDs: []int{flagID},
		Items: []types.StockAdjustmentItemRequest{
			{ProductID: amoxil, Mode: "set", Quantity: 10, Expiry: &expiry},
		},
	}
	second, err := f.svc.CreateStockAdjustment(manager, req, f.uid)
	if err != nil {
		t.Fatal(err)
	}
	if got := f.stock(t, amoxil); got != 10 {
		t.Fatalf("expected 10 after setting the count, got %d", got)
	}
	var flagStatus string
	f.db.Get(&flagStatus, `SELECT status FROM stock_flag WHERE id = $1`, flagID)
	if flagStatus != "resolved" {
		t.Fatalf("the flag should be resolved, got %q", flagStatus)
	}
	var currentExpiry string
	f.db.Get(&currentExpiry, `SELECT current_expiry::text FROM product WHERE id = $1`, amoxil)
	if currentExpiry != "2027-06-30" {
		t.Fatalf("expected the expiry to be corrected, got %q", currentExpiry)
	}

	// Retrying the same request returns the same adjustment and changes nothing.
	retried, err := f.svc.CreateStockAdjustment(manager, req, f.uid)
	if err != nil || retried != second {
		t.Fatalf("a retry should return adjustment %d, got %d, %v", second, retried, err)
	}
	if got := f.stock(t, amoxil); got != 10 {
		t.Fatalf("a retry must not move stock again, got %d", got)
	}

	// Removing more than is in stock is refused.
	_, err = f.svc.CreateStockAdjustment(manager, types.StockAdjustmentRequest{
		Reason: "theft_loss",
		Items:  []types.StockAdjustmentItemRequest{stockItem(amoxil, "remove", 11)},
	}, f.uid)
	if status(err) != http.StatusBadRequest {
		t.Fatalf("removing 11 of 10 should be a 400, got %v", err)
	}

	// The history shows both adjustments, newest first, with before and after.
	page, err := f.svc.ListStockAdjustments(manager, 1)
	if err != nil {
		t.Fatal(err)
	}
	if page.TotalCount != 2 || page.Data[0].ID != second || page.Data[1].ID != id {
		t.Fatalf("unexpected history %+v", page)
	}
	latest := page.Data[0].Items[0]
	if latest.QuantityBefore != 12 || latest.QuantityChange != -2 ||
		latest.ProductName != "Amoxil" {
		t.Fatalf("unexpected item %+v", latest)
	}

	// A second flag is dismissed with a note.
	other, _ := f.svc.FlagStock(
		cashier,
		types.StockFlagRequest{ProductID: amoxil, Note: "looks low"},
		f.uid,
	)
	if err := f.svc.DismissStockFlag(manager, other, "counted, it's right", f.uid); err != nil {
		t.Fatal(err)
	}
	if err := f.svc.DismissStockFlag(
		manager,
		other,
		"again",
		f.uid,
	); status(
		err,
	) != http.StatusConflict {
		t.Fatalf("dismissing twice should be a 409, got %v", err)
	}

	var audits int
	f.db.Get(
		&audits,
		`SELECT count(*) FROM audit_log WHERE entity_type IN ('stock_adjustment', 'stock_flag')`,
	)
	// 2 adjustments + flag created, resolved + flag created, dismissed.
	if audits != 6 {
		t.Fatalf("expected 6 audit entries, got %d", audits)
	}
}

func TestConcurrentStockAdjustmentsQueueOnTheProduct(t *testing.T) {
	f := newFixture(t)
	manager := asUser(constant.AdjustStockPermissionKey)
	product := f.product(t, "Paracetamol")
	if _, err := f.svc.CreateStockAdjustment(manager, types.StockAdjustmentRequest{
		Reason: "opening_stock",
		Items:  []types.StockAdjustmentItemRequest{stockItem(product, "add", 5)},
	}, f.uid); err != nil {
		t.Fatal(err)
	}

	// Eight managers each remove 1 at the same moment: exactly five can succeed, and stock never
	// goes below zero.
	var wg sync.WaitGroup
	var mu sync.Mutex
	succeeded := 0
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := f.svc.CreateStockAdjustment(manager, types.StockAdjustmentRequest{
				Reason: "internal_use",
				Items:  []types.StockAdjustmentItemRequest{stockItem(product, "remove", 1)},
			}, f.uid)
			if err == nil {
				mu.Lock()
				succeeded++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if succeeded != 5 {
		t.Fatalf("exactly 5 removals fit 5 in stock, %d succeeded", succeeded)
	}
	if got := f.stock(t, product); got != 0 {
		t.Fatalf("expected 0 left, got %d", got)
	}
}

func TestConcurrentRetriesOfOneAdjustmentApplyOnce(t *testing.T) {
	f := newFixture(t)
	manager := asUser(constant.AdjustStockPermissionKey)
	product := f.product(t, "Ibuprofen")
	if _, err := f.svc.CreateStockAdjustment(manager, types.StockAdjustmentRequest{
		Reason: "opening_stock",
		Items:  []types.StockAdjustmentItemRequest{stockItem(product, "add", 5)},
	}, f.uid); err != nil {
		t.Fatal(err)
	}

	// The same save sent four times at once (a double click plus retries): every copy must get
	// the one adjustment back, and stock must drop by 5 once.
	req := types.StockAdjustmentRequest{
		Reason: "expired", IdempotencyKey: "same-save",
		Items: []types.StockAdjustmentItemRequest{stockItem(product, "remove", 5)},
	}
	var wg sync.WaitGroup
	ids := make([]int, 4)
	errs := make([]error, 4)
	for i := range ids {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ids[i], errs[i] = f.svc.CreateStockAdjustment(manager, req, f.uid)
		}(i)
	}
	wg.Wait()

	for i := range ids {
		if errs[i] != nil || ids[i] != ids[0] {
			t.Fatalf("every copy should return adjustment %d, got %v", ids[0], ids)
		}
	}
	if got := f.stock(t, product); got != 0 {
		t.Fatalf("the removal must apply once, stock is %d", got)
	}
}

func TestAFlagRemembersTheStockWhenItWasRaised(t *testing.T) {
	f := newFixture(t)
	manager := asUser(constant.AdjustStockPermissionKey, constant.FlagStockPermissionKey)
	product := f.product(t, "Vitamin C")
	if _, err := f.svc.CreateStockAdjustment(manager, types.StockAdjustmentRequest{
		Reason: "opening_stock",
		Items:  []types.StockAdjustmentItemRequest{stockItem(product, "add", 12)},
	}, f.uid); err != nil {
		t.Fatal(err)
	}
	counted := 10
	if _, err := f.svc.FlagStock(manager, types.StockFlagRequest{
		ProductID: product, Note: "short", CountedQuantity: &counted,
	}, f.uid); err != nil {
		t.Fatal(err)
	}
	// Four sold after the count.
	if _, err := f.svc.CreateStockAdjustment(manager, types.StockAdjustmentRequest{
		Reason: "internal_use",
		Items:  []types.StockAdjustmentItemRequest{stockItem(product, "remove", 4)},
	}, f.uid); err != nil {
		t.Fatal(err)
	}

	flags, err := f.svc.ListStockFlags(manager)
	if err != nil || len(flags) != 1 {
		t.Fatalf("expected one open flag, got %v, %v", flags, err)
	}
	if flags[0].StockAtFlag != 12 || flags[0].CurrentStock != 8 {
		t.Fatalf("expected stock 12 when flagged and 8 now, got %d and %d",
			flags[0].StockAtFlag, flags[0].CurrentStock)
	}
}

func TestAFlagEstimateCountsSalesNotCorrections(t *testing.T) {
	f := newFixture(t)
	manager := asUser(constant.AdjustStockPermissionKey, constant.FlagStockPermissionKey)
	product := f.product(t, "Vitamin C")
	if _, err := f.svc.CreateStockAdjustment(manager, types.StockAdjustmentRequest{
		Reason: "opening_stock",
		Items:  []types.StockAdjustmentItemRequest{stockItem(product, "add", 12)},
	}, f.uid); err != nil {
		t.Fatal(err)
	}
	counted := 10
	if _, err := f.svc.FlagStock(manager, types.StockFlagRequest{
		ProductID: product, Note: "short", CountedQuantity: &counted,
	}, f.uid); err != nil {
		t.Fatal(err)
	}
	// After the count: 3 sold and 5 received (shelf movements), and a correction of -2 from a
	// stock taking, which the cashier's count already reflected.
	f.db.MustExec(`INSERT INTO stock_movement (product_id, reference_id, movement_type, quantity)
		VALUES ($1, 1, 'OUT_SALE', 3), ($1, 1, 'IN_PURCHASE', 5), ($1, 1, 'OUT_STOCK_TAKING', 2)`,
		product)

	flags, err := f.svc.ListStockFlags(manager)
	if err != nil || len(flags) != 1 {
		t.Fatalf("expected one open flag, got %v, %v", flags, err)
	}
	if flags[0].MovedSince != 2 {
		t.Fatalf("expected +2 moved since the count (5 in, 3 sold), got %d", flags[0].MovedSince)
	}
}

func TestOnePersonHasOneOpenFlagPerProduct(t *testing.T) {
	f := newFixture(t)
	cashier := asUser(constant.FlagStockPermissionKey)
	manager := asUser(constant.AdjustStockPermissionKey, constant.FlagStockPermissionKey)
	product := f.product(t, "Amoxil")

	first, err := f.svc.FlagStock(cashier, types.StockFlagRequest{
		ProductID: product, Note: "two packs crushed",
	}, f.uid)
	if err != nil {
		t.Fatal(err)
	}
	// A double click, or the same person flagging again before a manager has looked.
	_, err = f.svc.FlagStock(cashier, types.StockFlagRequest{
		ProductID: product, Note: "two packs crushed",
	}, f.uid)
	if status(err) != http.StatusConflict {
		t.Fatalf("a second open flag from the same person should be a 409, got %v", err)
	}

	// Once the first is handled, the product can be flagged again.
	if err := f.svc.DismissStockFlag(manager, first, "counted, fine", f.uid); err != nil {
		t.Fatal(err)
	}
	if _, err := f.svc.FlagStock(cashier, types.StockFlagRequest{
		ProductID: product, Note: "short again",
	}, f.uid); err != nil {
		t.Fatalf("flagging after the first was dismissed should work, got %v", err)
	}
}

func TestAProductCountedInAnOpenStockTakingCannotBeAdjusted(t *testing.T) {
	f := newFixture(t)
	manager := asUser(constant.AdjustStockPermissionKey)
	counted := f.product(t, "Amoxil")
	notYet := f.product(t, "Paracetamol")
	for _, p := range []int{counted, notYet} {
		if _, err := f.svc.CreateStockAdjustment(manager, types.StockAdjustmentRequest{
			Reason: "opening_stock",
			Items:  []types.StockAdjustmentItemRequest{stockItem(p, "add", 50)},
		}, f.uid); err != nil {
			t.Fatal(err)
		}
	}
	// A stock taking is open and Amoxil has been counted, so its snapshot (50) is taken. Writing
	// off 5 now would be counted twice when the stock taking completes (count 45 - snapshot 50).
	var takingID int
	if err := f.db.Get(&takingID, `INSERT INTO stock_taking (name, status, created_by_id)
		VALUES ('APRIL', 'in_progress', $1) RETURNING id`, f.uid); err != nil {
		t.Fatal(err)
	}
	f.db.MustExec(`INSERT INTO stock_taking_item (stock_taking_id, product_id, snapshot_quantity)
		VALUES ($1, $2, 50)`, takingID, counted)

	_, err := f.svc.CreateStockAdjustment(manager, types.StockAdjustmentRequest{
		Reason: "expired",
		Items:  []types.StockAdjustmentItemRequest{stockItem(counted, "remove", 5)},
	}, f.uid)
	if status(err) != http.StatusConflict {
		t.Fatalf("adjusting a product already counted should be a 409, got %v", err)
	}

	// A product not counted yet has no snapshot, so it can still be adjusted.
	if _, err := f.svc.CreateStockAdjustment(manager, types.StockAdjustmentRequest{
		Reason: "expired",
		Items:  []types.StockAdjustmentItemRequest{stockItem(notYet, "remove", 5)},
	}, f.uid); err != nil {
		t.Fatalf("a product not yet counted should adjust, got %v", err)
	}

	// Once the stock taking is completed, Amoxil can be adjusted again.
	f.db.MustExec(`UPDATE stock_taking SET status = 'completed' WHERE id = $1`, takingID)
	if _, err := f.svc.CreateStockAdjustment(manager, types.StockAdjustmentRequest{
		Reason: "expired",
		Items:  []types.StockAdjustmentItemRequest{stockItem(counted, "remove", 5)},
	}, f.uid); err != nil {
		t.Fatalf("after the stock taking, adjusting should work, got %v", err)
	}
}

func TestARetryKeyBelongsToThePersonWhoSentIt(t *testing.T) {
	f := newFixture(t)
	manager := asUser(constant.AdjustStockPermissionKey)
	product := f.product(t, "Amoxil")
	var other int
	if err := f.db.Get(&other, `INSERT INTO users (username, password) VALUES ('second', 'x')
		RETURNING id`); err != nil {
		t.Fatal(err)
	}
	req := types.StockAdjustmentRequest{
		Reason: "opening_stock", IdempotencyKey: "same-key",
		Items: []types.StockAdjustmentItemRequest{stockItem(product, "add", 5)},
	}

	first, err := f.svc.CreateStockAdjustment(manager, req, f.uid)
	if err != nil {
		t.Fatal(err)
	}
	// Another manager's save that happens to carry the same key is their own adjustment, not a
	// retry of the first one.
	second, err := f.svc.CreateStockAdjustment(manager, req, other)
	if err != nil || second == first {
		t.Fatalf("expected a second adjustment, got %d (first %d), %v", second, first, err)
	}
	if got := f.stock(t, product); got != 10 {
		t.Fatalf("both adjustments should apply, stock is %d", got)
	}
	// The same person's retry still returns their adjustment.
	if again, err := f.svc.CreateStockAdjustment(
		manager,
		req,
		other,
	); err != nil ||
		again != second {
		t.Fatalf("a retry should return %d, got %d, %v", second, again, err)
	}
}

func TestAnAdjustmentLockDoesNotBlockSalesOfTheProduct(t *testing.T) {
	f := newFixture(t)
	product := f.product(t, "Amoxil")
	store := repository.NewStore(f.db)
	ctx := context.Background()

	// An adjustment holds its product lock.
	tx, err := store.BeginTx(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer store.RollbackTx(tx)
	if _, err := store.LockProductStockTx(ctx, tx, []int{product}); err != nil {
		t.Fatal(err)
	}

	// A sale's stock movement references the product (a foreign-key check on its row); it must not
	// wait for the adjustment to finish.
	done := make(chan error, 1)
	go func() {
		_, err := f.db.Exec(`INSERT INTO stock_movement (product_id, reference_id, movement_type,
			quantity) VALUES ($1, 1, 'OUT_SALE', 1)`, product)
		done <- err
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("a sale's movement waited on the adjustment's product lock")
	}
}
