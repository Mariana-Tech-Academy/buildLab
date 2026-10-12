package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"Fhub/auth"
	"Fhub/models"
	"Fhub/repository"

	"gorm.io/gorm"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrValidation   = errors.New("validation failed")
)

type UserService interface {
	GetUserByID(ctx context.Context, id uint) (*models.User, error)
	UpdateProfile(ctx context.Context, id uint, firstName, lastName, phone string) (*models.User, error)
}

type userService struct {
	userRepo repository.UserRepository
	tokens   *auth.JWTManager
}

func NewUserService(repo repository.UserRepository, tokens *auth.JWTManager) UserService {
	return &userService{userRepo: repo, tokens: tokens}
}

func (s *userService) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	return user, err
}

func (s *userService) UpdateProfile(ctx context.Context, id uint, firstName, lastName, phone string) (*models.User, error) {
	firstName = strings.TrimSpace(firstName)
	lastName = strings.TrimSpace(lastName)
	phone = strings.TrimSpace(phone)

	if firstName == "" || lastName == "" || phone == "" {
		return nil, fmt.Errorf("%w: first name, last name and phone are required", ErrValidation)
	}

	user, err := s.userRepo.UpdateProfile(ctx, id, firstName, lastName, phone)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	return user, err
}
