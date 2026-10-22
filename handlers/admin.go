package handlers

import (
	"Fhub/auth"
	"Fhub/models"
	"Fhub/service"
	"Fhub/utils"
	"encoding/json"
	"errors"
	"fmt"

	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// // request payload
// type setRoleReq struct {
// 	Role string `json:"role"`
// }

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

// SetUserRole handles PATCH /api/admin/users/{id}/role (ADM extra: promote a
// manager so they can register a restaurant).
// Body: {"role": "manager"}
// SetUserRoleHandler handles PATCH /api/admin/users/{id}/role
func (h *AdminHandler) SetUserRoleHandler(w http.ResponseWriter, r *http.Request) {
    // parse id path param using gorilla/mux
    vars := mux.Vars(r)
    idStr, ok := vars["id"]
    if !ok || idStr == "" {
        http.Error(w, "missing id", http.StatusBadRequest)
        return
    }
    id64, err := strconv.ParseUint(idStr, 10, 64)
    if err != nil {
        http.Error(w, "invalid id", http.StatusBadRequest)
        return
    }
    targetID := uint(id64)

    // decode body
    var payload struct   {
		Role string `json:"role"`
	}            
	     
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        http.Error(w, "invalid JSON body", http.StatusBadRequest)
        return
    }

    // map string to models.Role (validate explicitly)
    var role models.Role
    switch payload.Role {
    case string(models.RoleCustomer):
        role = models.RoleCustomer
    case string(models.RoleManager):
        role = models.RoleManager
    case string(models.RoleAdmin):
        role = models.RoleAdmin
    default:
        http.Error(w, "invalid role", http.StatusBadRequest)
        return
    }

    // call service with request context so cancellations propagate
    updatedUser, err := h.AdminService.SetUserRole(r.Context(), targetID, role)
    if err != nil {
        // map service errors to HTTP
        if errors.Is(err, service.ErrUserNotFound) {
            http.Error(w, "user not found", http.StatusNotFound)
            return
        }
        if errors.Is(err, service.ErrInvalidRole) {
            http.Error(w, "invalid role", http.StatusBadRequest)
            return
        }
        utils.Error(w, http.StatusBadRequest, "invalid request body")
        return
    }

    // respond with the updated user
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    if err := json.NewEncoder(w).Encode(updatedUser); err != nil {
        // if encoding fails, there's not much to do. log in real app.
        http.Error(w, "failed to encode response", http.StatusInternalServerError)
    }
}
