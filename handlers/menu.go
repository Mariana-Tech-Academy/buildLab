package handlers

import (
	"Fhub/auth"
	"Fhub/service"
	"encoding/json"
	"errors"
	"net/http"
)

type MenuHandler struct {
	MenuService service.MenuService
}

// NewMenuHandler creates a handler connected to the menu service.
func NewMenuHandler(menuService service.MenuService) *MenuHandler {
	return &MenuHandler{MenuService: menuService}
}

// CreateMenuitem received a request to create a new menu item.
func (h *MenuHandler) CreateMenuItem(w http.ResponseWriter, r *http.Request) {

	//First step is to get the authenticated user's information from the JWT.
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unathorized", http.StatusUnauthorized)
		return
	}

	//Second step is to prepare a variable to hold the JSON request data.
	var input service.CreateMenuItemInput

	//Third step is to convert the JSON request body into the input struct.
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Invalid JSON request", http.StatusBadRequest)
		return
	}

	//Fourth Step is to send the request to the menu service, the service validates
	//the information and checks restaurant ownership.
	item, err := h.MenuService.CreateMenuItem(r.Context(), claims.UserID, claims.Role, input)

	//Fifth step is to handle any erros returned by the service.
	if err != nil {

		//It returns 400 if the menu item infromation is invalid.
		if errors.Is(err, service.ErrInvalidMenuItemInput) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// It returns 403 if the user does not manage the restaurant.
		if errors.Is(err, service.ErrRestaurantForbidden) {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		// It returns 500 if an unexpected server error occurs.
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	// Sixt step is to return the HTTP 201 created when the menu item is saved.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	//Seventh step is to return the created menu item as JSON.
	json.NewEncoder(w).Encode(item)
}
