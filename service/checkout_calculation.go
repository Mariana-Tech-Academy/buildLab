package service

import (
	"errors"
	"math"

	"Fhub/models"
)

// ErrInvalidPrice is returned when an item has an invalid price.
var ErrInvalidPrice = errors.New("invalid item price or quantity")

// CalculateCheckoutSubtotal calculates the total cost of cart items.
func CalculateCheckoutSubtotal(
	items []models.CartItem,
	prices map[uint]int64,
) (int64, error) {

	var subtotal int64

	// Calculate the cost of each item in the cart.
	for _, item := range items {
		// Find the current price of the menu item.
		price, exists := prices[item.MenuItemID]

		// Reject missing prices, negative prices or invalid quantities.
		if !exists || price < 0 || item.Quantity <= 0 {
			return 0, ErrInvalidPrice
		}
		// Prevent integer overflow when multiplying price by quantity.
		if price > 0 && int64(item.Quantity) > math.MaxInt64/price {
			return 0, ErrInvalidPrice
		}
		lineTotal := price * int64(item.Quantity)

		// Prevent integer overflow when adding to the subtotal.
		if subtotal > math.MaxInt64-lineTotal {
			return 0, ErrInvalidPrice
		}
		subtotal += lineTotal
	}
	return subtotal, nil
}