package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"Fhub/auth"
	"Fhub/models"
	"Fhub/service"
	"Fhub/utils"
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

func writeMenuError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidMenuItemInput):
		utils.Error(w, http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, service.ErrRestaurantNotFound):
		utils.Error(w, http.StatusNotFound, "restaurant not found")
	case errors.Is(err, service.ErrCategoryNotFound):
		utils.Error(w, http.StatusNotFound, "category not found")
	case errors.Is(err, service.ErrRestaurantForbidden):
		utils.Error(w, http.StatusForbidden, "you do not have permission to modify this restaurant")
	default:
		log.Printf("create menu item: %v", err)
		utils.Error(w, http.StatusInternalServerError, "internal server error")
	}
}
