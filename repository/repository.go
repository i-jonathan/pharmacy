package repository

import "pharmacy/model"

type UserRepository interface {
	FetchUserWithPassword(userName string) (model.User, error)
	CreateUserAccount(user model.User) error
}

type PharmacyRepository interface {
	UserRepository
}