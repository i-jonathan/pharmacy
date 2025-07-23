package service

import (
	"context"
	"log"
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
		log.Println(err)
		return httperror.BadRequest("invalid username", err)
	}
	
	usernameExists, err := s.repo.CheckUserNameExists(ctx, user.UserName)
	if err != nil || usernameExists {
		log.Println(err)
		return httperror.BadRequest("username already exists", err)
	}
	
	err = user.ValidatePassword()
	if err != nil {
		log.Println(err)
		return httperror.BadRequest("invalid password", err)
	}
	
	user.HashPassword()
	err = s.repo.CreateUserAccount(ctx, user)
	if err != nil {
		log.Println(err)
	}
	return nil
}

func (s *userService) AuthenticateUser(ctx context.Context, user *model.User) error {
	storedUser, err := s.repo.FetchUserWithPassword(ctx, user.UserName)
	if err != nil {
		log.Println(err)
		return httperror.Unauthorized("invalid username or password", err)
	}
	
	correct, err := user.VerifyPassword(storedUser.Password)
	if err != nil {
		log.Println(err)
		return httperror.Unauthorized("invalid username or password", err)
	}
	
	if !correct {
		log.Println(err)
		return httperror.Unauthorized("invalid username or password", err)
	}
	return nil
}