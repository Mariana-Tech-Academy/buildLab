package handlers

import (
	"Fhub/service"
	"Fhub/utils"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type UserHandler struct {
	UserService service.UserService
}

func NewUserHandler(service service.UserService) *UserHandler {
	return &UserHandler{UserService: service}

}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	input.Email = strings.TrimSpace(input.Email)

	if input.Email == "" || input.Password == "" {
		utils.Error(w, http.StatusBadRequest, "email and password are required")
		return
	}

	result, err := h.UserService.Login(input.Email, input.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			utils.Error(w, http.StatusUnauthorized, "Invalid email or password")
		case errors.Is(err, service.ErrAccountSuspended):
			utils.Error(w, http.StatusForbidden, "Your account is suspended")
		default:
			utils.Error(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}
	utils.Success(w, http.StatusOK, "Login successful", result)
}
