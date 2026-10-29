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

// adminOrderResponse is one order as seen by an admin.
// It reuses the restaurant manager's order response (customer, items, totals and delivery details)
// and adds the customer's email and the restaurant.
type adminOrderResponse struct {
	restaurantOrderResponse
	CustomerEmail  string `json:"customer_email"`
	RestaurantID   uint   `json:"restaurant_id"`
	RestaurantName string `json:"restaurant_name"`
}

// newAdminOrderResponse builds the admin's view of an order.
func newAdminOrderResponse(order models.Order) adminOrderResponse {
	return adminOrderResponse{
		restaurantOrderResponse: newRestaurantOrderResponse(order),
		CustomerEmail:           order.Customer.Email,
		RestaurantID:            order.RestaurantID,
		RestaurantName:          order.Restaurant.Name,
	}
}

// adminOrdersResponse is one page of orders for the admin.
type adminOrdersResponse struct {
	Items []adminOrderResponse `json:"items"`
	Page  int                  `json:"page"`
	Limit int                  `json:"limit"`
	Total int64                `json:"total"`
}

// ListAllOrders handles GET /api/admin/orders (ADM-06).
// Optional query parameters: status, search (order number or customer email), page and limit.
func (h *AdminHandler) ListAllOrders(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	search := r.URL.Query().Get("search")

	page, limit, message := restaurantListPage(r)
	if message != "" {
		utils.Error(w, http.StatusBadRequest, message)
		return
	}

	result, err := h.AdminService.ListAllOrders(r.Context(), status, search, page, limit)
	if err != nil {
		if errors.Is(err, service.ErrInvalidOrderStatus) || errors.Is(err, service.ErrInvalidOrderPage) {
			utils.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		log.Printf("admin list orders: %v", err)
		utils.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	items := make([]adminOrderResponse, 0, len(result.Orders))
	for _, order := range result.Orders {
		items = append(items, newAdminOrderResponse(order))
	}

	utils.Success(w, http.StatusOK, "orders retrieved successfully", adminOrdersResponse{
		Items: items,
		Page:  result.Page,
		Limit: result.Limit,
		Total: result.Total,
	})
}

// GetOrderForAdmin handles GET /api/admin/orders/{id} (ADM-06).
func (h *AdminHandler) GetOrderForAdmin(w http.ResponseWriter, r *http.Request) {
	orderID, ok := muxID(w, r)
	if !ok {
		return
	}

	order, err := h.AdminService.GetOrderForAdmin(r.Context(), orderID)
	if err != nil {
		if errors.Is(err, service.ErrOrderNotFound) {
			utils.Error(w, http.StatusNotFound, "order not found")
			return
		}
		log.Printf("admin get order: %v", err)
		utils.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	utils.Success(w, http.StatusOK, "order retrieved successfully", newAdminOrderResponse(*order))
}
