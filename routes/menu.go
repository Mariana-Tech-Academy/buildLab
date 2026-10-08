package routes
import (
	"Fhub/handlers"
	"Fhub/repository"
	"Fhub/service"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func MenuRoutes(router *mux.Router, db *gorm.DB) {
	repo := repository.NewMenuRepository(db)

	menuService := service.NewMenuService(repo)

	menuHandler := handlers.NewMenuHandler(*menuService)

	// define public routes

	// define protected routes

	_ = menuHandler
}