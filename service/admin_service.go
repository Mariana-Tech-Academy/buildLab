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

type AdminService interface {
	SetUserRole(ctx context.Context, targetID uint, role models.Role) (*models.User, error)
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
