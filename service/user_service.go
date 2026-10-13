package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"Fhub/auth"
	"Fhub/models"
	"Fhub/repository"
	"Fhub/utils"

	"gorm.io/gorm"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrValidation   = errors.New("validation failed")
	ErrIncorrectPassword = errors.New("incorrect current password")
)

type UserService interface {
	GetUserByID(ctx context.Context, id uint) (*models.User, error)
	UpdateProfile(ctx context.Context, id uint, firstName, lastName, phone string) (*models.User, error)
	ChangePassword(ctx context.Context, id uint, currentPassword, newPassword string) error
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

func (s *userService) ChangePassword(ctx context.Context, id uint, currentPassword, newPassword string) error {
    if currentPassword == "" || len(newPassword) < 8 {
        return ErrValidation
    }

    user, err := s.GetUserByID(ctx, id)
    if err != nil {
        return err
    }

    if !utils.CheckPassword(user.Password, currentPassword) {
        return ErrIncorrectPassword
    }

    newHash, err := utils.HashPassword(newPassword)
    if err != nil {
        return err
    }

    return s.userRepo.UpdatePassword(ctx, id, newHash)
}

