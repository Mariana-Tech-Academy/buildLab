package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"Fhub/auth"
	"Fhub/service"
	"Fhub/utils"

	"github.com/gorilla/mux"
)

type RestaurantHandler struct {
	RestaurantService service.RestaurantService
}

func NewRestaurantHandler(
	restaurantService service.RestaurantService,
) *RestaurantHandler {
	return &RestaurantHandler{
		RestaurantService: restaurantService,
	}
}

func (h *RestaurantHandler) UpdateRestaurant(
	w http.ResponseWriter,
	r *http.Request,
) {
	// 1. Get the authenticated user's claims.
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	// 2. Get restaurant ID from the URL.
	restaurantID, err := strconv.ParseUint(
		mux.Vars(r)["id"], 10, 64,
	)
	if err != nil || restaurantID == 0 {
		utils.Error(w, http.StatusBadRequest, "invalid restaurant ID")
		return
	}

	// 3. Decode the submitted restaurant details.
	var input service.UpdateRestaurantInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// 4. Call the service to check ownership and update.
	restaurant, err := h.RestaurantService.UpdateRestaurant(
		uint(restaurantID),
		claims.UserID,
		input,
	)

	if err != nil {
		switch {
		case errors.Is(err, service.ErrRestaurantNotFound):
			utils.Error(w, http.StatusNotFound, err.Error())

		case errors.Is(err, service.ErrRestaurantForbidden):
			utils.Error(w, http.StatusForbidden, err.Error())

		case errors.Is(err, service.ErrInvalidRestaurant):
			utils.Error(w, http.StatusBadRequest, err.Error())

		default:
			utils.Error(w, http.StatusInternalServerError,
				"failed to update restaurant")
		}
		return
	}

	// 5. Return the updated restaurant.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": "Restaurant updated successfully",
		"data":    restaurant,
	})
}
