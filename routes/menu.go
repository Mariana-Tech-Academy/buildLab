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

func MenuRoutes(router *mux.Router, db *gorm.DB, tokens *auth.JWTManager, blacklist *middleware.TokenBlacklist) {
	repo := repository.NewMenuRepository(db)
	restaurantRepo := repository.NewRestaurantRepository(db)
	restaurantService := service.NewRestaurantService(restaurantRepo)
	categoryRepo := repository.NewCategoryRepository(db)
	menuService := service.NewMenuService(repo, categoryRepo, restaurantService)
	menuHandler := handlers.NewMenuHandler(menuService)

	// Public routes: customers can view menu item details.
    router.HandleFunc("/menu/{id}", menuHandler.GetMenuItemByID).Methods(http.MethodGet)
	
	// manager routes: POST /menu, PUT /menu/{id}, DELETE /menu/{id}
	// The service must also check the item belongs to the calling manager's restaurant.
	manager := protected(router, tokens, blacklist)
	manager.Use(middleware.RequireRole(models.RoleManager, models.RoleAdmin))

	manager.HandleFunc("/menu", menuHandler.CreateMenuItem).Methods(http.MethodPost)
	manager.HandleFunc("/menu/{id}", menuHandler.UpdateMenuItem).Methods(http.MethodPut)
	manager.HandleFunc("/menu/{id}", menuHandler.DeleteMenuItem).Methods(http.MethodDelete)
}
