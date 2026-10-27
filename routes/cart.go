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

func CartRoutes(router *mux.Router, db *gorm.DB, tokens *auth.JWTManager, blacklist *middleware.TokenBlacklist) {
	// Create the repository responsible for shopping cart database operations.
	cartRepo := repository.NewCartRepository(db)

	// Create the repository responsible for retrieving restaurant menu items.
	menuRepo := repository.NewMenuRepository(db)
	restaurantRepo := repository.NewRestaurantRepository(db)
	restaurantService := service.NewRestaurantService(restaurantRepo)

	// Give the cart service access to cart, menu, and restaurant data.
	cartService := service.NewCartService(cartRepo, menuRepo, restaurantService)

	// Connect the cart service to the HTTP handler.
	cartHandler := handlers.NewCartHandler(cartService)

	// customer routes: GET /cart, POST /cart/items, PUT /cart/items/{id}, DELETE /cart/items/{id}, DELETE /cart
	customer := protected(router, tokens, blacklist)
	customer.Use(middleware.RequireRole(models.RoleCustomer))
	customer.HandleFunc("/cart/items/{id}", cartHandler.UpdateQuantity).Methods(http.MethodPut)

	// CART-01: Allow authenticated customers to add menu items to their cart.
	// POST means the customer is sending data to the server.
	// /cart/items is the endpoint Postman will call.
	// AddItemToCart is the handler that processes the request.
	customer.HandleFunc("/cart/items", cartHandler.AddItemToCart).Methods(http.MethodPost)
}
