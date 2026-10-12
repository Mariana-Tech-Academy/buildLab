package handlers

import (
	"errors"
	"net/http"
	"strings"

	"Fhub/auth"
	"Fhub/middleware"
	"Fhub/models"
	"Fhub/service"
	"Fhub/utils"

	"gorm.io/gorm"
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

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := h.UserService.GetUserByID(int(claims.UserID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Error(w, http.StatusNotFound, "User not found")
			return
		}

		utils.Error(w, http.StatusInternalServerError, "could not retrieve user")
		return
	}

	profile := struct {
		ID        uint        `json:"id"`
		FirstName string      `json:"first_name"`
		LastName  string      `json:"last_name"`
		Email     string      `json:"email"`
		Phone     string      `json:"phone"`
		Role      models.Role `json:"role"`
	}{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Phone:     user.Phone,
		Role:      user.Role,
	}

	utils.Success(
		w,
		http.StatusOK,
		"Current user retrieved successfully",
		profile,
	)
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