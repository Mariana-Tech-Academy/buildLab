package service

import (
	"context"
	"errors"

	"Fhub/repository"
)

var ErrRestaurantNotActive = errors.New("restaurant is not active")

type OrderService interface {
	CanAcceptOrder(ctx context.Context, restaurantID uint) error
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
