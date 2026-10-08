package routes


import (
	"Fhub/handlers"
	"Fhub/repository"
	"Fhub/service"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func UserRoutes(router *mux.Router, db *gorm.DB) {
	repo := repository.NewUserRepository(db)

	userService := service.NewUserService(repo)

	userHandler := handlers.NewUserHandler(*userService)

	// define public routes

	// define protected routes

	_ = userHandler
}