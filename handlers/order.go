package handlers

import (
	"net/http"
	"strconv"

	"Fhub/auth"
	"Fhub/service"
	"Fhub/utils"

	"errors"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type OrderHandler struct {
	OrderService service.OrderService
}

func NewOrderHandler(service service.OrderService) *OrderHandler {
	return &OrderHandler{OrderService: service}

}

// GetOrderByID retrieves an order using its ID.
func (h *OrderHandler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	orderID, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 64)
	if err != nil || orderID == 0 {
		utils.Error(w, http.StatusBadRequest, "invalid order ID")
		return
	}
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	order, err := h.OrderService.GetOrderByID(
		r.Context(),
		uint(orderID),
		claims.UserID,
	)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Error(w, http.StatusNotFound, "order not found")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "unable to retrieve order")
		return
	}
	utils.Success(w, http.StatusOK, "order retrieved successfully", order)

}
