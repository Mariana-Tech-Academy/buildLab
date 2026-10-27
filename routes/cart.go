
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

func CartRoutes(
    router *mux.Router,
    db *gorm.DB,
    tokens *auth.JWTManager,
    blacklist *middleware.TokenBlacklist,
) {
    repo := repository.NewCartRepository(db)
    cartService := service.NewCartService(repo)
    cartHandler := handlers.NewCartHandler(cartService)

    // Customer routes
    customer := protected(router, tokens, blacklist)
    customer.Use(middleware.RequireRole(models.RoleCustomer))

    // CART-04: Remove a selected item from the cart
    customer.HandleFunc(
        "/cart/items/{id:[0-9]+}",
        cartHandler.RemoveItem,
    ).Methods(http.MethodDelete)

    // CART-04: Clear all items from the cart
    customer.HandleFunc(
        "/cart",
        cartHandler.Clear,
    ).Methods(http.MethodDelete)
}
