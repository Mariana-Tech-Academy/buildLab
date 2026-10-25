package handlers

import (
	"encoding/json"
	"net/http"

	"Fhub/auth"
	"Fhub/service"
)

// CartHandler connects HTTP requests to the cart business logic.
type CartHandler struct {
	CartService service.CartService
}

// NewCartHandler creates a new cart handler with its service.
func NewCartHandler(cartService service.CartService) *CartHandler {
	return &CartHandler{CartService: cartService}
}

// AddItemToCart handles requests to add menu items to a customer's cart.
func (h *CartHandler) AddItemToCart(w http.ResponseWriter, r *http.Request) {

	// STEP 1: Read the authenticated customer's details from the JWT context.
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// STEP 2: Define the JSON fields expected from Postman.
	var request struct {
		MenuItemID uint `json:"menu_item_id"`
		Quantity   int  `json:"quantity"`
	}

	// STEP 3: Convert the incoming JSON into our request structure.
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON request", http.StatusBadRequest)
		return
	}

	// STEP 4: Check that the menu item ID and quantity are valid.
	if request.MenuItemID == 0 || request.Quantity < 1 {
		http.Error(w, "Invalid menu item ID or quantity", http.StatusBadRequest)
		return
	}

	// STEP 5: Call the service to add the item to this customer's cart.
	cart, err := h.CartService.AddItemToCart(
		r.Context(),
		claims.UserID,
		request.MenuItemID,
		request.Quantity,
	)

	// STEP 6: Return an error if the cart operation fails.
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// STEP 7: Tell the client that the request succeeded.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// STEP 8: Return the updated cart as JSON.
	if err := json.NewEncoder(w).Encode(cart); err != nil {
		return
	}
}
