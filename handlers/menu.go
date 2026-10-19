package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"Fhub/auth"
	"Fhub/service"
	"Fhub/utils"

	"github.com/gorilla/mux"
)

// MenuHandler stores the service used to process menu requests.
type MenuHandler struct {
	MenuService service.MenuService
}

// NewMenuHandler receives and stores the menu service.
func NewMenuHandler(menuService service.MenuService) *MenuHandler {
	return &MenuHandler{MenuService: menuService}
}

// UpdateMenuItem handles PUT /api/menu/{id}.
func (h *MenuHandler) UpdateMenuItem(w http.ResponseWriter, r *http.Request) {
	// Read the authenticated user's details.
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Read and validate the menu item ID from the URL.
	itemID, err := parseMenuItemID(r)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid menu item ID")
		return
	}

	// Limit the request body to 1 MiB.
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	// Prepare a variable to hold the fields being edited.
	var input service.UpdateMenuItemInput

	// Decode the JSON body into the input variable.
	decoder := json.NewDecoder(r.Body)

	// Reject fields that are not part of UpdateMenuItemInput.
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid update JSON")
		return
	}

	// Reject additional JSON values after the first.
	if err := decoder.Decode(new(any)); err != io.EOF {
		utils.Error(w, http.StatusBadRequest, "request must contain one JSON object")
		return
	}

	// Call the service to check ownership, validate and save changes.
	item, err := h.MenuService.UpdateMenuItem(r.Context(), itemID, claims.UserID, claims.Role, input)
	if err != nil {
		writeMenuOperationError(w, err)
		return
	}

	// Return the updated item.
	utils.Success(w, http.StatusOK, "menu item updated", item)
}

// DeleteMenuItem handles DELETE /api/menu/{id}.
func (h *MenuHandler) DeleteMenuItem(w http.ResponseWriter, r *http.Request) {
	// Read the authenticated user's details.
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Read and validate the menu item ID from the URL.
	itemID, err := parseMenuItemID(r)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid menu item ID")
		return
	}

	// Call the service to check ownership and request deletion.
	err = h.MenuService.DeleteMenuItem(r.Context(), itemID, claims.UserID, claims.Role)
	if err != nil {
		writeMenuOperationError(w, err)
		return
	}

	// Confirm deletion.
	utils.Success(w, http.StatusOK, "menu item deleted", nil)
}

// parseMenuItemID converts the URL's ID into a positive uint.
// Both handlers use this helper.
func parseMenuItemID(r *http.Request) (uint, error) {
	value := mux.Vars(r)["id"]

	id, err := strconv.ParseUint(value, 10, strconv.IntSize)
	if err != nil {
		return 0, err
	}

	if id == 0 {
		return 0, errors.New("menu item ID must be greater than zero")
	}

	return uint(id), nil
}

// writeMenuOperationError sends the appropriate HTTP error response.
// Both handlers use this helper.
func writeMenuOperationError(w http.ResponseWriter, err error) {
	switch {
	// 404: the menu item does not exist.
	case errors.Is(err, service.ErrMenuItemNotFound):
		utils.Error(w, http.StatusNotFound, "menu item not found")

	// 404: the restaurant does not exist.
	case errors.Is(err, service.ErrRestaurantNotFound):
		utils.Error(w, http.StatusNotFound, "restaurant not found")

	// 403: the user does not have permission.
	case errors.Is(err, service.ErrRestaurantForbidden):
		utils.Error(w, http.StatusForbidden, "you cannot manage this restaurant's menu")

	// 400: the supplied edit details are invalid.
	case errors.Is(err, service.ErrInvalidMenuItem):
		utils.Error(w, http.StatusBadRequest, err.Error())

	// 500: an unexpected error occurred.
	default:
		utils.Error(w, http.StatusInternalServerError, "unable to complete menu operation")
	}
}
