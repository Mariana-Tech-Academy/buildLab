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

	// Decode the JSON request body into the delivery information structure.
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid JSON request body")
		return
	}

	// Call the service layer to validate the delivery information.
	validationErrors := service.ValidateDeliveryInfo(&input)

	// If validation fails, return HTTP 400 with the specific field errors.
	if len(validationErrors) > 0 {
		utils.JSON(w, http.StatusBadRequest, utils.Response{
			Success: false,
			Message: "Delivery information validation failed",
			Data:    validationErrors,
		})
		return
	}

	// If validation succeeds, return HTTP 200 with the validated details.
	utils.Success(w, http.StatusOK, "Delivery information is valid", input)
}
