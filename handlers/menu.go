package handlers

import (
    "encoding/json"
    "errors"
    "io"
    "net/http"
    "strconv"
    "strings"

    "Fhub/auth"
    "Fhub/models"
    "Fhub/repository"
    "Fhub/service"
    "Fhub/utils"

    "github.com/gorilla/mux"
)


type MenuHandler struct {
	MenuService service.MenuService
}

func NewMenuHandler(service service.MenuService) *MenuHandler {
	return &MenuHandler{MenuService: service}

}
 
func menuID(r *http.Request, key string) (uint, error) {
    n, err := strconv.ParseUint(mux.Vars(r)[key], 10, 32)
    return uint(n), err
}

func menuError(w http.ResponseWriter, err error) {
    switch {
    case errors.Is(err, repository.ErrMenuItemNotFound):
        utils.Error(w, 404, err.Error())
    case errors.Is(err, repository.ErrMenuItemUnavailable):
        utils.Error(w, 409, err.Error())
    case errors.Is(err, repository.ErrMenuItemForbidden):
        utils.Error(w, 403, err.Error())
    default:
        utils.Error(w, 500, "database error")
    }
}

func (h *MenuHandler) Get(w http.ResponseWriter, r *http.Request) {
    id, err := menuID(r, "id")
    if err != nil || id == 0 {
        utils.Error(w, 400, "invalid menu item id")
        return
    }

    item, err := h.MenuService.Get(id)
    if err != nil {
        menuError(w, err)
        return
    }

    utils.Success(w, 200, "menu item", item)
}

func (h *MenuHandler) List(w http.ResponseWriter, r *http.Request) {
    id, err := menuID(r, "id")
    if err != nil || id == 0 {
        utils.Error(w, 400, "invalid restaurant id")
        return
    }

    items, err := h.MenuService.List(id)
    if err != nil {
        menuError(w, err)
        return
    }

    utils.Success(w, 200, "restaurant menu", items)
}

func (h *MenuHandler) SetAvailability(w http.ResponseWriter, r *http.Request) {
    id, err := menuID(r, "id")
    if err != nil || id == 0 {
        utils.Error(w, 400, "invalid menu item id")
        return
    }

    claims, err := auth.ClaimsFromContext(r.Context())
    if err != nil {
        utils.Error(w, 401, "unauthenticated")
        return
    }

    var req struct {
        Available *bool `json:"available"`
    }

    dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024))
    dec.DisallowUnknownFields()

    if err := dec.Decode(&req); err != nil || req.Available == nil {
        utils.Error(w, 400, "provide available as true or false")
        return
    }

    var extra any
    if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
        utils.Error(w, 400, "request must contain one JSON object")
        return
    }

    item, err := h.MenuService.SetAvailability(
        id,
        claims.UserID,
        claims.Role == models.RoleAdmin,
        *req.Available,
    )

    if err != nil {
        menuError(w, err)
        return
    }

    utils.Success(w, 200, "menu item availability updated", item)
}


func (h *MenuHandler) Search(w http.ResponseWriter, r *http.Request) {
    params := r.URL.Query()

    filters := repository.MenuSearchFilters{
        Name: strings.TrimSpace(params.Get("name")),
    }

    // Filter by category
    if params.Has("category_id") {
        id, err := strconv.ParseUint(params.Get("category_id"), 10, 32)
        if err != nil || id == 0 {
            http.Error(w, "Invalid category_id", http.StatusBadRequest)
            return
        }

        categoryID := uint(id)
        filters.CategoryID = &categoryID
    }

    // Filter by restaurant
    if params.Has("restaurant_id") {
        id, err := strconv.ParseUint(params.Get("restaurant_id"), 10, 32)
        if err != nil || id == 0 {
            http.Error(w, "Invalid restaurant_id", http.StatusBadRequest)
            return
        }

        restaurantID := uint(id)
        filters.RestaurantID = &restaurantID
    }

    // Minimum price
    if params.Has("min_price") {
        price, err := strconv.ParseInt(params.Get("min_price"), 10, 64)
        if err != nil || price < 0 {
            http.Error(w, "Invalid min_price", http.StatusBadRequest)
            return
        }

        filters.MinPrice = &price
    }

    // Maximum price
    if params.Has("max_price") {
        price, err := strconv.ParseInt(params.Get("max_price"), 10, 64)
        if err != nil || price < 0 {
            http.Error(w, "Invalid max_price", http.StatusBadRequest)
            return
        }

        filters.MaxPrice = &price
    }

    // Check price range
    if filters.MinPrice != nil && filters.MaxPrice != nil {
        if *filters.MinPrice > *filters.MaxPrice {
            http.Error(w, "Minimum price cannot exceed maximum price", http.StatusBadRequest)
            return
        }
    }

    // Filter by availability
    if params.Has("available") {
        raw := params.Get("available")

        if raw != "true" && raw != "false" {
            http.Error(w, "Available must be true or false", http.StatusBadRequest)
            return
        }

        available, _ := strconv.ParseBool(raw)
        filters.Available = &available
    }

    // Call service to search database
    items, err := h.MenuService.Search(filters)

    if err != nil {
        http.Error(w, "Failed to search menu items", http.StatusInternalServerError)
        return
    }

    // Return JSON response
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)

    if items == nil {
        items = []models.MenuItem{}
    }

    json.NewEncoder(w).Encode(items)
}
