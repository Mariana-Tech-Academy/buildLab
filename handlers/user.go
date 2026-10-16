package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"Fhub/auth"
	"Fhub/middleware"
	"Fhub/models"
	"Fhub/service"
	"Fhub/utils"
)

type UserHandler struct {
	UserService service.UserService
	Tokens      *auth.JWTManager
	Blacklist   *middleware.TokenBlacklist
}

func NewUserHandler(userService service.UserService, tokens *auth.JWTManager, blacklist *middleware.TokenBlacklist) *UserHandler {
	return &UserHandler{
		UserService: userService,
		Tokens:      tokens,
		Blacklist:   blacklist,
	}
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

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	tokenString, ok := strings.CutPrefix(
		r.Header.Get("Authorization"),
		"Bearer ",
	)
	if !ok || tokenString == "" {
		utils.Error(
			w,
			http.StatusUnauthorized,
			"missing or malformed authorization header",
		)
		return
	}

	claims, err := h.Tokens.Verify(tokenString)
	if err != nil {
		utils.Error(
			w,
			http.StatusUnauthorized,
			"invalid or expired token",
		)
		return
	}

	if claims.ExpiresAt == nil {
		utils.Error(w, http.StatusUnauthorized, "token expiry missing")
		return
	}

	h.Blacklist.Revoke(tokenString, claims.ExpiresAt.Time)

	utils.Success(w, http.StatusOK, "Logged out successfully", nil)
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
