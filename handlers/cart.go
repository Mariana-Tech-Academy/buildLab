package handlers

import (
	"Fhub/auth"
	"Fhub/repository"
	"Fhub/service"
	"Fhub/utils"
	"encoding/json"
	"errors"
	"net/http"
)

type CartHandler struct {
	CartService service.CartService
}

func NewCartHandler(service service.CartService) *CartHandler {
	return &CartHandler{CartService: service}
	
}
// 1.AddItem handles the HTTP request to add an item to the user's cart.
// It expects a JSON body with the menu_item_id and quantity.
func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
// 2. Parse the request body to get the menu item ID and quantity.
	var input struct {
		MenuItemID uint `json:"menu_item_id"`
		Quantity   int  `json:"quantity"`
	}
// 3. Validate the input and call the CartService to add the item to the cart.
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
// 4. Check if the input is valid.
	if input.MenuItemID == 0 || input.Quantity <= 0 {
		utils.Error(w, http.StatusBadRequest,
			"menu_item_id and a positive quantity are required")
		return
	}
// 5. Call the CartService to add the item to the cart.
	err = h.CartService.AddItem(claims.UserID, input.MenuItemID, input.Quantity)
	if errors.Is(err, repository.ErrCartRestaurantMismatch) {
		utils.Error(w, http.StatusConflict, err.Error())
		return
	}
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.Success(w, http.StatusOK, "item added to cart", nil)
}