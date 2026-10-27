package repository

import (
	"context"
	"errors"
	"pharmacy/model"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

func (r *repo) FetchUserWithPassword(ctx context.Context, username string) (model.User, error) {
	var user model.User
	err := r.Data.GetContext(ctx, &user, fetchUserByNameQuery, username)
	if err != nil {
		return model.User{}, err
	}

	return user, nil
}

func (r *repo) BulkFetchUserByIDTx(ctx context.Context, tx *sqlx.Tx, userIDs []int) ([]model.User, error) {
	var users []model.User

	err := tx.SelectContext(ctx, &users, bulkFetchUserByIDQuery, pq.Array(userIDs))
	if err != nil {
		return nil, err
	}
	return users, nil
}

// ErrUsernameTaken is returned when a new user's username is already in use.
var ErrUsernameTaken = errors.New("username already exists")

// CreateUserTx inserts a user with its role and returns the new id. The unique index on
// username decides duplicates, so two admins adding the same name at once can't both succeed.
func (r *repo) CreateUserTx(ctx context.Context, tx *sqlx.Tx, user model.User) (int, error) {
	var id int
	err := tx.GetContext(
		ctx, &id, createUserWithRoleQuery, user.UserName, user.Password, user.RoleID,
	)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return 0, ErrUsernameTaken
	}
	return id, err
}

func (r *repo) RoleExistsTx(ctx context.Context, tx *sqlx.Tx, roleID int) (bool, error) {
	var exists bool
	err := tx.GetContext(ctx, &exists, roleExistsQuery, roleID)
	return exists, err
}

func (r *repo) FetchRoleIDByName(ctx context.Context, name string) (int, error) {
	var id int
	err := r.Data.GetContext(ctx, &id, roleIDByNameQuery, name)
	return id, err
}
