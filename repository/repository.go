package repository

import (
	"context"
	"pharmacy/model"
)

type UserRepository interface {
	FetchUserWithPassword(ctx context.Context, userName string) (model.User, error)
	CheckUserNameExists(ctx context.Context, userName string) (bool, error)
	CreateUserAccount(ctx context.Context, user model.User) error
}

type PharmacyRepository interface {
	UserRepository
}