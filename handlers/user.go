package handlers

import (
	"net/http"

	"Fhub/auth"
	"Fhub/service"
	"Fhub/utils"
)

type UserHandler struct {
	UserService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{UserService: userService}
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := h.UserService.GetUserByID(int(claims.UserID))
	if err != nil {
		utils.Error(w, http.StatusNotFound, "User not found")
		return
	}

	profile := struct {
		ID        uint   `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Phone     string `json:"phone"`
		Role      string `json:"role"`
	}{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Phone:     user.Phone,
		Role:      string(user.Role),
	}

	utils.Success(w, http.StatusOK, "Current user retrieved successfully", profile)
}
