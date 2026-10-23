package service

import "testing"

// Test that missing or empty delivery information is rejected.
func TestCheckoutRequestInvalid(t *testing.T) {

	// Each test checks a different invalid delivery field.
	tests := []struct {
		name    string
		request CheckoutRequest
	}{
		{
			name: "Missing name",
			request: CheckoutRequest{
				DeliveryPhone:   "07123456789",
				DeliveryAddress: "10 High Street",
				DeliveryCity:    "London",
				DeliveryState:   "Greater London",
			},
		},
		{
			name: "Missing phone",
			request: CheckoutRequest{
				DeliveryName:    "Alex Smith",
				DeliveryAddress: "10 High Street",
				DeliveryCity:    "London",
				DeliveryState:   "Greater London",
			},
		},
		{
			name: "Missing address",
			request: CheckoutRequest{
				DeliveryName:  "Alex Smith",
				DeliveryPhone: "07123456789",
				DeliveryCity:  "London",
				DeliveryState: "Greater London",
			},
		},
		{
			name: "Missing city",
			request: CheckoutRequest{
				DeliveryName:    "Alex Smith",
				DeliveryPhone:   "07123456789",
				DeliveryAddress: "10 High Street",
				DeliveryState:   "Greater London",
			},
		},
		{
			name: "Missing state",
			request: CheckoutRequest{
				DeliveryName:    "Alex Smith",
				DeliveryPhone:   "07123456789",
				DeliveryAddress: "10 High Street",
				DeliveryCity:    "London",
			},
		},
		{
			name: "Spaces only",
			request: CheckoutRequest{
				DeliveryName:    "   ",
				DeliveryPhone:   "07123456789",
				DeliveryAddress: "10 High Street",
				DeliveryCity:    "London",
				DeliveryState:   "Greater London",
			},
		},
	}

	// Run each test and check that validation rejects it.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.request.Validate(); err == nil {
				t.Error("Expected validation error, but got none")
			}
		})
	}
}
