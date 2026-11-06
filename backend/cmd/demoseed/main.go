// Command demoseed fills an empty, migrated database on this machine with made-up data, for
// running the app to take screenshots. See internal/demoseed.
//
//	go run ./cmd/demoseed -throwaway -dsn "postgres://user:pass@localhost:5544/db?sslmode=disable"
package main

import (
	"flag"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"pharmacy/internal/demoseed"
)

func main() {
	dsn := flag.String("dsn", "", "connection string of an empty, migrated database on this machine")
	throwaway := flag.Bool("throwaway", false,
		"confirm the database is disposable (a tunnel to a real database also looks local)")
	flag.Parse()
	if !*throwaway {
		fmt.Fprintln(os.Stderr, "demoseed writes fake users, products and sales: "+
			"pass -throwaway to confirm this database is disposable")
		os.Exit(2)
	}
	if *dsn == "" {
		fmt.Fprintln(os.Stderr,
			"usage: demoseed -dsn postgres://user:pass@localhost:PORT/db?sslmode=disable")
		os.Exit(2)
	}
	if err := demoseed.CheckTarget(*dsn); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	db, err := sqlx.Connect("pgx", *dsn)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer db.Close()
	s, err := demoseed.Run(db)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("Seeded %d products and %d sales.\nSign in as:\n", s.Products, s.Sales)
	for _, u := range demoseed.Users {
		fmt.Printf("  %-6s (%s)  password %s\n", u.Name, u.Role, demoseed.Password)
	}
}
