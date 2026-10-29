package handlers

import (
	"encoding/json"
	"net/http"

	"Fhub/service"
	"Fhub/utils"
	"errors"
	"strconv"
	"Fhub/auth"
	"Fhub/models"

	"github.com/gorilla/mux"
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

// updateOrderStatusRequest represents the new order status.
type updateOrderStatusRequest struct {
	Status models.OrderStatus `json:"status"`
}

// UpdateOrderStatus allows managers to move orders through their stages.
func (h *OrderHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {

	// Get the authenticated user's details.
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	// Get restaurant ID from URL.
	restaurantID, err := strconv.ParseUint(
		mux.Vars(r)["restaurantId"], 10, 64,
	)
	if err != nil || restaurantID == 0 {
		utils.Error(w, http.StatusBadRequest, "invalid restaurant ID")
		return
	}

	// Get order ID from URL.
	orderID, err := strconv.ParseUint(
		mux.Vars(r)["orderId"], 10, 64,
	)
	if err != nil || orderID == 0 {
		utils.Error(w, http.StatusBadRequest, "invalid order ID")
		return
	}

	// Read the new status from the request body.
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req updateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Call the service to validate and update the order.
	err = h.OrderService.UpdateOrderStatus(
		r.Context(),
		uint(orderID),
		uint(restaurantID),
		claims.UserID,
		claims.Role,
		req.Status,
	)

	if err != nil {
		switch {
		case errors.Is(err, service.ErrOrderNotFound),
			errors.Is(err, service.ErrRestaurantNotFound):
			utils.Error(w, http.StatusNotFound, err.Error())

		case errors.Is(err, service.ErrOrderForbidden),
			errors.Is(err, service.ErrRestaurantForbidden):
			utils.Error(w, http.StatusForbidden, err.Error())

		case errors.Is(err, service.ErrInvalidOrderStatus):
			utils.Error(w, http.StatusBadRequest, err.Error())

		case errors.Is(err, service.ErrInvalidStatusTransition),
			errors.Is(err, service.ErrOrderStatusFinal):
			utils.Error(w, http.StatusConflict, err.Error())

		default:
			utils.Error(w, http.StatusInternalServerError, "could not update order status")
		}
		return
	}

	utils.Success(
		w,
		http.StatusOK,
		"order status updated successfully",
		map[string]interface{}{
			"order_id": orderID,
			"status":   req.Status,
		},
	)
}
