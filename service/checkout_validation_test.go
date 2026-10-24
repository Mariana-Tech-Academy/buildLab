package service

import (
	"Fhub/models"
	"Fhub/repository"
	"errors"
	"testing"
)

// Test that a cart containing items is accepted.
func TestValidateCheckoutCartValid(t *testing.T) {

	// Create a sample cart containing 1 item.
	cart := &repository.CheckoutCart{
		Items: []models.CartItem{
			{Quantity: 2},
		},
	}
	// A cart with items should not return an error.
	if err := ValidateCheckoutCart(cart); err != nil {
		t.Errorf("Expected valid cart, got: %v", err)
	}
}

// Test that an empty cart is rejected.
func TestValidateCheckoutCartEmpty(t *testing.T) {

	// Create a cart with 0 items.
	cart := &repository.CheckoutCart{
		Items: []models.CartItem{},
	}

	// Check that the correct error is returned.
	if err := ValidateCheckoutCart(cart); !errors.Is(err, ErrCartEmpty) {
		t.Errorf("Expected ErrCartEmpty, got: %v", err)
	}
}

// Test that a missing cart is rejected.
func TestValidateCheckoutCartMissing(t *testing.T) {

	// A nil cart represents a missing cart.
	var cart *repository.CheckoutCart

	// Check that the correct error is returned.
	if err := ValidateCheckoutCart(cart); !errors.Is(err, ErrCartEmpty) {
		t.Errorf("Expected ErrCartEmpty, got: %v", err)
	}
}
