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

// OrderHandler handles HTTP requests related to orders.
type OrderHandler struct {
	OrderService service.OrderService
}

// NewOrderHandler creates a new order handler
// and connects it to the order service.
func NewOrderHandler(service service.OrderService) *OrderHandler {
	return &OrderHandler{
		OrderService: service,
	}
}

// ValidateDeliveryInfo receives delivery information from the customer,
// checks that the required fields are provided,
// and returns a JSON response.
func (h *OrderHandler) ValidateDeliveryInfo(w http.ResponseWriter, r *http.Request) {

	// Create a variable to store the customer's delivery information.
	var input service.DeliveryInfoInput

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	validationErrors := h.OrderService.ValidateDeliveryInfo(&input)
	if len(validationErrors) > 0 {
		utils.JSON(w, http.StatusUnprocessableEntity, utils.Response{
			Success: false,
			Message: "delivery information is invalid",
			Data:    validationErrors,
		})
		return
	}

	utils.Success(w, http.StatusOK, "delivery information is valid", input)
}

func (h *OrderHandler) GetCheckoutSummary(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	summary, err := h.OrderService.GetCheckoutSummary(
		r.Context(),
		claims.UserID,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmptyCart):
			utils.Error(w, http.StatusUnprocessableEntity, "cannot checkout with an empty cart")
		case errors.Is(err, service.ErrUnavailableItem):
			utils.Error(w, http.StatusConflict, "cart contains an unavailable item")
		case errors.Is(err, service.ErrRestaurantNotFound):
			utils.Error(w, http.StatusNotFound, "restaurant not found")
		default:
			log.Printf("checkout summary: %v", err)
			utils.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	utils.Success(w, http.StatusOK, "checkout summary", summary)
}
