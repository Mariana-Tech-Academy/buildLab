package routes

import (
	"Fhub/handlers"
	"Fhub/repository"
	"Fhub/service"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func CategoryRoutes(router *mux.Router, db *gorm.DB) {
	repo := repository.NewCategoryRepository(db)

	categoryService := service.NewCategoryService(repo)

	categoryHandler := handlers.NewCategoryHandler(*categoryService)

	// define public routes

	// define protected routes

	_ = categoryHandler
}