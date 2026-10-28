package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"Fhub/auth"
	"Fhub/models"
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
		utils.Error(w, http.StatusConflict, "cart contains items from another restaurant; clear the cart before adding this item")
	case errors.Is(err, service.ErrCartItemNotFound):
		utils.Error(w, http.StatusNotFound, "cart item not found")
	default:
		log.Printf("cart handler: %v", err)
		utils.Error(w, http.StatusInternalServerError, "internal server error")
	}
}

type updateQuantityRequest struct {
	Quantity int `json:"quantity"`
}

type cartLineResponse struct {
	ID         uint   `json:"id"`
	MenuItemID uint   `json:"menu_item_id"`
	Name       string `json:"name"`
	UnitPrice  int64  `json:"unit_price"`
	Quantity   int    `json:"quantity"`
	LineTotal  int64  `json:"line_total"`
}

type updatedCartResponse struct {
	ID           uint               `json:"id"`
	RestaurantID *uint              `json:"restaurant_id"`
	Items        []cartLineResponse `json:"items"`
	Total        int64              `json:"total"`
}

func newUpdatedCartResponse(cart *models.Cart) updatedCartResponse {
	items := make([]cartLineResponse, 0, len(cart.Items))
	var total int64
	for _, item := range cart.Items {
		lineTotal := item.MenuItem.Price * int64(item.Quantity)
		total += lineTotal
		items = append(items, cartLineResponse{
			ID:         item.ID,
			MenuItemID: item.MenuItemID,
			Name:       item.MenuItem.Name,
			UnitPrice:  item.MenuItem.Price,
			Quantity:   item.Quantity,
			LineTotal:  lineTotal,
		})
	}
	return updatedCartResponse{
		ID:           cart.ID,
		RestaurantID: cart.RestaurantID,
		Items:        items,
		Total:        total,
	}
}

// UpdateQuantity handles PUT /api/cart/items/{id}.
func (h *CartHandler) UpdateQuantity(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	cartItemID, ok := muxID(w, r)
	if !ok {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req updateQuantityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cart, err := h.CartService.UpdateQuantity(r.Context(), claims.UserID, cartItemID, service.UpdateCartQuantityInput{
		Quantity: req.Quantity,
	})
	if err != nil {
		writeCartError(w, err)
		return
	}

	utils.Success(w, http.StatusOK, "cart item quantity updated", newUpdatedCartResponse(cart))
}

// RemoveItem handles DELETE /api/cart/items/{id}.
func (h *CartHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	cartItemID, ok := muxID(w, r)
	if !ok {
		return
	}

	cart, err := h.CartService.RemoveItem(r.Context(), claims.UserID, cartItemID)
	if err != nil {
		writeCartError(w, err)
		return
	}
	utils.Success(w, http.StatusOK, "cart item removed", newUpdatedCartResponse(cart))
}

// Clear handles DELETE /api/cart.
func (h *CartHandler) Clear(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	cart, err := h.CartService.Clear(r.Context(), claims.UserID)
	if err != nil {
		writeCartError(w, err)
		return
	}
	utils.Success(w, http.StatusOK, "cart cleared", newUpdatedCartResponse(cart))
}
