package handlers

import (
	"Fhub/auth"
	"Fhub/models"
	"Fhub/service"
	"Fhub/utils"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

// CartHandler connect cart HTTP requests to the cart service
type CartHandler struct {
	CartService service.CartService
}

// NewcartHandler creates the cart handler with it's service
func NewCartHandler(service service.CartService) *CartHandler {
	return &CartHandler{CartService: service}

}

type updateQuantityRequest struct {
	Quantity int `json:"quantity"`
}

type cartItemResponse struct {
	ID         uint   `json:"id"`
	MenuItemID uint   `json:"menu_item_id"`
	Name       string `json:"name"`
	UnitPrice  int64  `json:"unit_price"`
	Quantity   int    `json:"quantity"`
	LineTotal  int64  `json:"line_total"`
}

type cartResponse struct {
	ID           uint               `json:"id"`
	RestaurantID *uint              `json:"restaurant_id"`
	Items        []cartItemResponse `json:"items"`
	Total        int64              `json:"total"`
}

func newCartResponse(cart *models.Cart) cartResponse {
	items := make([]cartItemResponse, 0, len(cart.Items))
	var total int64

	for _, item := range cart.Items {
		lineTotal := item.MenuItem.Price * int64(item.Quantity)
		total += lineTotal

		items = append(items, cartItemResponse{
			ID:         item.ID,
			MenuItemID: item.MenuItemID,
			Name:       item.MenuItem.Name,
			UnitPrice:  item.MenuItem.Price,
			Quantity:   item.Quantity,
			LineTotal:  lineTotal,
		})
	}

	return cartResponse{
		ID:           cart.ID,
		RestaurantID: cart.RestaurantID,
		Items:        items,
		Total:        total,
	}
}

// writeCartError turns service errrors into the right status code & message
func writeCartError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidQuantity):
		utils.Error(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrCartItemNotFound):
		utils.Error(w, http.StatusNotFound, "cart item not found")
	default:
		log.Printf("cart handler: %v", err)
		utils.Error(w, http.StatusInternalServerError, "something went wrong")
	}
}

func (h *CartHandler) UpdateQuantity(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	cartItemID, ok := muxID(w, r)
	if !ok {
		return
	}

	var req updateQuantityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// asks the service to update the qty in this customer's cart
	cart, err := h.CartService.UpdateQuantity(r.Context(), claims.UserID, cartItemID, service.UpdateCartQuantityInput{
		Quantity: req.Quantity,
	})
	if err != nil {
		writeCartError(w, err)
		return
	}

	//sends back the updated cart with the new line totals and cart total
	utils.Success(w, http.StatusOK, "cart item quantity updated", newCartResponse(cart))
}
