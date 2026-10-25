package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"Fhub/auth"
	"Fhub/service"
	"Fhub/utils"
)

// CheckoutHandler connects checkout requests to the checkout service.
type CheckoutHandler struct {
	checkoutService service.CheckoutService
}

// NewCheckoutHandler creates a checkout handler.
func NewCheckoutHandler(checkoutService service.CheckoutService) *CheckoutHandler {
	return &CheckoutHandler{
		checkoutService: checkoutService,
	}
}

// PlaceOrder handles a customer's checkout request.
func (h *CheckoutHandler) PlaceOrder(w http.ResponseWriter, r *http.Request) {

	// Get the authenticated customer's details from the request.
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}
	// Check that the customer ID is valid.
	if claims.UserID == 0 {
		utils.Error(w, http.StatusUnauthorized, "invalid customer")
		return
	}
	// Read the delivery details sent by the customer.
	var request service.CheckoutRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid checkout request")
		return
	}
	// Check that the required delivery details have been provided.
	if err := request.Validate(); err != nil {
		utils.Error(w, http.StatusBadRequest, "delivery details are incomplete")
		return
	}
	// Pass the checkout request to the service.
	order, err := h.checkoutService.PlaceOrder(
		r.Context(),
		claims.UserID,
		request,
	)
	if err != nil {

		// Return a validation error if the cart is empty.
		if errors.Is(err, service.ErrCartEmpty) {
			utils.Error(w, http.StatusBadRequest, "your cart is empty")
			return
		}
		// Return a server error if checkout cannot be completed.
		utils.Error(w, http.StatusInternalServerError, "checkout could not be completed")
		return
	}
	// Return the new order when checkout succeeds.
	utils.Success(w, http.StatusCreated, "order placed successfully", order)
}
