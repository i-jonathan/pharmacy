package repository

import "pharmacy/model"

func (r *repo) FetchUserWithPassword(username string) (model.User, error) {
	// todo implement
	return model.User{}, nil
}

func (r *repo) CreateUserAccount(user model.User) error {
	// todo implement
	return nil
}