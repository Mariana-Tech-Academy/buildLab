package handlers

import (
	"Fhub/auth"
	"Fhub/service"
	"Fhub/utils"
	"errors"
	"net/http"

	"gorm.io/gorm"
)

var ErrUnavailableItem = errors.New("Item is not available")
var ErrEmptyCart = errors.New("Empty cart")

type OrderHandler struct {
	OrderService service.OrderService
}

func NewOrderHandler(service service.OrderService) *OrderHandler {
	return &OrderHandler{OrderService: service}

}

func (h *OrderHandler) GetCheckoutSummary(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ClaimsFromContext(r.Context())   
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	summary, err := h.OrderService.GetCheckoutSummary(
		r.Context(),
		claims.UserID,
	)

	if err != nil {
		switch {
		case errors.Is(err, ErrEmptyCart):
			utils.Error(
				w,
				http.StatusUnprocessableEntity,
				"cannot checkout with an empty cart",
			)
		case errors.Is(err, ErrUnavailableItem):
			utils.Error(w, http.StatusConflict, err.Error())
		case errors.Is(err, gorm.ErrRecordNotFound):
			utils.Error(w, http.StatusNotFound, "cart not found")
		default:
			utils.Error(
				w,
				http.StatusInternalServerError,
				"failed to get checkout summary",
			)
		}
		return
	}

	utils.Success(w, http.StatusOK, "checkout summary", summary)
}
