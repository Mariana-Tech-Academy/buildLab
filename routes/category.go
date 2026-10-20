
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

func CategoryRoutes(
    router *mux.Router,
    db *gorm.DB,
    tokens *auth.JWTManager,
    blacklist *middleware.TokenBlacklist,
) {
    // Initialise repositories
    categoryRepo := repository.NewCategoryRepository(db)
    restaurantRepo := repository.NewRestaurantRepository(db)

    // Initialise services
    restaurantService := service.NewRestaurantService(restaurantRepo)

    categoryService := service.NewCategoryService(
        categoryRepo,
        restaurantService,
    )

    // Initialise handler
    categoryHandler := handlers.NewCategoryHandler(categoryService)

    // Public route: customers can view categories
    router.HandleFunc(
        "/restaurants/{restaurantId}/categories",
        categoryHandler.GetCategories,
    ).Methods(http.MethodGet)

    // Protected routes: managers and admins
    manager := protected(router, tokens, blacklist)
    manager.Use(middleware.RequireRole(
        models.RoleManager,
        models.RoleAdmin,
    ))

    // Create a category
    manager.HandleFunc(
        "/restaurants/{restaurantId}/categories",
        categoryHandler.CreateCategory,
    ).Methods(http.MethodPost)

    // Update and delete routes will be added
    // when their handler methods are implemented.
}
