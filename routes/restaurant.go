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

func RestaurantRoutes(router *mux.Router, db *gorm.DB, tokens *auth.JWTManager) {
	repo := repository.NewRestaurantRepository(db)
	restaurantService := service.NewRestaurantService(repo)
	restaurantHandler := handlers.NewRestaurantHandler(restaurantService)

	// public routes: GET /restaurants, GET /restaurants/{id}

	// manager routes: POST /restaurants, PUT /restaurants/{id}, PATCH /restaurants/{id}/status
	// The service must also check the restaurant belongs to the calling manager.
	manager := protected(router, tokens)
	manager.Use(middleware.RequireRole(models.RoleManager, models.RoleAdmin))

	manager.HandleFunc(
		"/restaurants/{id}/image",
		restaurantHandler.UploadRestaurantImage,
	).Methods(http.MethodPut)

	// Update restaurant details and opening hours
	manager.HandleFunc(
		"/restaurants/{id}",
		restaurantHandler.UpdateRestaurant,
	).Methods(http.MethodPut)
}
