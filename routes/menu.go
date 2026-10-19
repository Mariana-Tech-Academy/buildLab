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
	//Used to create the menu repository to communicate with the database.
	repo := repository.NewMenuRepository(db)

	// We use to create the restaurant repository to retrieve restaurant information.
	restaurantRepo := repository.NewRestaurantRepository(db)

	//Used to create the restaurant service to check restaurant ownership.
	restaurantService := service.NewRestaurantService(restaurantRepo)

	//Used to connect both dependencies to the menu service.
	menuService := service.NewMenuService(repo, restaurantService)

	//Create the menu handler to process HTTO requests.
	menuHandler := handlers.NewMenuHandler(menuService)

	// public routes: GET /restaurants/{id}/menu, GET /menu/{id}

	// manager routes: POST /menu, PUT /menu/{id}, DELETE /menu/{id}
	// The service must also check the item belongs to the calling manager's restaurant.
	manager := protected(router, tokens, blacklist)
	manager.Use(middleware.RequireRole(models.RoleManager, models.RoleAdmin))

	//This route allows authenticated managers and admins to create menu items.
	manager.HandleFunc("/menu", menuHandler.CreateMenuItem).Methods(http.MethodPost)

	_, _ = menuHandler, manager
}
