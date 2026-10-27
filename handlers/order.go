package handlers

import (
	"encoding/json"
	"net/http"

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
