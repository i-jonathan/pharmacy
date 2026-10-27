// Package testdb gives integration tests a throwaway Postgres database with the real migrations
// applied. Tests using it are skipped unless PHARMACY_TEST_DSN is set to a connection string
// for a server where the user may create and drop databases, for example
// postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable. Each test gets its own
// freshly created database, dropped when the test ends, so no existing data is touched.
package testdb

import (
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

const envDSN = "PHARMACY_TEST_DSN"

var migrationFile = regexp.MustCompile(`^(\d+)_.+\.(up|down)\.sql$`)

// New creates a database and applies the migrations up to and including version upTo, or all
// of them when upTo is 0.
func New(t *testing.T, upTo int) *sqlx.DB {
	t.Helper()
	dsn := os.Getenv(envDSN)
	if dsn == "" {
		t.Skipf("set %s to run database integration tests", envDSN)
	}

	admin, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		t.Fatalf("connect to %s: %v", envDSN, err)
	}
	name := fmt.Sprintf("pharmacy_test_%d_%d", time.Now().UnixNano(), rand.Intn(1000))
	if _, err = admin.Exec("CREATE DATABASE " + name); err != nil {
		admin.Close()
		t.Fatalf("create test database: %v", err)
	}

	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("parse %s: %v", envDSN, err)
	}
	u.Path = "/" + name
	db, err := sqlx.Connect("pgx", u.String())
	if err != nil {
		t.Fatalf("connect to test database: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
		if _, err := admin.Exec("DROP DATABASE IF EXISTS " + name + " WITH (FORCE)"); err != nil {
			t.Logf("drop test database %s: %v", name, err)
		}
		admin.Close()
	})

	Migrate(t, db, 0, upTo)
	return db
}

// Migrate applies the up migrations with versions in (after, upTo]; upTo 0 means no upper
// limit.
func Migrate(t *testing.T, db *sqlx.DB, after, upTo int) {
	t.Helper()
	for _, m := range list(t, "up") {
		if m.version <= after || (upTo != 0 && m.version > upTo) {
			continue
		}
		run(t, db, m)
	}
}

// Rollback applies the down migration of one version.
func Rollback(t *testing.T, db *sqlx.DB, version int) {
	t.Helper()
	for _, m := range list(t, "down") {
		if m.version == version {
			run(t, db, m)
			return
		}
	}
	t.Fatalf("no down migration for version %d", version)
}

type migration struct {
	version int
	path    string
}

func list(t *testing.T, direction string) []migration {
	t.Helper()
	_, thisFile, _, _ := runtime.Caller(0)
	dir := filepath.Join(filepath.Dir(thisFile), "..", "..", "db", "migrations")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read migrations: %v", err)
	}
	var found []migration
	for _, entry := range entries {
		match := migrationFile.FindStringSubmatch(entry.Name())
		if match == nil || match[2] != direction {
			continue
		}
		version, _ := strconv.Atoi(match[1])
		found = append(found, migration{version: version, path: filepath.Join(dir, entry.Name())})
	}
	sort.Slice(found, func(i, j int) bool { return found[i].version < found[j].version })
	return found
}

func run(t *testing.T, db *sqlx.DB, m migration) {
	t.Helper()
	body, err := os.ReadFile(m.path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(string(body)); err != nil {
		t.Fatalf("migration %s: %v", filepath.Base(m.path), err)
	}
}
