package repository

import (
	"context"
	"fmt"
	"pharmacy/config"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type repo struct {
	Data *sqlx.DB
}

func InitStore() (*repo, error) {
	c := config.Conf
	dsn := fmt.Sprintf(
        "postgres://%s:%s@%s:%s/%s?sslmode=disable",
        c.DBUser, c.DBPass, c.DBHost, c.DBPort, c.DBName,
	)
	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		return nil, err
	}
	
	dataRepo := new(repo)
	dataRepo.Data = db
	return dataRepo, nil
}

func (r *repo) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return r.Data.BeginTxx(ctx, nil)
}

func (r *repo) CommitTx(tx *sqlx.Tx) error {
	return tx.Commit()
}

func (r *repo) Rollback(tx *sqlx.Tx) {
	_ = tx.Rollback()
}