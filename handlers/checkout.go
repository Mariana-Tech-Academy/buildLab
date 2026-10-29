package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"Fhub/auth"
	"Fhub/service"
	"Fhub/utils"
)

// CheckoutHandler connects checkout requests to the checkout service.
type CheckoutHandler struct {
	checkoutService service.CheckoutService
	orders          service.OrderService
}

// NewCheckoutHandler creates a checkout handler.
func NewCheckoutHandler(checkoutService service.CheckoutService, orders service.OrderService) *CheckoutHandler {
	return &CheckoutHandler{
		checkoutService: checkoutService,
		orders:          orders,
	}
}

// PlaceOrder handles a customer's checkout request.
func (h *CheckoutHandler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}
	if claims.UserID == 0 {
		utils.Error(w, http.StatusUnauthorized, "invalid customer")
		return
	}

	var request service.DeliveryInfoInput
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	validationErrors := h.orders.ValidateDeliveryInfo(&request)
	if len(validationErrors) > 0 {
		utils.JSON(w, http.StatusUnprocessableEntity, utils.Response{
			Success: false,
			Message: "delivery information is invalid",
			Data:    validationErrors,
		})
		return
	}

	order, err := h.checkoutService.PlaceOrder(r.Context(), claims.UserID, request)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmptyCart):
			utils.Error(w, http.StatusUnprocessableEntity, "cannot checkout with an empty cart")
		case errors.Is(err, service.ErrUnavailableItem):
			utils.Error(w, http.StatusConflict, "cart contains an unavailable item")
		case errors.Is(err, service.ErrRestaurantNotFound):
			utils.Error(w, http.StatusNotFound, "restaurant not found")
		case errors.Is(err, service.ErrInvalidDelivery):
			utils.Error(w, http.StatusUnprocessableEntity, "delivery information is invalid")
		default:
			log.Printf("checkout: %v", err)
			utils.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	utils.Success(w, http.StatusCreated, "order placed successfully", newCustomerOrderResponse(order))
}
