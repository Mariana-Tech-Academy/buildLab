package routes

import (
	"Fhub/handlers"
	"Fhub/repository"
	"Fhub/service"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func CartRoutes(router *mux.Router, db *gorm.DB) {
	repo := repository.NewCartRepository(db)

	cartService := service.NewCartService(repo)

	cartHandler := handlers.NewCartHandler(*cartService)

	// define public routes

	// define protected routes

	_ = cartHandler
}