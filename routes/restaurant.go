package routes

import (
	"Fhub/auth"
	"Fhub/handlers"
	"Fhub/middleware"
	"Fhub/models"
	"Fhub/repository"
	"Fhub/service"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func RestaurantRoutes(router *mux.Router, db *gorm.DB, tokens *auth.JWTManager, blacklist *middleware.TokenBlacklist) {
	repo := repository.NewRestaurantRepository(db)
	restaurantService := service.NewRestaurantService(repo)
	restaurantHandler := handlers.NewRestaurantHandler(restaurantService)

	// public routes: GET /restaurants, GET /restaurants/{id}
	// SearchActive lists restaurants, and filters them when q is present.
	router.HandleFunc("/restaurants", restaurantHandler.SearchActive).Methods(http.MethodGet)

	// manager routes: POST /restaurants, PUT /restaurants/{id}, PATCH /restaurants/{id}/status
	manager := protected(router, tokens, blacklist)
	manager.Use(middleware.RequireRole(models.RoleManager, models.RoleAdmin))
	manager.HandleFunc("/restaurants", restaurantHandler.Create).Methods("POST")

	manager.HandleFunc(
		"/restaurants/{id}/image",
		restaurantHandler.UploadRestaurantImage,
	).Methods(http.MethodPut)

	// Update restaurant details and opening hours
	manager.HandleFunc(
		"/restaurants/{id}",
		restaurantHandler.UpdateRestaurant,
	).Methods(http.MethodPut)

	// New route for activating/deactivating a restaurant
	manager.HandleFunc("/restaurants/{id}/status", restaurantHandler.UpdateStatus).Methods(http.MethodPatch)
}
