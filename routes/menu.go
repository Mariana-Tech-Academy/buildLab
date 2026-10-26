
package routes

import (
	"net/http"

	"Fhub/auth"
	"Fhub/handlers"
	"Fhub/middleware"
	"Fhub/models"
	"Fhub/repository"
	"Fhub/service"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func MenuRoutes(
	router *mux.Router,
	db *gorm.DB,
	tokens *auth.JWTManager,
	blacklist *middleware.TokenBlacklist,
) {
	repo := repository.NewMenuRepository(db)
	menuService := service.NewMenuService(repo)
	menuHandler := handlers.NewMenuHandler(menuService)

	// Public routes
	router.HandleFunc(
		"/restaurants/{id:[0-9]+}/menu",
		menuHandler.List,
	).Methods(http.MethodGet)

	router.HandleFunc(
		"/menu/{id:[0-9]+}",
		menuHandler.Get,
	).Methods(http.MethodGet)

	// Manager and admin routes
	manager := protected(router, tokens, blacklist)

	manager.Use(
		middleware.RequireRole(
			models.RoleManager,
			models.RoleAdmin,
		),
	)

	// Update menu item availability
	manager.HandleFunc(
		"/menu/{id:[0-9]+}/availability",
		menuHandler.SetAvailability,
	).Methods(http.MethodPatch)
}
