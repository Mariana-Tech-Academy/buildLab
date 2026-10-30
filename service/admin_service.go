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
	AdminUpdateOrderStatus(ctx context.Context, adminID, orderID uint, newStatus models.OrderStatus) (*models.Order, error)
	ListUsers(ctx context.Context, search string, page, limit int) (*AdminUsersPage, error)
	GetUserDetails(ctx context.Context, id uint) (*models.User, error)
}

// AdminUsersPage is one page of users for an admin.
type AdminUsersPage struct {
	Users []models.User
	Page  int
	Limit int
	Total int64
}

// AdminOrdersPage is one page of orders for the admin.
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
	status = strings.TrimSpace(status)
	search = strings.TrimSpace(search)
	if !isValidOrderStatus(status) {
		return nil, ErrInvalidOrderStatus
	}

	page, limit, offset, err := orderPageBounds(page, limit)
	if err != nil {
		return nil, err
	}

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

// AdminUpdateOrderStatus checks the status change is allowed, then changes it and records which admin did it.
// It reuses the team's existing rules: nextOrderStatus from ORD-05 for moving forward,
// and CanCancel from ORD-06 for cancelling.
func (s *adminService) AdminUpdateOrderStatus(ctx context.Context, adminID, orderID uint, newStatus models.OrderStatus) (*models.Order, error) {
	newStatus = models.OrderStatus(strings.TrimSpace(string(newStatus)))
	if !knownOrderStatus(newStatus) {
		return nil, ErrInvalidOrderStatus
	}

	order, err := s.adminRepo.GetOrderForAdmin(ctx, orderID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}

	// Delivered and cancelled are final, so they can't be changed.
	if order.Status == models.OrderStatusDelivered || order.Status == models.OrderStatusCancelled {
		return nil, ErrOrderStatusFinal
	}

	if newStatus == models.OrderStatusCancelled {
		// Cancelling follows the team's cancellation rule (Pending or Confirmed only).
		if !CanCancel(order.Status) {
			return nil, ErrOrderCannotBeCancelled
		}
	} else {
		// Anything else must be the next step, so no stages can be skipped.
		expected, ok := nextOrderStatus(order.Status)
		if !ok || newStatus != expected {
			if ok {
				return nil, fmt.Errorf("%w: allowed next status is %s", ErrInvalidStatusTransition, expected)
			}
			return nil, ErrInvalidStatusTransition
		}
	}

	err = s.adminRepo.AdminUpdateOrderStatus(ctx, order.ID, order.Status, newStatus, adminID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Someone else changed the order just before us.
		return nil, fmt.Errorf("%w: order status has changed, please refresh and try again", ErrInvalidStatusTransition)
	}
	if err != nil {
		return nil, err
	}

	order.Status = newStatus
	return order, nil
}

// ListUsers checks the page, then gets one page of users matching a name or email.
func (s *adminService) ListUsers(ctx context.Context, search string, page, limit int) (*AdminUsersPage, error) {
	search = strings.TrimSpace(search)
	page, limit, offset, err := orderPageBounds(page, limit)
	if err != nil {
		return nil, err
	}

	users, total, err := s.adminRepo.ListUsers(ctx, search, limit, offset)
	if err != nil {
		return nil, err
	}
	return &AdminUsersPage{Users: users, Page: page, Limit: limit, Total: total}, nil
}

// GetUserDetails gets one user, and returns ErrUserNotFound if that user does not exist.
func (s *adminService) GetUserDetails(ctx context.Context, id uint) (*models.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}
