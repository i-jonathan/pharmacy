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
	purchaseOrderMigration = 40
	auditLogMigration      = 41
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
