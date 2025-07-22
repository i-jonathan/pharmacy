package service

import (
	"context"
	"pharmacy/httperror"
	"pharmacy/model"
	"pharmacy/repository"
	"strings"
)

type userService struct {
	repo repository.PharmacyRepository
}

func NewUserService(repo repository.PharmacyRepository) *userService {
	return &userService{repo: repo}
}

func (s *userService) CreateUserAccount(ctx context.Context, user model.User) error {
	user.UserName = strings.TrimSpace(user.UserName)
	
	err := user.ValidateUsername()
	if err != nil {
		return httperror.BadRequest("invalid username", err)
	}
	
	usernameExists, err := s.repo.CheckUserNameExists(ctx, user.UserName)
	if err != nil || usernameExists {
		return httperror.BadRequest("username already exists", err)
	}
	
	err = user.ValidatePassword()
	if err != nil {
		return httperror.BadRequest("invalid password", err)
	}
	
	user.HashPassword()
	return s.repo.CreateUserAccount(ctx, user)
}