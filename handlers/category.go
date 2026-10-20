package handlers

import (
    "encoding/json"
    "errors"
    "log"
    "net/http"
    "strconv"

    "Fhub/auth"
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

func (h *CategoryHandler) GetCategories(
    w http.ResponseWriter,
    r *http.Request,
) {
    vars := mux.Vars(r)

    restaurantID, err := strconv.ParseUint(
        vars["restaurantId"], 10, 64,
    )

    if err != nil || restaurantID == 0 {
        http.Error(
            w,
            "Invalid restaurant ID",
            http.StatusBadRequest,
        )
        return
    }

    categories, err := h.CategoryService.GetByRestaurantID(
        r.Context(),
        uint(restaurantID),
    )

    if err != nil {
        http.Error(
            w,
            "Failed to retrieve categories",
            http.StatusInternalServerError,
        )
        return
    }

    w.Header().Set("Content-Type", "application/json")

    if err := json.NewEncoder(w).Encode(categories); err != nil {
        return
    }
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
        category,
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
