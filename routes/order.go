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

    // Initialise repositories.
    orderRepo := repository.NewOrderRepository(db)
    restaurantRepo := repository.NewRestaurantRepository(db)

    // Initialise services.
    restaurantService := service.NewRestaurantService(restaurantRepo)
    orderService := service.NewOrderService(orderRepo, restaurantService)

    // Initialise handler.
    orderHandler := handlers.NewOrderHandler(orderService)

    // Customer routes.
    customer := protected(router, tokens, blacklist)
    customer.Use(middleware.RequireRole(models.RoleCustomer))

    customer.HandleFunc(
        "/orders/delivery/validate",
        orderHandler.ValidateDeliveryInfo,
    ).Methods(http.MethodPost)

    // Authenticated routes for future order retrieval.
    authed := protected(router, tokens, blacklist)

    // Manager/Admin routes.
    staff := protected(router, tokens, blacklist)
    staff.Use(middleware.RequireRole(models.RoleManager, models.RoleAdmin))

    // Update order status.
    staff.HandleFunc(
        "/restaurants/{restaurantId:[0-9]+}/orders/{orderId:[0-9]+}/status",
        orderHandler.UpdateOrderStatus,
    ).Methods(http.MethodPatch)

    _ = authed
}
