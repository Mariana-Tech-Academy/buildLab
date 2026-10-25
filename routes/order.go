package routes

import (
	"Fhub/auth"
	"Fhub/handlers"
	"Fhub/middleware"
	"Fhub/models"
	"Fhub/repository"
	"Fhub/service"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"net/http"
)

func OrderRoutes(router *mux.Router, db *gorm.DB, tokens *auth.JWTManager, blacklist *middleware.TokenBlacklist) {
	repo := repository.NewOrderRepository(db)
	orderService := service.NewOrderService(repo)
	orderHandler := handlers.NewOrderHandler(orderService)

	// customer routes: POST /orders (checkout)
	customer := protected(router, tokens, blacklist)
	customer.Use(middleware.RequireRole(models.RoleCustomer))

	// CHK-02: Allow authenticated customers to validate delivery information.
	customer.HandleFunc("/orders/delivery/validate", orderHandler.ValidateDeliveryInfo).Methods(http.MethodPost)

	// authenticated routes: GET /orders, GET /orders/{id}
	// The service scopes results by role: own orders, own restaurant's orders, or all for admin.
	authed := protected(router, tokens, blacklist)

	// manager/admin routes: PATCH /orders/{id}/status
	staff := protected(router, tokens, blacklist)
	staff.Use(middleware.RequireRole(models.RoleManager, models.RoleAdmin))

	_, _, _, _ = orderHandler, customer, authed, staff
}
