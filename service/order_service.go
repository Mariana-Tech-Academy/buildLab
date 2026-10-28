package service

import (
	"context"
	"errors"

	"Fhub/models"
	"Fhub/repository"
)

var ErrRestaurantNotActive = errors.New("restaurant is not active")

type OrderService interface {
	CanAcceptOrder(ctx context.Context, restaurantID uint) error
	GetOrderByID(ctx context.Context, orderID uint, customerID uint) (*models.Order, error)
}

type orderService struct {
	orderRepo repository.OrderRepository
}

func NewOrderService(repo repository.OrderRepository) OrderService {
	return &orderService{orderRepo: repo}
}

// CanAcceptOrder is used by checkout (when implemented) to reject orders
// for restaurants that are not active.
func (s *orderService) CanAcceptOrder(ctx context.Context, restaurantID uint) error {
	active, err := s.orderRepo.IsRestaurantActive(ctx, restaurantID)
	if err != nil {
		return err
	}
	if !active {
		return ErrRestaurantNotActive
	}
	return nil
}

// GetOrderByID retrieves an order using the order ID and customer ID.
// Returns the order details if found, or an error if no matching order exists.
func (s *orderService) GetOrderByID(ctx context.Context, orderID uint, customerID uint) (*models.Order, error) {
	return s.orderRepo.GetOrderByID(ctx, orderID, customerID)

}
