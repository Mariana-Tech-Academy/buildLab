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

// CartHandler connects HTTP requests to the cart business logic.
type CartHandler struct {
	CartService service.CartService
}

// NewCartHandler creates a new cart handler with its service.
func NewCartHandler(cartService service.CartService) *CartHandler {
	return &CartHandler{CartService: cartService}
}

type cartItemResponse struct {
	CartID       uint   `json:"cart_id"`
	RestaurantID uint   `json:"restaurant_id"`
	MenuItemID   uint   `json:"menu_item_id"`
	Name         string `json:"name"`
	Quantity     int    `json:"quantity"`
	Price        int64  `json:"price"`
}

// AddItemToCart handles requests to add menu items to a customer's cart.
func (h *CartHandler) AddItemToCart(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var request struct {
		MenuItemID uint `json:"menu_item_id"`
		Quantity   int  `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if request.MenuItemID == 0 {
		utils.Error(w, http.StatusBadRequest, "invalid menu item ID")
		return
	}

	item, err := h.CartService.AddItemToCart(
		r.Context(),
		claims.UserID,
		request.MenuItemID,
		request.Quantity,
	)
	if err != nil {
		writeCartError(w, err)
		return
	}

	utils.Success(w, http.StatusOK, "item added to cart", cartItemResponse{
		CartID:       item.CartID,
		RestaurantID: item.RestaurantID,
		MenuItemID:   item.MenuItemID,
		Name:         item.Name,
		Quantity:     item.Quantity,
		Price:        item.Price,
	})
}

func writeCartError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidCartInput):
		utils.Error(w, http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, service.ErrMenuItemNotFound):
		utils.Error(w, http.StatusNotFound, "menu item not found")
	case errors.Is(err, service.ErrCartRestaurantConflict):
		utils.Error(w, http.StatusConflict, "cannot add items from different restaurants")
	default:
		log.Printf("cart handler: %v", err)
		utils.Error(w, http.StatusInternalServerError, "internal server error")
	}
}
