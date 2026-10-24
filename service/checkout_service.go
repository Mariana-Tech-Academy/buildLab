package service

import (
	"Fhub/models"
	"Fhub/repository"
	"context"
	"errors"
)

// CheckoutService defines the checkout operation.
type CheckoutService interface {
	PlaceOrder(
		ctx context.Context,
		customerID uint,
		request CheckoutRequest,
	) (*models.Order, error)
}
// checkoutService holds the repository used for checkout.
type checkoutService struct {
	checkoutRepo repository.CheckoutRepository
}
// NewCheckoutService creates a checkout service.
func NewCheckoutService(
	repo repository.CheckoutRepository,
) CheckoutService {
	return &checkoutService{
		checkoutRepo: repo,
	}
}
// PlaceOrder will coordinate checkout and order creation.
func (s *checkoutService) PlaceOrder(

	ctx context.Context,
	customerID uint,
	request CheckoutRequest,

) (*models.Order, error) {
	return nil, errors.New("checkout is not implemented yet")
}
