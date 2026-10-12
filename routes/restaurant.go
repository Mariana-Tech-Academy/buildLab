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
)

func RestaurantRoutes(router *mux.Router, db *gorm.DB, tokens *auth.JWTManager, blacklist *middleware.TokenBlacklist) {
	repo := repository.NewRestaurantRepository(db)
	restaurantService := service.NewRestaurantService(repo)
	restaurantHandler := handlers.NewRestaurantHandler(restaurantService)

	// public routes: GET /restaurants, GET /restaurants/{id}

	// manager routes: POST /restaurants, PUT /restaurants/{id}, PATCH /restaurants/{id}/status
	// The service must also check the restaurant belongs to the calling manager.
	manager := protected(router, tokens, blacklist)
	manager.Use(middleware.RequireRole(models.RoleManager, models.RoleAdmin))

	_, _ = restaurantHandler, manager
}
