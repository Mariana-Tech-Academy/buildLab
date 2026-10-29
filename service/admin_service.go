package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"Fhub/models"
	"Fhub/repository"

	"gorm.io/gorm"
)

// ErrInvalidRole means the role is not customer, manager, or admin.
var ErrInvalidRole = errors.New("invalid role")

type AdminService interface {
	SetUserRole(ctx context.Context, targetID uint, role models.Role) (*models.User, error)
	ListAllOrders(ctx context.Context, status, search string, page, limit int) (*AdminOrdersPage, error)
	GetOrderForAdmin(ctx context.Context, orderID uint) (*models.Order, error)
}

// AdminOrderspage is one page of orders for the admin
type AdminOrdersPage struct {
	Orders []models.Order
	Page   int
	Limit  int
	Total  int64
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

// ListAllOrders checks the status filter and page, then gets one page of orders.
func (s *adminService) ListAllOrders(ctx context.Context, status, search string, page, limit int) (*AdminOrdersPage, error) {
	if !isValidOrderStatus(status) {
		return nil, ErrInvalidOrderStatus
	}

	page, limit, offset, err := orderPageBounds(page, limit)
	if err != nil {
		return nil, err
	}

	search = strings.TrimSpace(search)

	orders, total, err := s.adminRepo.ListAllOrders(ctx, status, search, limit, offset)
	if err != nil {
		return nil, err
	}

	return &AdminOrdersPage{Orders: orders, Page: page, Limit: limit, Total: total}, nil
}

// GetOrderForAdmin gets one order, and returns ErrOrderNotFound if it doesn't exist.
func (s *adminService) GetOrderForAdmin(ctx context.Context, orderID uint) (*models.Order, error) {
	order, err := s.adminRepo.GetOrderForAdmin(ctx, orderID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}

	return order, nil
}
