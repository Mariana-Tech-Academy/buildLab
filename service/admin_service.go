package service

import (
	"context"
	"errors"
	"fmt"

	"Fhub/models"
	"Fhub/repository"

	"gorm.io/gorm"
)

// ErrInvalidRole means the role is not customer, manager, or admin.
var ErrInvalidRole = errors.New("invalid role")

// Stores the list of users and the pagination information.
type AdminUsersResult struct {
	Users []models.User
	Page  int
	Limit int
	Total int64
}

type AdminService interface {
	SetUserRole(ctx context.Context, targetID uint, role models.Role) (*models.User, error)
	ListUsers(ctx context.Context, search string, page, limit int) (*AdminUsersResult, error)
	GetUserDetails(ctx context.Context, id uint) (*models.User, error)
}

type adminService struct {
	adminRepo repository.AdminRepository
	userRepo  repository.UserRepository
}

func NewAdminService(adminRepo repository.AdminRepository, userRepo repository.UserRepository) AdminService {
	return &adminService{adminRepo: adminRepo, userRepo: userRepo}
}

// SetUserRole lets an admin assign customer, manager, or admin.
func (s *adminService) SetUserRole(ctx context.Context, targetID uint, role models.Role) (*models.User, error) {
	switch role {
	case models.RoleCustomer, models.RoleManager, models.RoleAdmin:
	default:
		return nil, ErrInvalidRole
	}

	if err := s.userRepo.SetRole(ctx, targetID, role); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("set role: %w", err)
	}

	user, err := s.userRepo.GetUserByID(ctx, targetID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	return user, err
}

// Gets a list of users with searching and pagination.
func (s *adminService) ListUsers(

	ctx context.Context,
	search string,
	page, limit int,

) (*AdminUsersResult, error) {

	// Checks that the page number is valid.
	if page < 1 {
		return nil, fmt.Errorf("%w: page must be greater than 0", ErrValidation)
	}
	if limit < 1 || limit > 50 {
		return nil, fmt.Errorf("%w: limit must be between 1 and 50", ErrValidation)
	}
	// Calculates where the results should start.
	maxInt := int(^uint(0) >> 1)
	if page-1 > maxInt/limit {
		return nil, fmt.Errorf("%w: page number is too large", ErrValidation)
	}
	offset := (page - 1) * limit

	// Gets the matching users and total count from the database.
	users, total, err := s.adminRepo.ListUsers(
		ctx,
		search,
		limit,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	// Returns the users with their pagination information.
	return &AdminUsersResult{
		Users: users,
		Page:  page,
		Limit: limit,
		Total: total,
	}, nil

}

// Gets the details of a user using their ID.
func (s *adminService) GetUserDetails(
	ctx context.Context,
	id uint,
) (*models.User, error) {

	user, err := s.userRepo.GetUserByID(ctx, id)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("get user details: %w", err)
	}
	return user, nil
}
