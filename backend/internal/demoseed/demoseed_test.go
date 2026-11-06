package demoseed_test

import (
	"testing"

	"pharmacy/internal/demoseed"
	"pharmacy/internal/testdb"
)

func TestCheckTargetOnlyAllowsThisMachine(t *testing.T) {
	ok := []string{
		"postgres://u:p@localhost:5544/db?sslmode=disable",
		"postgres://u:p@127.0.0.1:5432/db",
		"postgres://u:p@[::1]:5432/db",
		"postgres://u:p@localhost/db?sslmode=disable",
	}
	for _, dsn := range ok {
		if err := demoseed.CheckTarget(dsn); err != nil {
			t.Errorf("%s should be allowed: %v", dsn, err)
		}
	}
	refused := []string{
		"postgres://u:p@db.example.com:5432/db",
		"postgres://u:p@10.0.0.5:5432/db",
		"postgres://u:p@localhost.example.com/db",
		"postgres://u:p@localhost/db?host=prod.example.com",
		"postgres://u:p@localhost/db?hostaddr=10.0.0.5&sslmode=disable",
		"host=prod.example.com user=u dbname=db",
		"not a dsn",
		"",
	}
	for _, dsn := range refused {
		if err := demoseed.CheckTarget(dsn); err == nil {
			t.Errorf("%q should be refused", dsn)
		}
	}
}

func TestSeedFillsAnEmptyDatabaseOnceOnly(t *testing.T) {
	db := testdb.New(t, 0)

	summary, err := demoseed.Run(db)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Products < 20 || summary.Sales < 8 {
		t.Fatalf("expected a believable shop, got %+v", summary)
	}
	for query, want := range map[string]int{
		`SELECT COUNT(*) FROM users`: 3,
		`SELECT COUNT(*) FROM purchase_order WHERE status = 'pending_approval'`: 1,
		`SELECT COUNT(*) FROM stock_flag WHERE status = 'open'`:                 1,
		`SELECT COUNT(*) FROM stock_adjustment`:                                 1,
		`SELECT COUNT(*) FROM returns`:                                          1,
	} {
		var got int
		if err := db.Get(&got, query); err != nil || got != want {
			t.Errorf("%s: got %d (%v), want %d", query, got, err, want)
		}
	}

	if _, err := demoseed.Run(db); err == nil {
		t.Fatal("a database that already has users must be refused")
	}
}
