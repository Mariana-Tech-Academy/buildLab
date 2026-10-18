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
	Checkout(ctx context.Context, userID uint, order *models.Order) error
}

type orderService struct {
	orderRepo repository.OrderRepository
}

func NewOrderService(repo repository.OrderRepository) OrderService {
	return &orderService{orderRepo: repo}
}

func (s *orderService) CanAcceptOrder(
	ctx context.Context,
	restaurantID uint,
) error {
	active, err := s.orderRepo.IsRestaurantActive(ctx, restaurantID)
	if err != nil {
		return err
	}
	if !active {
		return ErrRestaurantNotActive
	}
	return nil
}

func (s *orderService) Checkout(
	ctx context.Context,
	userID uint,
	order *models.Order,
) error {
	if err := s.CanAcceptOrder(ctx, order.RestaurantID); err != nil {
		return err
	}
	return s.orderRepo.Checkout(ctx, userID, order)
}