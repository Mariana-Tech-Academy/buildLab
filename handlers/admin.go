package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"Fhub/auth"
	"Fhub/models"
	"Fhub/service"
	"Fhub/utils"

	"github.com/gorilla/mux"
)

type AdminHandler struct {
	AdminService service.AdminService
}

func NewAdminHandler(service service.AdminService) *AdminHandler {
	return &AdminHandler{AdminService: service}

}

func (h *AdminHandler) TestAdmin(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		http.Error(w, "UnathoriZed", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)

	fmt.Fprintf(w, "Welcome Admin! Your User ID is %d and your role is %s", claims.UserID, claims.Role)
}

// SetUserRoleHandler handles PATCH /api/admin/users/{id}/role.
// An admin may set the role to customer, manager, or admin.
func (h *AdminHandler) SetUserRoleHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 64)
	if err != nil || id == 0 {
		utils.Error(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var payload struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var role models.Role
	switch payload.Role {
	case string(models.RoleCustomer):
		role = models.RoleCustomer
	case string(models.RoleManager):
		role = models.RoleManager
	case string(models.RoleAdmin):
		role = models.RoleAdmin
	default:
		utils.Error(w, http.StatusUnprocessableEntity, "role must be customer, manager, or admin")
		return
	}

	user, err := h.AdminService.SetUserRole(r.Context(), uint(id), role)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			utils.Error(w, http.StatusNotFound, "user not found")
		case errors.Is(err, service.ErrInvalidRole):
			utils.Error(w, http.StatusUnprocessableEntity, "role must be customer, manager, or admin")
		default:
			log.Printf("set user role: %v", err)
			utils.Error(w, http.StatusInternalServerError, "could not update role")
		}
		return
	}

	utils.Success(w, http.StatusOK, "role updated successfully", newUserResponse(user))
}

// Gets a list of users and allows searching by name or email.
// This uses pagination to control how many users are shown.
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {

	// Gets the search term from the URL.
	search := strings.TrimSpace(r.URL.Query().Get("search"))

	page := 1
	limit := 10

	// This checks if a page number has been provided.
	if value := r.URL.Query().Get("page"); value != "" {
		number, err := strconv.Atoi(value)
		if err != nil || number < 1 {
			utils.Error(w, http.StatusBadRequest, "invalid page number")
			return
		}
		page = number
	}
	// This checks if a limit has been provided.
	if value := r.URL.Query().Get("limit"); value != "" {
		number, err := strconv.Atoi(value)
		if err != nil || number < 1 || number > 50 {
			utils.Error(w, http.StatusBadRequest, "limit must be between 1 and 50")
			return
		}
		limit = number
	}
	// Gets the matching users from the service.
	result, err := h.AdminService.ListUsers(r.Context(), search, page, limit)
	if err != nil {
		if errors.Is(err, service.ErrValidation) {
			utils.Error(w, http.StatusBadRequest, "invalid pagination")
			return
		}

		log.Printf("list users: %v", err)
		utils.Error(w, http.StatusInternalServerError, "could not retrieve users")
		return
	}
	// Converts the users into responses without password information.
	users := make([]UserResponse, 0, len(result.Users))
	for i := range result.Users {
		users = append(users, newUserResponse(&result.Users[i]))
	}
	// Returns the users and pagination information.
	utils.Success(w, http.StatusOK, "users retrieved successfully", map[string]interface{}{
		"users": users,
		"page":  result.Page,
		"limit": result.Limit,
		"total": result.Total,
	})
}

// Gets the details of a user using their ID. The password information is not included in the response.
func (h *AdminHandler) GetUserDetails(w http.ResponseWriter, r *http.Request) {

	// Gets the user ID from the URL.
	id, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 64)
	if err != nil || id == 0 || id > uint64(^uint(0)) {
		utils.Error(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	// Gets the user details from the service.
	user, err := h.AdminService.GetUserDetails(r.Context(), uint(id))
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			utils.Error(w, http.StatusNotFound, "user not found")
			return
		}

		log.Printf("get user details: %v", err)
		utils.Error(w, http.StatusInternalServerError, "could not retrieve user")
		return
	}
	// Returns the user details without password information.
	utils.Success(w, http.StatusOK, "user retrieved successfully", newUserResponse(user))
}
