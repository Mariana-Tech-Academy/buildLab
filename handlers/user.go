package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode"

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
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	if input.Email == "" || input.Password == "" {
		utils.Error(w, http.StatusBadRequest, "email and password are required")
		return
	}

	result, err := h.UserService.Login(r.Context(), input.Email, input.Password)
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

type registerRequest struct {
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	Email           string `json:"email"`
	Phone           string `json:"phone"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

var ukPhoneRegex = regexp.MustCompile(`^(0|\+44)\d{10}$`)

func (req *registerRequest) validate() map[string]string {
	errs := map[string]string{}

	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	req.Email = strings.TrimSpace(req.Email)
	req.Phone = strings.NewReplacer(" ", "", "-", "").Replace(req.Phone)

	if req.FirstName == "" {
		errs["first_name"] = "first name is required"
	}
	if req.LastName == "" {
		errs["last_name"] = "last name is required"
	}

	if req.Email == "" {
		errs["email"] = "email is required"
	} else if !isValidEmail(req.Email) {
		errs["email"] = "email must be a valid email address"
	}

	if req.Phone == "" {
		errs["phone"] = "phone is required"
	} else if !ukPhoneRegex.MatchString(req.Phone) {
		errs["phone"] = "phone number must be a valid UK number, e.g. 07123456789 or +447123456789"
	}

	if req.Password == "" {
		errs["password"] = "password is required"
	} else if !isStrongPassword(req.Password) {
		errs["password"] = "password must be a minimum of 8 characters and include an uppercase letter, a lowercase letter and a number"
	}

	if req.ConfirmPassword == "" {
		errs["confirm_password"] = "please confirm your password"
	} else if req.ConfirmPassword != req.Password {
		errs["confirm_password"] = "passwords do not match"
	}

	return errs
}

func isValidEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	return err == nil && addr.Address == email
}

func isStrongPassword(password string) bool {
	if len(password) < 8 || len(password) > 72 {
		return false
	}

	var hasUpper, hasLower, hasDigit bool
	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		}
	}
	return hasUpper && hasLower && hasDigit
}

type registerUserResponse struct {
	ID        uint        `json:"id"`
	FirstName string      `json:"first_name"`
	LastName  string      `json:"last_name"`
	Email     string      `json:"email"`
	Phone     string      `json:"phone"`
	Role      models.Role `json:"role"`
	CreatedAt time.Time   `json:"created_at"`
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if errs := req.validate(); len(errs) > 0 {
		utils.JSON(w, http.StatusUnprocessableEntity, utils.Response{
			Success: false,
			Message: "validation failed",
			Data:    errs,
		})
		return
	}

	user, err := h.UserService.Register(r.Context(), service.RegisterInput{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Phone:     req.Phone,
		Password:  req.Password,
	})
	if errors.Is(err, service.ErrEmailTaken) {
		utils.Error(w, http.StatusConflict, "email is already taken")
		return
	}
	if err != nil {
		log.Printf("register user: %v", err)
		utils.Error(w, http.StatusInternalServerError, "could not create account")
		return
	}

	utils.Success(w, http.StatusCreated, "account created successfully", registerUserResponse{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Phone:     user.Phone,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	})
}
