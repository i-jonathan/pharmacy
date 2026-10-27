package service

import (
	"context"
	"log"
	"pharmacy/httperror"
	"pharmacy/model"
	"pharmacy/repository"
)

type userService struct {
	repo repository.PharmacyRepository
}

func NewUserService(repo repository.PharmacyRepository) *userService {
	return &userService{repo: repo}
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
	user.ID = storedUser.ID
	user.RoleID = storedUser.RoleID
	user.RoleName = storedUser.RoleName
	user.Password = ""
	user.Permissions = storedUser.Permissions
	return nil
}
