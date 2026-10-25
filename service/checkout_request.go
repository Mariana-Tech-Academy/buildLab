package service

import "strings"

// CheckoutRequest contains the customer's delivery details.
// The customer ID is taken from their JWT rather than the request.
type CheckoutRequest struct {
	DeliveryName         string `json:"delivery_name"`
	DeliveryPhone        string `json:"delivery_phone"`
	DeliveryAddress      string `json:"delivery_address"`
	DeliveryCity         string `json:"delivery_city"`
	DeliveryState        string `json:"delivery_state"`
	DeliveryInstructions string `json:"delivery_instructions"`
}

// Validate checks that the required delivery details have been provided.
func (r CheckoutRequest) Validate() error {

	// Check that the customer has provided their name.
	if strings.TrimSpace(r.DeliveryName) == "" {
		return ErrInvalidDelivery
	}

	// Check that the customer has provided their phone number.
	if strings.TrimSpace(r.DeliveryPhone) == "" {
		return ErrInvalidDelivery
	}

	// Check that the delivery address has been provided.
	if strings.TrimSpace(r.DeliveryAddress) == "" {
		return ErrInvalidDelivery
	}

	// Check that the delivery city has been provided.
	if strings.TrimSpace(r.DeliveryCity) == "" {
		return ErrInvalidDelivery
	}

	// Check that the delivery state has been provided.
	if strings.TrimSpace(r.DeliveryState) == "" {
		return ErrInvalidDelivery
	}

	return nil
}
