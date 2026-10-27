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

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrValidation         = errors.New("validation failed")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountSuspended   = errors.New("your account is suspeneded")
	ErrEmailTaken         = errors.New("email is already taken")
	ErrIncorrectPassword  = errors.New("incorrect current password")
)

type RegisterInput struct {
	FirstName string
	LastName  string
	Email     string
	Phone     string
	Password  string
}

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
	Login(ctx context.Context, email, password string) (*LoginResponse, error)
	Register(ctx context.Context, input RegisterInput) (*models.User, error)
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

func (s *userService) Login(ctx context.Context, email, password string) (*LoginResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
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

func (s *userService) Register(ctx context.Context, input RegisterInput) (*models.User, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))

	exists, err := s.userRepo.EmailExists(ctx, email)
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

	if err := s.userRepo.CreateWithCart(ctx, user); err != nil {
		if isUniqueViolation(err) {
			return nil, ErrEmailTaken
		}
		return nil, err
	}

	return user, nil
}

func isUniqueViolation(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func (s *userService) ChangePassword(ctx context.Context, id uint, currentPassword, newPassword string) error {
	if currentPassword == "" {
		return fmt.Errorf("%w: current password is required", ErrValidation)
	}
	if !utils.IsStrongPassword(newPassword) {
		return fmt.Errorf("%w: password must be a minimum of 8 characters and include an uppercase letter, a lowercase letter and a number", ErrValidation)
	}
	if currentPassword == newPassword {
		return fmt.Errorf("%w: new password must be different from the current password", ErrValidation)
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

	if err := s.userRepo.UpdatePassword(ctx, id, newHash); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	return nil
}
