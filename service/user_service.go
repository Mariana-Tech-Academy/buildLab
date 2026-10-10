package service

import (
	"Fhub/auth"
	"Fhub/repository"
	"Fhub/models"

)

type UserService interface {
	GetUserByID(id int) (models.User, error)
}

type userService struct {
	userRepo repository.UserRepository
	tokens   *auth.JWTManager
}

func NewUserService(repo repository.UserRepository, tokens *auth.JWTManager) UserService {
	return &userService{userRepo: repo, tokens: tokens}
}

func (s *userService) GetUserByID(id int) (models.User, error) {
	return s.userRepo.GetUserByID(id)
}