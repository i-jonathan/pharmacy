package repository

import (
	"fmt"
	"pharmacy/config"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type repo struct {
	Data *sqlx.DB
}

func InitStore() (*repo, error) {
	c := config.GetConfig()
	
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