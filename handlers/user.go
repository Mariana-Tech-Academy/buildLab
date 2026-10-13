
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
	UserService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
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

	user, err := h.UserService.UpdateProfile(
		r.Context(),
		claims.UserID,
		req.FirstName,
		req.LastName,
		req.Phone,
	)

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

// Change Password Request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// Change Password Handler
func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {

	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req ChangePasswordRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err = h.UserService.ChangePassword(
		r.Context(),
		claims.UserID,
		req.CurrentPassword,
		req.NewPassword,
	)

	if err != nil {
		switch {

		case errors.Is(err, service.ErrIncorrectPassword):
			utils.Error(w, http.StatusUnauthorized, "Incorrect current password")

		case errors.Is(err, service.ErrValidation):
			utils.Error(w, http.StatusBadRequest, err.Error())

		case errors.Is(err, service.ErrUserNotFound):
			utils.Error(w, http.StatusNotFound, "User not found")

		default:
			log.Printf("change password: %v", err)
			utils.Error(w, http.StatusInternalServerError, "Could not change password")
		}
		return
	}

	utils.Success(w, http.StatusOK, "Password changed successfully", nil)
}
