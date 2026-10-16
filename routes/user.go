package routes

import (
	"net/http"

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

	// public routes: POST /auth/register, POST /auth/login
	public := router.PathPrefix("/auth").Subrouter()
	public.HandleFunc("/register", userHandler.Register).Methods(http.MethodPost)
	public.HandleFunc("/login", userHandler.Login).Methods(http.MethodPost)

	// protected routes: POST /auth/logout, GET /auth/me, PUT /users/me, PUT /users/me/password
	authed := protected(router, tokens, blacklist)
	authed.HandleFunc("/auth/me", userHandler.GetMe).Methods(http.MethodGet)
	authed.HandleFunc("/auth/logout", userHandler.Logout).Methods(http.MethodPost)
	authed.HandleFunc("/users/me", userHandler.UpdateProfile).Methods(http.MethodPut)

	_ = public
}
