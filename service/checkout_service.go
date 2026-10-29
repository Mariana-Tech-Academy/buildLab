package service

import (
	"context"
	"errors"

	"Fhub/models"
	"Fhub/repository"

	"gorm.io/gorm"
)

// CheckoutService defines the checkout operation.
type CheckoutService interface {
	PlaceOrder(ctx context.Context, customerID uint, request DeliveryInfoInput) (*models.Order, error)
}

type checkoutService struct {
	checkoutRepo repository.CheckoutRepository
}

// NewCheckoutService creates a checkout service.
func NewCheckoutService(repo repository.CheckoutRepository) CheckoutService {
	return &checkoutService{checkoutRepo: repo}
}

// PlaceOrder validates the checkout details and creates the order.
func (s *checkoutService) PlaceOrder(ctx context.Context, customerID uint, request DeliveryInfoInput) (*models.Order, error) {
	if customerID == 0 {
		return nil, errors.New("invalid customer ID")
	}
	if errs := (&orderService{}).ValidateDeliveryInfo(&request); len(errs) > 0 {
		return nil, ErrInvalidDelivery
	}

	checkoutCart, err := s.checkoutRepo.GetCheckoutCart(ctx, customerID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrEmptyCart
	}
	if err != nil {
		return nil, err
	}
	if err := ValidateCheckoutCart(checkoutCart); err != nil {
		return nil, err
	}

	order, err := s.checkoutRepo.CreateCheckoutOrder(ctx, repository.CheckoutInput{
		CustomerID:           customerID,
		DeliveryName:         request.RecipientName,
		DeliveryPhone:        request.Phone,
		DeliveryAddress:      request.Address,
		DeliveryCity:         request.City,
		DeliveryState:        request.State,
		DeliveryInstructions: request.DeliveryInstructions,
		DeliveryFee:          repository.DeliveryFee,
		Subtotal:             CalculateCheckoutSubtotal,
	})
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrEmptyCart):
			return nil, ErrEmptyCart
		case errors.Is(err, repository.ErrUnavailableItem):
			return nil, ErrUnavailableItem
		case errors.Is(err, repository.ErrRestaurantInactive), errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrRestaurantNotFound
		default:
			return nil, err
		}
	}
	return order, nil
}
