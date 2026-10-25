package service

import (
	"context"
	"errors"

	"Fhub/models"
	"Fhub/repository"
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

// PlaceOrder validates the checkout details and creates the order.
func (s *checkoutService) PlaceOrder(
	ctx context.Context,
	customerID uint,
	request CheckoutRequest,
) (*models.Order, error) {

	// Check that the customer ID is valid.
	if customerID == 0 {
		return nil, errors.New("invalid customer ID")
	}

	// Check that the delivery details are complete.
	if err := request.Validate(); err != nil {
		return nil, err
	}

	// Retrieve the customer's cart.
	checkoutCart, err := s.checkoutRepo.GetCheckoutCart(ctx, customerID)
	if err != nil {
		return nil, err
	}

	// Check that the cart contains valid items.
	if err := ValidateCheckoutCart(checkoutCart); err != nil {
		return nil, err
	}

	// Generate an order number using the team's format.
	orderNumber, err := GenerateCheckoutOrderNumber()
	if err != nil {
		return nil, err
	}

	// Use a temporary delivery fee until the team confirms the calculation.
	deliveryFee := int64(0)

	// Pass the checkout details to the repository.
	input := repository.CheckoutInput{
		CustomerID:           customerID,
		DeliveryName:         request.DeliveryName,
		DeliveryPhone:        request.DeliveryPhone,
		DeliveryAddress:      request.DeliveryAddress,
		DeliveryCity:         request.DeliveryCity,
		DeliveryState:        request.DeliveryState,
		DeliveryInstructions: request.DeliveryInstructions,
		OrderNumber:          orderNumber,
		DeliveryFee:          deliveryFee,
	}

	// Create the order and clear the cart in one transaction.
	return s.checkoutRepo.CreateCheckoutOrder(ctx, input)
}