package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"Fhub/auth"
	"Fhub/models"
	"Fhub/service"
	"Fhub/utils"
)

type UserHandler struct {
	UserService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{UserService: userService}
}

type UserResponse struct {
	ID        uint        `json:"id"`
	FirstName string      `json:"first_name"`
	LastName  string      `json:"last_name"`
	Email     string      `json:"email"`
	Phone     string      `json:"phone"`
	Role      models.Role `json:"role"`
}

func newUserResponse(user *models.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Phone:     user.Phone,
		Role:      user.Role,
	}
}

type updateProfileRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.UserService.GetUserByID(r.Context(), claims.UserID)
	if err != nil {
		writeUserError(w, err)
		return
	}

	utils.Success(w, http.StatusOK, "Current user retrieved successfully", newUserResponse(user))
}

// UpdateProfile only changes first name, last name and phone. Any other fields in the
// body (e.g. email) are ignored.
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req updateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.UserService.UpdateProfile(r.Context(), claims.UserID, req.FirstName, req.LastName, req.Phone)
	if err != nil {
		writeUserError(w, err)
		return
	}

	utils.Success(w, http.StatusOK, "Profile updated successfully", newUserResponse(user))
}

func writeUserError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrValidation):
		utils.Error(w, http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, service.ErrUserNotFound):
		utils.Error(w, http.StatusNotFound, "user not found")
	default:
		log.Printf("user handler: %v", err)
		utils.Error(w, http.StatusInternalServerError, "something went wrong")
	}
}
