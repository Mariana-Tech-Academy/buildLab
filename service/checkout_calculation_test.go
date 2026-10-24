package service

import (
	"errors"
	"math"
	"testing"

	"Fhub/models"
)

// Test checkout subtotal calculations using different scenarios.
func TestCalculateCheckoutSubtotal(t *testing.T) {
	tests := []struct {
		name     string
		items    []models.CartItem
		prices   map[uint]int64
		expected int64
		wantErr  bool
	}{
		{
			name: "Valid order with two items",
			items: []models.CartItem{
				{MenuItemID: 1, Quantity: 2},
				{MenuItemID: 2, Quantity: 1},
			},
			prices: map[uint]int64{
				1: 850, // Burger costs £8.50
				2: 300, // Chips cost £3.00
			},
			expected: 2000, // Total is £20.00
		},
		{
			name: "Missing item price",
			items: []models.CartItem{
				{MenuItemID: 1, Quantity: 1},
			},
			prices:  map[uint]int64{},
			wantErr: true,
		},
		{
			name: "Negative price",
			items: []models.CartItem{
				{MenuItemID: 1, Quantity: 1},
			},
			prices: map[uint]int64{
				1: -100,
			},
			wantErr: true,
		},
		{
			name: "Invalid quantity",
			items: []models.CartItem{
				{MenuItemID: 1, Quantity: 0},
			},
			prices: map[uint]int64{
				1: 850,
			},
			wantErr: true,
		},
		{
			name: "Calculation overflow",
			items: []models.CartItem{
				{MenuItemID: 1, Quantity: 2},
			},
			prices: map[uint]int64{
				1: math.MaxInt64,
			},
			wantErr: true,
		},
	}
	// Run each scenario as a separate test.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Calculate the subtotal using the supplied items and prices.
			result, err := CalculateCheckoutSubtotal(tt.items, tt.prices)

			// Check if an error was expected.
			if tt.wantErr {
				if !errors.Is(err, ErrInvalidPrice) {
					t.Errorf("Expected ErrInvalidPrice, got: %v", err)
				}
				return
			}
			// Valid calculations should not return an error.
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			// Check that the calculated subtotal is correct.
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}