package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

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
