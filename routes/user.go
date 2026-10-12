package routes

import (
	"Fhub/auth"
	"Fhub/handlers"
	"Fhub/repository"
	"Fhub/service"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func UserRoutes(router *mux.Router, db *gorm.DB, tokens *auth.JWTManager) {
	repo := repository.NewUserRepository(db)
	userService := service.NewUserService(repo, tokens)
	userHandler := handlers.NewUserHandler(userService)

	// public routes: POST /auth/register, POST /auth/login
	public := router.PathPrefix("/auth").Subrouter()

	// protected routes: POST /auth/logout, GET /auth/me, PUT /users/me, PUT /users/me/password
	authed := protected(router, tokens)
	authed.HandleFunc("/auth/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("You are authenticated!"))

	}).Methods(http.MethodGet)

	_, _, _ = userHandler, public, authed
}
