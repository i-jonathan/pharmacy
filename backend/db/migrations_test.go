package db_test

import (
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"

	"pharmacy/internal/testdb"
)

// The purchase order migration reshapes live data (it backfills suppliers from receipts), so it
// is exercised against the schema as production has it: everything up to 39 applied, with
// receipts already in place.
const (
	purchaseOrderMigration   = 40
	auditLogMigration        = 41
	approvalMigration        = 42
	permissionsMigration     = 43
	stockAdjustmentMigration = 44
)

func count(t *testing.T, db *sqlx.DB, query string, args ...any) int {
	t.Helper()
	var n int
	if err := db.Get(&n, query, args...); err != nil {
		t.Fatalf("%s: %v", query, err)
	}
	return n
}

func mustFail(t *testing.T, db *sqlx.DB, wantInError, query string, args ...any) {
	t.Helper()
	_, err := db.Exec(query, args...)
	if err == nil {
		t.Fatalf("expected %q to fail", query)
	}
	if wantInError != "" && !strings.Contains(err.Error(), wantInError) {
		t.Fatalf("expected error containing %q, got %v", wantInError, err)
	}
}

func seedReceipts(t *testing.T, db *sqlx.DB) {
	t.Helper()
	db.MustExec(`INSERT INTO users (username, password) VALUES ('legacy', 'x')`)
	db.MustExec(`
		INSERT INTO receiving_batch (supplier_name, received_by_id)
		SELECT name, (SELECT id FROM users LIMIT 1)
		FROM unnest(ARRAY['Emzor', 'emzor ', 'Emzor', 'Fidson', '  FIDSON', '']) AS name`)
}

func TestPurchaseOrderMigrationBackfillsSuppliers(t *testing.T) {
	db := testdb.New(t, purchaseOrderMigration-1)
	seedReceipts(t, db)
	testdb.Migrate(t, db, purchaseOrderMigration-1, purchaseOrderMigration)

	var suppliers []string
	if err := db.Select(&suppliers, `SELECT name FROM supplier ORDER BY name`); err != nil {
		t.Fatal(err)
	}
	// Spellings differing only by case or spaces collapse into one supplier. The most used
	// spelling wins; with a tie the C-collation order decides.
	if len(suppliers) != 2 || suppliers[0] != "Emzor" || suppliers[1] != "FIDSON" {
		t.Fatalf("unexpected suppliers: %q", suppliers)
	}

	if n := count(t, db, `SELECT count(*) FROM receiving_batch WHERE supplier_id IS NULL`); n != 1 {
		t.Fatalf("only the blank-named receipt should stay unlinked, got %d", n)
	}
	if n := count(t, db, `
		SELECT count(*) FROM receiving_batch rb JOIN supplier s ON s.id = rb.supplier_id
		WHERE lower(btrim(rb.supplier_name)) = lower(btrim(s.name))`); n != 5 {
		t.Fatalf("receipts should link to the supplier with their name, got %d", n)
	}
	if n := count(t, db, `SELECT count(*) FROM receiving_batch`); n != 6 {
		t.Fatalf("no receipt may be lost, got %d", n)
	}
	var original string
	if err := db.Get(&original, `SELECT supplier_name FROM receiving_batch WHERE supplier_name LIKE '  %'`); err != nil || original != "  FIDSON" {
		t.Fatalf("the name written on a receipt must not be rewritten: %q %v", original, err)
	}
}

func TestPurchaseOrderMigrationConstraints(t *testing.T) {
	db := testdb.New(t, 0)
	db.MustExec(`INSERT INTO users (username, password) VALUES ('u', 'x')`)

	db.MustExec(`INSERT INTO supplier (name) VALUES ('Emzor')`)
	mustFail(t, db, "supplier_name_unique_idx", `INSERT INTO supplier (name) VALUES ('  EMZOR ')`)
	mustFail(t, db, "", `INSERT INTO supplier (name) VALUES ('   ')`)

	mustFail(t, db, "purchase_order_status_check", `INSERT INTO purchase_order (status) VALUES ('bogus')`)
	for _, status := range []string{"open", "partial", "received", "closed"} {
		db.MustExec(`INSERT INTO purchase_order (status) VALUES ($1)`, status)
	}
	if v := count(t, db, `SELECT version FROM purchase_order LIMIT 1`); v != 1 {
		t.Fatalf("orders start at version 1, got %d", v)
	}

	var orderID int
	db.Get(&orderID, `SELECT id FROM purchase_order LIMIT 1`)
	mustFail(t, db, "", `INSERT INTO purchase_order_item (purchase_order_id, item_name, quantity) VALUES ($1, 'x', 0)`, orderID)
	mustFail(t, db, "", `INSERT INTO purchase_order_item (purchase_order_id, item_name, quantity, received_quantity) VALUES ($1, 'x', 5, 6)`, orderID)
	db.MustExec(`INSERT INTO purchase_order_item (purchase_order_id, item_name, quantity, received_quantity) VALUES ($1, 'x', 5, 5)`, orderID)

	// An order's leftovers can be taken over by only one other order.
	db.MustExec(`INSERT INTO purchase_order (source_purchase_order_id) VALUES ($1)`, orderID)
	mustFail(t, db, "purchase_order_source_unique_idx", `INSERT INTO purchase_order (source_purchase_order_id) VALUES ($1)`, orderID)
}

func TestAuditLogIsAppendOnly(t *testing.T) {
	db := testdb.New(t, 0)
	db.MustExec(`INSERT INTO audit_log (entity_type, entity_id, action) VALUES ('purchase_order', 1, 'created')`)

	mustFail(t, db, "append-only", `UPDATE audit_log SET action = 'edited'`)
	mustFail(t, db, "append-only", `DELETE FROM audit_log`)
	if n := count(t, db, `SELECT count(*) FROM audit_log`); n != 1 {
		t.Fatalf("the entry must survive, got %d rows", n)
	}
	if action := count(t, db, `SELECT count(*) FROM audit_log WHERE action = 'created'`); action != 1 {
		t.Fatal("the entry must be unchanged")
	}
}

func TestPurchaseOrderMigrationRollsBackCleanly(t *testing.T) {
	db := testdb.New(t, purchaseOrderMigration-1)
	seedReceipts(t, db)
	testdb.Migrate(t, db, purchaseOrderMigration-1, purchaseOrderMigration)

	testdb.Rollback(t, db, purchaseOrderMigration)

	for _, table := range []string{"supplier", "purchase_order", "purchase_order_item"} {
		if n := count(t, db, `SELECT count(*) FROM information_schema.tables WHERE table_name = $1`, table); n != 0 {
			t.Fatalf("table %s should be gone after rollback", table)
		}
	}
	for _, column := range [][2]string{{"receiving_batch", "supplier_id"}, {"product_batch", "purchase_order_item_id"}} {
		q := `SELECT count(*) FROM information_schema.columns WHERE table_name = $1 AND column_name = $2`
		if n := count(t, db, q, column[0], column[1]); n != 0 {
			t.Fatalf("column %s.%s should be gone after rollback", column[0], column[1])
		}
	}
	if n := count(t, db, `SELECT count(*) FROM receiving_batch`); n != 6 {
		t.Fatalf("rolling back must keep the existing receipts, got %d", n)
	}

	// And it applies again on top of the rolled-back schema.
	testdb.Migrate(t, db, purchaseOrderMigration-1, purchaseOrderMigration)
	if n := count(t, db, `SELECT count(*) FROM supplier`); n != 2 {
		t.Fatalf("re-applying should backfill again, got %d suppliers", n)
	}
}

func TestAuditLogMigrationRollsBackCleanly(t *testing.T) {
	db := testdb.New(t, 0)
	db.MustExec(`INSERT INTO audit_log (entity_type, entity_id, action) VALUES ('purchase_order', 1, 'created')`)

	testdb.Rollback(t, db, auditLogMigration)
	if n := count(t, db, `SELECT count(*) FROM information_schema.tables WHERE table_name = 'audit_log'`); n != 0 {
		t.Fatal("audit_log should be gone after rollback")
	}
	if n := count(t, db, `SELECT count(*) FROM pg_proc WHERE proname = 'audit_log_append_only'`); n != 0 {
		t.Fatal("the trigger function should be gone after rollback")
	}
	// The purchase order schema from the previous migration is untouched.
	if n := count(t, db, `SELECT count(*) FROM information_schema.tables WHERE table_name = 'purchase_order'`); n != 1 {
		t.Fatal("purchase orders must survive rolling back the audit log")
	}

	testdb.Migrate(t, db, auditLogMigration-1, auditLogMigration)
	db.MustExec(`INSERT INTO audit_log (entity_type, entity_id, action) VALUES ('purchase_order', 2, 'created')`)
}

func TestApprovalMigrationAddsDraftStatesAndRollsBackSafely(t *testing.T) {
	db := testdb.New(t, approvalMigration-1)
	db.MustExec(`INSERT INTO users (username, password) VALUES ('u', 'x')`)
	mustFail(t, db, "purchase_order_status_check", `INSERT INTO purchase_order (status) VALUES ('draft')`)
	db.MustExec(`INSERT INTO purchase_order (status, created_by_id) SELECT s, (SELECT id FROM users) FROM unnest(ARRAY['open', 'closed']) s`)

	testdb.Migrate(t, db, approvalMigration-1, approvalMigration)
	// Orders from before approval existed count as approved by their author.
	if n := count(t, db, `SELECT count(*) FROM purchase_order WHERE approved_by_id = created_by_id AND approved_at IS NOT NULL`); n != 2 {
		t.Fatalf("existing orders should be marked approved, got %d", n)
	}
	for _, status := range []string{"draft", "pending_approval", "open"} {
		db.MustExec(`INSERT INTO purchase_order (status, created_by_id) VALUES ($1, (SELECT id FROM users))`, status)
	}
	mustFail(t, db, "purchase_order_status_check", `INSERT INTO purchase_order (status) VALUES ('approved')`)
	db.MustExec(`UPDATE purchase_order SET approved_by_id = (SELECT id FROM users), approved_at = NOW() WHERE status = 'open'`)

	// Rolling back closes the orders that were never approved and keeps every row.
	testdb.Rollback(t, db, approvalMigration)
	if n := count(t, db, `SELECT count(*) FROM purchase_order`); n != 5 {
		t.Fatalf("no order may be lost on rollback, got %d", n)
	}
	if n := count(t, db, `SELECT count(*) FROM purchase_order WHERE status IN ('draft', 'pending_approval')`); n != 0 {
		t.Fatalf("drafts cannot exist under the old rules, got %d", n)
	}
	if n := count(t, db, `SELECT count(*) FROM information_schema.columns WHERE table_name = 'purchase_order' AND column_name = 'approved_by_id'`); n != 0 {
		t.Fatal("approved_by_id should be gone after rollback")
	}

	testdb.Migrate(t, db, approvalMigration-1, approvalMigration)
	db.MustExec(`INSERT INTO purchase_order (status) VALUES ('draft')`)
}

func TestPurchaseOrderPermissionsMigration(t *testing.T) {
	db := testdb.New(t, permissionsMigration-1)

	granted := func() []string {
		t.Helper()
		var rows []string
		err := db.Select(&rows, `
			SELECT r.name || ':' || p.action
			FROM role_permissions rp
			JOIN roles r ON r.id = rp.role_id
			JOIN permissions p ON p.id = rp.permission_id
			WHERE p.resource = 'PurchaseOrder'
			ORDER BY 1`)
		if err != nil {
			t.Fatal(err)
		}
		return rows
	}
	if len(granted()) != 0 {
		t.Fatal("nothing should be granted before the migration")
	}

	testdb.Migrate(t, db, permissionsMigration-1, permissionsMigration)
	want := "Admin:Approve,Cashier:Draft,Manager:Approve"
	if got := strings.Join(granted(), ","); got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
	// The permissions the app checks for are named exactly as it builds them at login:
	// lowercase resource:action.
	var keys []string
	if err := db.Select(&keys, `SELECT lower(resource || ':' || action) FROM permissions WHERE resource = 'PurchaseOrder' ORDER BY 1`); err != nil {
		t.Fatal(err)
	}
	if strings.Join(keys, ",") != "purchaseorder:approve,purchaseorder:draft" {
		t.Fatalf("unexpected permission keys %v", keys)
	}

	// Running it again, for example after an administrator already created the permissions by
	// hand, adds nothing twice.
	db.MustExec(`DELETE FROM role_permissions rp USING permissions p, roles r
		WHERE rp.permission_id = p.id AND rp.role_id = r.id AND p.action = 'Approve' AND r.name = 'Admin'`)
	run := func() {
		testdb.Rollback(t, db, permissionsMigration)
		testdb.Migrate(t, db, permissionsMigration-1, permissionsMigration)
	}
	run()
	if got := strings.Join(granted(), ","); got != want {
		t.Fatalf("expected %s after re-applying, got %s", want, got)
	}
	testdb.Migrate(t, db, permissionsMigration-1, permissionsMigration)
	if got := strings.Join(granted(), ","); got != want {
		t.Fatalf("applying twice must not duplicate grants: %s", got)
	}

	testdb.Rollback(t, db, permissionsMigration)
	if len(granted()) != 0 {
		t.Fatal("rollback should remove the grants")
	}
	if n := count(t, db, `SELECT count(*) FROM permissions WHERE resource = 'PurchaseOrder'`); n != 0 {
		t.Fatalf("rollback should remove the permissions, got %d", n)
	}
}

func TestStockAdjustmentMigration(t *testing.T) {
	db := testdb.New(t, stockAdjustmentMigration-1)
	db.MustExec(`INSERT INTO users (username, password) VALUES ('m', 'x')`)
	db.MustExec(`INSERT INTO category (name) VALUES ('General')`)
	db.MustExec(`INSERT INTO product (name, category_id) SELECT 'Amoxil', id FROM category`)

	testdb.Migrate(t, db, stockAdjustmentMigration-1, stockAdjustmentMigration)

	var grants []string
	err := db.Select(&grants, `
		SELECT r.name || ':' || p.action
		FROM role_permissions rp
		JOIN roles r ON r.id = rp.role_id
		JOIN permissions p ON p.id = rp.permission_id
		WHERE p.resource = 'Stock' AND p.action IN ('Adjust', 'Flag')
		ORDER BY 1`)
	if err != nil {
		t.Fatal(err)
	}
	want := "Admin:Adjust,Admin:Flag,Cashier:Flag,Manager:Adjust,Manager:Flag"
	if got := strings.Join(grants, ","); got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}

	addAdjustment := `INSERT INTO stock_adjustment (reason, created_by_id)
		VALUES ($1, (SELECT id FROM users))`
	db.MustExec(addAdjustment, "damaged")
	mustFail(t, db, "stock_adjustment_reason_check", addAdjustment, "borrowed")
	// An item has to change the quantity or the expiry.
	mustFail(t, db, "stock_adjustment_item_changes_something", `
		INSERT INTO stock_adjustment_item
			(stock_adjustment_id, product_id, quantity_before, quantity_change)
		VALUES ((SELECT id FROM stock_adjustment), (SELECT id FROM product), 5, 0)`)
	db.MustExec(`
		INSERT INTO stock_adjustment_item
			(stock_adjustment_id, product_id, quantity_before, quantity_change)
		VALUES ((SELECT id FROM stock_adjustment), (SELECT id FROM product), 5, -2)`)
	db.MustExec(`INSERT INTO stock_flag (product_id, note, flagged_by_id)
		VALUES ((SELECT id FROM product), 'two packs crushed', (SELECT id FROM users))`)
	mustFail(t, db, "stock_flag_status_check", `UPDATE stock_flag SET status = 'maybe'`)
	mustFail(t, db, "stock_flag_one_open_per_user", `
		INSERT INTO stock_flag (product_id, note, flagged_by_id)
		VALUES ((SELECT id FROM product), 'again', (SELECT id FROM users))`)

	testdb.Rollback(t, db, stockAdjustmentMigration)
	tables := `SELECT count(*) FROM information_schema.tables
		WHERE table_name IN ('stock_adjustment', 'stock_adjustment_item', 'stock_flag')`
	if n := count(t, db, tables); n != 0 {
		t.Fatalf("the adjustment tables should be gone after rollback, %d left", n)
	}
	newPerms := `SELECT count(*) FROM permissions
		WHERE resource = 'Stock' AND action IN ('Adjust', 'Flag')`
	if n := count(t, db, newPerms); n != 0 {
		t.Fatal("the new permissions should be removed on rollback")
	}
	kept := `SELECT count(*) FROM permissions WHERE resource = 'Stock' AND action = 'Complete'`
	if n := count(t, db, kept); n != 1 {
		t.Fatal("existing stock permissions must survive the rollback")
	}

	testdb.Migrate(t, db, stockAdjustmentMigration-1, stockAdjustmentMigration)
	db.MustExec(addAdjustment, "expired")
}
