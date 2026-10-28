package routes

import (
	"net/http"

	"Fhub/auth"
	"Fhub/handlers"
	"Fhub/middleware"
	"Fhub/models"
	"Fhub/repository"
	"Fhub/service"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func OrderRoutes(router *mux.Router, db *gorm.DB, tokens *auth.JWTManager, blacklist *middleware.TokenBlacklist) {
	// Create repositories for orders and restaurants.
	repo := repository.NewOrderRepository(db)
	restaurantRepo := repository.NewRestaurantRepository(db)

	// Pass both repositories into the order service.
	orderService := service.NewOrderService(repo, restaurantRepo)
	orderHandler := handlers.NewOrderHandler(orderService)

	// customer routes: POST /orders (checkout)
	customer := protected(router, tokens, blacklist)
	customer.Use(middleware.RequireRole(models.RoleCustomer))
	customer.HandleFunc("/orders/checkout/summary", orderHandler.GetCheckoutSummary).Methods(http.MethodGet)

	// CHK-02: Allow authenticated customers to validate delivery information.
	customer.HandleFunc("/orders/delivery/validate", orderHandler.ValidateDeliveryInfo).Methods(http.MethodPost)

	// authenticated routes: GET /orders, GET /orders/{id}
	// The service scopes results by role: own orders, own restaurant's orders, or all for admin.
	authed := protected(router, tokens, blacklist)

	// manager/admin routes: PATCH /orders/{id}/status
	staff := protected(router, tokens, blacklist)
	staff.Use(middleware.RequireRole(models.RoleManager, models.RoleAdmin))

	// ORD-04: Allow managers to view orders belonging to their restaurant.
	staff.HandleFunc("/restaurants/{id}/orders", orderHandler.GetRestaurantOrders).Methods(http.MethodGet)

	_, _ = authed, staff
}
