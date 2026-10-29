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

	checkoutService := service.NewCheckoutService(repository.NewCheckoutRepository(db))
	checkoutHandler := handlers.NewCheckoutHandler(checkoutService, orderService)

	// customer routes: POST /orders (checkout)
	customer := protected(router, tokens, blacklist)
	customer.Use(middleware.RequireRole(models.RoleCustomer))
	customer.HandleFunc("/orders/checkout/summary", orderHandler.GetCheckoutSummary).Methods(http.MethodGet)

	// CHK-02: Allow authenticated customers to validate delivery information.
	customer.HandleFunc("/orders/delivery/validate", orderHandler.ValidateDeliveryInfo).Methods(http.MethodPost)

	// Connect checkout requests to the checkout handler.
	customer.HandleFunc("/orders", checkoutHandler.PlaceOrder).Methods(http.MethodPost)
	customer.HandleFunc("/orders", orderHandler.ViewOrder).Methods(http.MethodGet)
	customer.HandleFunc("/orders/{id}", orderHandler.GetOrderByID).Methods(http.MethodGet)

	authed := protected(router, tokens, blacklist)
	// ORD-06: customers, managers and admins can cancel an order.
	authed.HandleFunc("/orders/{id}/cancel", orderHandler.CancelOrder).Methods(http.MethodPatch)

	// manager/admin routes: PATCH /orders/{id}/status
	staff := protected(router, tokens, blacklist)
	staff.Use(middleware.RequireRole(models.RoleManager, models.RoleAdmin))

	// ORD-04: Allow managers to view orders belonging to their restaurant.
	staff.HandleFunc("/restaurants/{id}/orders", orderHandler.GetRestaurantOrders).Methods(http.MethodGet)
	staff.HandleFunc("/restaurants/{id}/orders/{orderId}/status", orderHandler.UpdateOrderStatus).Methods(http.MethodPatch)

	_, _ = authed, staff
}
