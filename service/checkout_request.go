package service

import "strings"

// CheckoutRequest contains the delivery details needed to place an order.
// The customer's ID comes from their JWT, not from this request.
// This structure defines the information our checkout endpoint expects to receive.
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

	if strings.TrimSpace(r.DeliveryName) == "" {
		return ErrInvalidDelivery
	}

	if strings.TrimSpace(r.DeliveryPhone) == "" {
		return ErrInvalidDelivery
	}

	if strings.TrimSpace(r.DeliveryAddress) == "" {
		return ErrInvalidDelivery
	}

	if strings.TrimSpace(r.DeliveryCity) == "" {
		return ErrInvalidDelivery
	}

	if strings.TrimSpace(r.DeliveryState) == "" {
		return ErrInvalidDelivery
	}
	return nil
}