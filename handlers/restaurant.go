package handlers

import (
	"Fhub/auth"
	"Fhub/models"
	"Fhub/service"
	"Fhub/utils"
	"encoding/json"
	"log"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"time"
)

type RestaurantHandler struct {
	RestaurantService service.RestaurantService
}

func NewRestaurantHandler(service service.RestaurantService) *RestaurantHandler {
	return &RestaurantHandler{RestaurantService: service}

}

type createRestaurantRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Address     string `json:"address"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	OpeningTime string `json:"opening_time"`
	ClosingTime string `json:"closing_time"`
}

type restaurantResponse struct {
	ID          uint                    `json:"id"`
	ManagerID   uint                    `json:"manager_id"`
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	Address     string                  `json:"address"`
	Phone       string                  `json:"phone"`
	Email       string                  `json:"email"`
	OpeningTime string                  `json:"opening_time"`
	ClosingTime string                  `json:"closing_time"`
	Status      models.RestaurantStatus `json:"status"`
	CreatedAt   time.Time               `json:"created_at"`
}

var restaurantPhoneRegex = regexp.MustCompile(`^(0|\+44)\d{10}$`)

func (req *createRestaurantRequest) validate() map[string]string { // Validate the request fields and return a map of field names to error messages
	errs := map[string]string{}

	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	req.Address = strings.TrimSpace(req.Address)
	req.Email = strings.TrimSpace(req.Email)
	req.Phone = strings.NewReplacer(" ", "", "-", "").Replace(req.Phone)
	req.OpeningTime = strings.TrimSpace(req.OpeningTime)
	req.ClosingTime = strings.TrimSpace(req.ClosingTime)
	if req.Name == "" {
		errs["name"] = "name is required"
	}

	if req.Description == "" {
		errs["description"] = "description is required"
	}

	if req.Address == "" {
		errs["address"] = "address is required"
	}

	if req.Phone == "" {
		errs["phone"] = "phone is required"
	} else if !restaurantPhoneRegex.MatchString(req.Phone) {
		errs["phone"] = "phone must be a valid UK number, e.g. 01131234567 or +441131234567"
	}
	if req.Email == "" {
		errs["email"] = "email is required"
	} else if addr, err := mail.ParseAddress(req.Email); err != nil || addr.Address != req.Email {
		errs["email"] = "email must be a valid email address"
	}

	opening, openingErr := time.Parse("15:04", req.OpeningTime) //
	if req.OpeningTime == "" {
		errs["opening_time"] = "opening time is required"
	} else if openingErr != nil {
		errs["opening_time"] = "opening time must be in HH:MM format, e.g. 09:00"
	} else {
		req.OpeningTime = opening.Format("15:04")
	}
	closing, closingErr := time.Parse("15:04", req.ClosingTime)
	if req.ClosingTime == "" {
		errs["closing_time"] = "closing time is required"
	} else if closingErr != nil {
		errs["closing_time"] = "closing time must be in HH:MM format, e.g. 22:00"
	} else {
		req.ClosingTime = closing.Format("15:04")
	}

	if openingErr == nil && closingErr == nil && !closing.After(opening) { // Check if closing time is after opening time
		errs["closing_time"] = "closing time must be after opening time"
	}

	return errs
}

func (h *RestaurantHandler) Create(w http.ResponseWriter, r *http.Request) { // Create a new restaurant
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	var req createRestaurantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if errs := req.validate(); len(errs) > 0 {
		utils.JSON(w, http.StatusBadRequest, utils.Response{
			Success: false,
			Message: "validation failed",
			Data:    errs,
		})
		return
	}
	restaurant, err := h.RestaurantService.Create(service.CreateRestaurantInput{ // Call the service layer to create the restaurant
		ManagerID:   claims.UserID,
		Name:        req.Name,
		Description: req.Description,
		Address:     req.Address,
		Phone:       req.Phone,
		Email:       req.Email,
		OpeningTime: req.OpeningTime,
		ClosingTime: req.ClosingTime,
	})
	if err != nil {
		log.Printf("create restaurant: %v", err) // Log the error for debugging purposes
		utils.Error(w, http.StatusInternalServerError, "could not create restaurant")
		return
	}
	utils.Success(w, http.StatusCreated, "restaurant created and pending approval", restaurantResponse{
		ID:          restaurant.ID,
		ManagerID:   restaurant.ManagerID,
		Name:        restaurant.Name,
		Description: restaurant.Description,
		Address:     restaurant.Address,
		Phone:       restaurant.Phone,
		Email:       restaurant.Email,
		OpeningTime: restaurant.OpeningTime,
		ClosingTime: restaurant.ClosingTime,
		Status:      restaurant.Status,
		CreatedAt:   restaurant.CreatedAt,
	})
}
