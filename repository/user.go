package repository

import (
	"context"
	"database/sql"
	"pharmacy/model"
)

func (r *repo) FetchUserWithPassword(ctx context.Context, username string) (model.User, error) {
	var user model.User
	err := r.Data.GetContext(ctx, &user, fetchUserByNameQuery, username)
	if err != nil {
		return model.User{}, err
	}
	
	return user, nil
}

func (r *repo) CheckUserNameExists(ctx context.Context, userName string) (bool, error) {
	var dummy int
	err := r.Data.GetContext(ctx, &dummy, usernameExistsQuery, userName)
	if err == sql.ErrNoRows {
		return false, nil
	}
	
	if err != nil {
		return false, err
	}
	
	return true, nil
}

func (r *repo) CreateUserAccount(ctx context.Context, user model.User) error {
	_, err := r.Data.ExecContext(ctx, createUserQuery, user.UserName, user.Password)
	return err
}