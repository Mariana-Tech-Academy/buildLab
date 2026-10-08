package routes

import (
	"Fhub/handlers"
	"Fhub/repository"
	"Fhub/service"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func AdminRoutes(router *mux.Router, db *gorm.DB) {
	repo := repository.NewAdminRepository(db)

	adminService := service.NewAdminService(repo)

	adminHandler := handlers.NewAdminHandler(*adminService)

	// define protected routes

	_ = adminHandler
}