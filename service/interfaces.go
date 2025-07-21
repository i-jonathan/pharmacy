package service

import "pharmacy/model"

type UserService interface {
	CreateUserAccount(user model.User) error
}