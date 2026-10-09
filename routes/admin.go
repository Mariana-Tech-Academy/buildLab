package routes

import (
	"Fhub/auth"
	"Fhub/handlers"
	"Fhub/middleware"
	"Fhub/models"
	"Fhub/repository"
	"Fhub/service"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func AdminRoutes(router *mux.Router, db *gorm.DB, tokens *auth.JWTManager) {
	repo := repository.NewAdminRepository(db)
	adminService := service.NewAdminService(repo)
	adminHandler := handlers.NewAdminHandler(adminService)

	// admin routes: GET /admin/users, GET /admin/restaurants, GET /admin/orders, GET /admin/statistics, ...
	admin := router.PathPrefix("/admin").Subrouter()
	admin.Use(middleware.Authenticate(tokens), middleware.RequireRole(models.RoleAdmin))

	_, _ = adminHandler, admin
}
