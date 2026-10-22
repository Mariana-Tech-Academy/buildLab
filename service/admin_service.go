package service

import (
	"Fhub/models"
	"Fhub/repository"
	"context"
	"errors"
	"fmt"

	
)


var (     
	ErrInvalidRole = errors.New("role must be customer, manager or admin")   // 422)
	ErrUserNotFound = errors.New("user not found") // 404
)


type AdminService interface {
    SetUserRole(ctx context.Context, targetID uint, role models.Role) (*models.User, error)
	
}

type adminService struct {
	adminRepo repository.AdminRepository
	userRepo repository.UserRepository
}

func NewAdminService(adminRepo repository.AdminRepository, userRepo repository.UserRepository) AdminService {
	return &adminService{adminRepo: adminRepo, userRepo: userRepo}
}





func (s *adminService) SetUserRole(ctx context.Context, targetID uint, role models.Role) (*models.User, error) {
    // Validate role explicitly (avoid accepting arbitrary values)
    switch role {
    case models.RoleCustomer, models.RoleManager, models.RoleAdmin:
    default:
        return nil, ErrInvalidRole
    }

    // Pass context down to the repo so cancellations and deadlines work.
    if err := s.userRepo.SetRole(ctx, targetID, role); err != nil {
        // Map repo-level not-found to service-level not-found
        if errors.Is(err, ErrUserNotFound) {
            return nil, ErrUserNotFound
        }
        return nil, fmt.Errorf("set role: %w", err)
    }

    // Return the updated user (repo should fetch fresh data)
    return s.userRepo.GetUserByID(ctx, targetID)
}

