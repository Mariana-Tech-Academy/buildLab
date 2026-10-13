package handlers

import (
	"Fhub/models"
	"Fhub/service"
	"Fhub/utils"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode"
)

type UserHandler struct {
	UserService service.UserService
}

func NewUserHandler(service service.UserService) *UserHandler {
	return &UserHandler{UserService: service}

}

type registerRequest struct {
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	Email           string `json:"email"`
	Phone           string `json:"phone"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

var ukPhoneRegex = regexp.MustCompile(`^(0|\+44\d{10}$)`)

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

type userResponse struct {
	ID        uint        `json:"id"`
	FirstName string      `json:"first_name"`
	LastName  string      `json:"last_name"`
	Email     string      `json:"email"`
	Phone     string      `json:"phone"`
	Role      models.Role `json:"role"`
	CreatedAt time.Time   `json:"created_at"`
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
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

	user, err := h.UserService.Register(service.RegisterInput{
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
		log.Printf("register user: %V", err)
		utils.Error(w, http.StatusInternalServerError, "could not create account")
		return
	}

	utils.Success(w, http.StatusCreated, "account created successfully", userResponse{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Phone:     user.Phone,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	})
}
