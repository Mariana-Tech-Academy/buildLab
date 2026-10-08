package routes

import (
	"Fhub/handlers"
	"Fhub/repository"
	"Fhub/service"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func RestaurantRoutes(router *mux.Router, db *gorm.DB) {
	repo := repository.NewRestaurantRepository(db)

	restaurantService := service.NewRestaurantService(repo)

	restaurantHandler := handlers.NewRestaurantHandler(*restaurantService)

	// define public routes

	// define protected routes

	_ = restaurantHandler
}