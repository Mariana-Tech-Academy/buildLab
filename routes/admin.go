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

func AdminRoutes(router *mux.Router, db *gorm.DB, tokens *auth.JWTManager, blacklist *middleware.TokenBlacklist) {
	adminRepo := repository.NewAdminRepository(db)
	userRepo := repository.NewUserRepository(db)
	adminService := service.NewAdminService(adminRepo, userRepo)
	adminHandler := handlers.NewAdminHandler(adminService)

	// admin routes: GET /admin/users, GET /admin/restaurants, GET /admin/orders, GET /admin/statistics, ...
	admin := router.PathPrefix("/admin").Subrouter()
	admin.Use(middleware.Authenticate(tokens, blacklist), middleware.RequireRole(models.RoleAdmin))

	admin.HandleFunc("/test", adminHandler.TestAdmin).Methods(http.MethodGet)
	admin.HandleFunc("/users/{id}/role", adminHandler.SetUserRoleHandler).Methods(http.MethodPatch)

	//ADM-06 admins can list, search and filter all orders and view an order in full
	admin.HandleFunc("/orders", adminHandler.ListAllOrders).Methods(http.MethodGet)
	admin.HandleFunc("/orders/{id}", adminHandler.GetOrderForAdmin).Methods(http.MethodGet)
}
