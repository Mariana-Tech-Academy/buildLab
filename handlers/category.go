package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"Fhub/auth"
	"Fhub/models"
	"Fhub/service"
	"Fhub/utils"

	"github.com/gorilla/mux"
)

type CategoryHandler struct {
	CategoryService service.CategoryService
}

func NewCategoryHandler(service service.CategoryService) *CategoryHandler {
	return &CategoryHandler{CategoryService: service}
}

type categoryMenuItemResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int64  `json:"price"`
	ImageURL    string `json:"image_url"`
}

type categoryResponse struct {
	ID           uint                       `json:"id"`
	RestaurantID uint                       `json:"restaurant_id"`
	Name         string                     `json:"name"`
	MenuItems    []categoryMenuItemResponse `json:"menu_items,omitempty"`
}

func newCategoryResponse(category *models.Category) categoryResponse {
	return categoryResponse{
		ID:           category.ID,
		RestaurantID: category.RestaurantID,
		Name:         category.Name,
	}
}

func newCategoryListResponse(categories []models.Category) []categoryResponse {
	out := make([]categoryResponse, 0, len(categories))
	for _, category := range categories {
		items := make([]categoryMenuItemResponse, 0, len(category.MenuItems))
		for _, item := range category.MenuItems {
			if !item.Available {
				continue
			}
			items = append(items, categoryMenuItemResponse{
				ID:          item.ID,
				Name:        item.Name,
				Description: item.Description,
				Price:       item.Price,
				ImageURL:    item.ImageURL,
			})
		}
		out = append(out, categoryResponse{
			ID:           category.ID,
			RestaurantID: category.RestaurantID,
			Name:         category.Name,
			MenuItems:    items,
		})
	}
	return out
}

func (h *CategoryHandler) GetCategories(
	w http.ResponseWriter,
	r *http.Request,
) {
	restaurantID, err := strconv.ParseUint(
		mux.Vars(r)["restaurantId"], 10, 64,
	)
	if err != nil || restaurantID == 0 {
		utils.Error(w, http.StatusBadRequest, "invalid restaurant ID")
		return
	}

	categories, err := h.CategoryService.GetByRestaurantID(
		r.Context(),
		uint(restaurantID),
	)
	if err != nil {
		writeCategoryError(w, err)
		return
	}

	utils.Success(w, http.StatusOK, "categories retrieved successfully", newCategoryListResponse(categories))
}
func (h *CategoryHandler) CreateCategory(
	w http.ResponseWriter,
	r *http.Request,
) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	restaurantID, err := strconv.ParseUint(
		mux.Vars(r)["restaurantId"], 10, 64,
	)
	if err != nil || restaurantID == 0 {
		utils.Error(w, http.StatusBadRequest, "invalid restaurant ID")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var input service.CreateCategoryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	category, err := h.CategoryService.Create(
		r.Context(),
		uint(restaurantID),
		claims.UserID,
		claims.Role,
		input,
	)
	if err != nil {
		writeCategoryError(w, err)
		return
	}

	utils.Success(
		w,
		http.StatusCreated,
		"category created successfully",
		newCategoryResponse(category),
	)
}

func writeCategoryError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrRestaurantNotFound):
		utils.Error(w, http.StatusNotFound,
			"restaurant not found")

	case errors.Is(err, service.ErrRestaurantForbidden):
		utils.Error(w, http.StatusForbidden,
			"you do not have permission to modify this restaurant")

	case errors.Is(err, service.ErrInvalidCategory):
		utils.Error(w, http.StatusUnprocessableEntity,
			err.Error())

	default:
		log.Printf("category handler: %v", err)
		utils.Error(w, http.StatusInternalServerError,
			"something went wrong")
	}
}
