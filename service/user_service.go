package service

import (
	"Fhub/auth"
	"Fhub/models"
	"Fhub/repository"
	"Fhub/utils"
	"errors"
	"strings"
)

var ErrEmailTaken = errors.New("email is already taken")

type RegisterInput struct {
	FirstName string
	LastName  string
	Email     string
	Phone     string
	Password  string
}

type UserService interface {
	Register(input RegisterInput) (*models.User, error)
}

type userService struct {
	userRepo repository.UserRepository
	tokens   *auth.JWTManager
}

func NewUserService(repo repository.UserRepository, tokens *auth.JWTManager) UserService {
	return &userService{userRepo: repo, tokens: tokens}
}

func (s *userService) Register(input RegisterInput) (*models.User, error) {
	email := strings.ToLower(input.Email)

	exists, err := s.userRepo.EmailExists(email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailTaken
	}

	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     email,
		Phone:     input.Phone,
		Password:  hashedPassword,
		Role:      models.RoleCustomer,
		Status:    models.UserStatusActive,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}
