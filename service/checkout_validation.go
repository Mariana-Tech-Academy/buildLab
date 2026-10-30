package service

import "Fhub/repository"

// ValidateCheckoutCart checks that the cart contains items.
func ValidateCheckoutCart(cart *repository.CheckoutCart) error {

	// Reject checkout if the cart doesn't exist.
	if cart == nil {
		return ErrEmptyCart
	}
	// Reject checkout if there are no items in the cart.
	if len(cart.Items) == 0 {
		return ErrEmptyCart
	}
	// The cart contains items, so validation passes.
	return nil
}
