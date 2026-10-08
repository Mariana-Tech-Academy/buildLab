package routes

import (
	"Fhub/handlers"
	"Fhub/repository"
	"Fhub/service"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func CuisineRoutes(router *mux.Router, db *gorm.DB) {
	repo := repository.NewCuisineRepository(db)

	cuisineService := service.NewCuisineService(repo)

	cuisineHandler := handlers.NewCuisineHandler(*cuisineService)

	// define public routes

	// define protected routes

	_ = cuisineHandler
}