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

func MenuRoutes(router *mux.Router, db *gorm.DB, tokens *auth.JWTManager, blacklist *middleware.TokenBlacklist) {
	repo := repository.NewMenuRepository(db)
	restaurantRepo := repository.NewRestaurantRepository(db)
	restaurants := service.NewRestaurantService(restaurantRepo)
	menuService := service.NewMenuService(repo, restaurants)
	menuHandler := handlers.NewMenuHandler(menuService)

	// Public menu routes will be added separately.

	// Manager and admin routes: PUT /menu/{id}, DELETE /menu/{id}.
	// The service must also check the item belongs to the calling manager's restaurant.
	manager := protected(router, tokens, blacklist)
	manager.Use(middleware.RequireRole(models.RoleManager, models.RoleAdmin))

	// Send the dit requests to the update handler
	manager.HandleFunc("/menu/{id}", menuHandler.UpdateMenuItem).Methods(http.MethodPut)

	// Send delete requests to the delete handler
	manager.HandleFunc("/menu/{id}", menuHandler.DeleteMenuItem).Methods(http.MethodDelete)
}
