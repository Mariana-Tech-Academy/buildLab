package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"Fhub/auth"
	"Fhub/models"
	"Fhub/repository"
	"Fhub/utils"

	"gorm.io/gorm"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrValidation   = errors.New("validation failed")
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountSuspended   = errors.New("your account is suspeneded")
)

type LoginUser struct {
	ID        uint        `json:"id"`
	FirstName string      `json:"first_name"`
	LastName  string      `json:"last_name"`
	Email     string      `json:"email"`
	Role      models.Role `json:"role"`
}

type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      LoginUser `json:"user"`
}

type UserService interface {
	GetUserByID(ctx context.Context, id uint) (*models.User, error)
	UpdateProfile(ctx context.Context, id uint, firstName, lastName, phone string) (*models.User, error)
	Login(email, password string) (*LoginResponse, error)
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

func (s *userService) Login(email, password string) (*LoginResponse, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if !utils.CheckPassword(user.Password, password) {
		return nil, ErrInvalidCredentials

	}
	if user.Status == models.UserStatusSuspended {
		return nil, ErrAccountSuspended
	}

	token, err := s.tokens.Generate(user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	claims, err := s.tokens.Verify(token)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token:     token,
		ExpiresAt: claims.ExpiresAt.Time,
		User: LoginUser{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
			Role:      user.Role,
		},
	}, nil
}
