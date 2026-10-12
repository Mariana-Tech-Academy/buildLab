package routes

import (
	"Fhub/auth"
	"Fhub/handlers"
	"Fhub/middleware"
	"Fhub/repository"
	"Fhub/service"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func UserRoutes(router *mux.Router, db *gorm.DB, tokens *auth.JWTManager, blacklist *middleware.TokenBlacklist) {
	repo := repository.NewUserRepository(db)
	userService := service.NewUserService(repo, tokens)
	userHandler := handlers.NewUserHandler(userService, tokens, blacklist)

	// Protected routes: GET /auth/me, POST /auth/logout
	authed := protected(router, tokens, blacklist)
	authed.HandleFunc("/auth/me", userHandler.GetMe).Methods("GET")
	authed.HandleFunc("/auth/logout", userHandler.Logout).Methods("POST")
}
