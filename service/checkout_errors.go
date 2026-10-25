package service

import "errors"

// Errors returned when checkout cannot be completed.
var (
	ErrCartEmpty = errors.New("your cart is empty")

	ErrItemUnavailable = errors.New("one or more items are unavailable")

	ErrInvalidDelivery = errors.New("invalid delivery information")

	ErrCartRestaurantMismatch = errors.New("cart items do not belong to the selected restaurant")
)
