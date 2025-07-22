package service

import (
	"context"
	"pharmacy/model"
)

type UserService interface {
	CreateUserAccount(ctx context.Context, user model.User) error
}