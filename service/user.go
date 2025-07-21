package service

import (
	"pharmacy/model"
	"pharmacy/repository"
)

type userService struct {
	repo repository.PharmacyRepository
}

func NewUserService(repo repository.PharmacyRepository) *userService {
	return &userService{repo: repo}
}

func (s *userService) CreateUserAccount(user model.User) error {
	user.HashPassword()
	return s.repo.CreateUserAccount(user)
}