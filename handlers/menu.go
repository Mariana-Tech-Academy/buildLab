package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"Fhub/auth"
	"Fhub/models"
	"Fhub/service"
	"Fhub/utils"

	"github.com/gorilla/mux"
)

type MenuHandler struct {
	MenuService service.MenuService
}

// NewMenuHandler creates a handler connected to the menu service.
func NewMenuHandler(menuService service.MenuService) *MenuHandler {
	return &MenuHandler{MenuService: menuService}
}

type menuItemResponse struct {
	ID           uint      `json:"id"`
	RestaurantID uint      `json:"restaurant_id"`
	CategoryID   uint      `json:"category_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Price        int64     `json:"price"`
	ImageURL     string    `json:"image_url"`
	Available    bool      `json:"available"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func newMenuItemResponse(item *models.MenuItem) menuItemResponse {
	return menuItemResponse{
		ID:           item.ID,
		RestaurantID: item.RestaurantID,
		CategoryID:   item.CategoryID,
		Name:         item.Name,
		Description:  item.Description,
		Price:        item.Price,
		ImageURL:     item.ImageURL,
		Available:    item.Available,
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
	}
}

// CreateMenuItem receives a request to create a new menu item.
// Price in the JSON body is minor units (pence or kobo): 500 is 5.00, not 500 pounds.
func (h *MenuHandler) CreateMenuItem(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var input service.CreateMenuItemInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	item, err := h.MenuService.CreateMenuItem(r.Context(), claims.UserID, claims.Role, input)
	if err != nil {
		writeMenuError(w, err)
		return
	}

	utils.Success(w, http.StatusCreated, "menu item created", newMenuItemResponse(item))
}

// GetMenuItemByID handles GET /api/menu/{id}.
func (h *MenuHandler) GetMenuItemByID(w http.ResponseWriter, r *http.Request) {
	itemID, err := parseMenuItemID(r)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid menu item ID")
		return
	}

	item, err := h.MenuService.GetMenuItemByID(r.Context(), itemID)
	if err != nil {
		writeMenuError(w, err)
		return
	}

	utils.Success(w, http.StatusOK, "menu item retrieved successfully", newMenuItemResponse(item))
}

// UpdateMenuItem handles PUT /api/menu/{id}.
func (h *MenuHandler) UpdateMenuItem(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	itemID, err := parseMenuItemID(r)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid menu item ID")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var input service.UpdateMenuItemInput
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid update JSON")
		return
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		utils.Error(w, http.StatusBadRequest, "request must contain one JSON object")
		return
	}

	item, err := h.MenuService.UpdateMenuItem(r.Context(), itemID, claims.UserID, claims.Role, input)
	if err != nil {
		writeMenuError(w, err)
		return
	}

	utils.Success(w, http.StatusOK, "menu item updated", newMenuItemResponse(item))
}

// SetAvailability handles PATCH /api/menu/{id}/availability.
func (h *MenuHandler) SetAvailability(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	itemID, err := parseMenuItemID(r)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid menu item ID")
		return
	}

	var req struct {
		Available *bool `json:"available"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil || req.Available == nil {
		utils.Error(w, http.StatusBadRequest, "provide available as true or false")
		return
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		utils.Error(w, http.StatusBadRequest, "request must contain one JSON object")
		return
	}

	item, err := h.MenuService.SetAvailability(r.Context(), itemID, claims.UserID, claims.Role, *req.Available)
	if err != nil {
		writeMenuError(w, err)
		return
	}
	utils.Success(w, http.StatusOK, "menu item availability updated", newMenuItemResponse(item))
}

// DeleteMenuItem handles DELETE /api/menu/{id}.
func (h *MenuHandler) DeleteMenuItem(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	itemID, err := parseMenuItemID(r)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid menu item ID")
		return
	}

	if err := h.MenuService.DeleteMenuItem(r.Context(), itemID, claims.UserID, claims.Role); err != nil {
		writeMenuError(w, err)
		return
	}

	utils.Success(w, http.StatusOK, "menu item deleted", nil)
}

func parseMenuItemID(r *http.Request) (uint, error) {
	id, err := strconv.ParseUint(mux.Vars(r)["id"], 10, strconv.IntSize)
	if err != nil {
		return 0, err
	}
	if id == 0 {
		return 0, errors.New("menu item ID must be greater than zero")
	}
	return uint(id), nil
}

func writeMenuError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidMenuItemInput):
		utils.Error(w, http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, service.ErrMenuItemNotFound):
		utils.Error(w, http.StatusNotFound, "menu item not found")
	case errors.Is(err, service.ErrRestaurantNotFound):
		utils.Error(w, http.StatusNotFound, "restaurant not found")
	case errors.Is(err, service.ErrCategoryNotFound):
		utils.Error(w, http.StatusNotFound, "category not found")
	case errors.Is(err, service.ErrRestaurantForbidden):
		utils.Error(w, http.StatusForbidden, "you do not have permission to modify this restaurant")
	default:
		log.Printf("menu handler: %v", err)
		utils.Error(w, http.StatusInternalServerError, "internal server error")
	}
}

func (h *MenuHandler) GetRestaurantMenu(w http.ResponseWriter, r *http.Request) {
	restaurantID, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 64)
	if err != nil || restaurantID == 0 {
		utils.Error(w, http.StatusBadRequest, "invalid restaurant ID")
		return
	}

	var categoryID *uint
	if value := r.URL.Query().Get("category_id"); value != "" {
		parsedID, err := strconv.ParseUint(value, 10, 64)
		if err != nil || parsedID == 0 {
			utils.Error(w, http.StatusBadRequest, "invalid category ID")
			return
		}
		id := uint(parsedID)
		categoryID = &id
	}

	items, err := h.MenuService.ListRestaurantMenu(r.Context(), uint(restaurantID), categoryID)
	if err != nil {
		writeMenuError(w, err)
		return
	}

	responses := make([]menuItemResponse, 0, len(items))
	for i := range items {
		responses = append(responses, newMenuItemResponse(&items[i]))
	}
	utils.Success(w, http.StatusOK, "menu retrieved successfully", responses)
}
