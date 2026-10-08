package routes

import (
	"Fhub/handlers"
	"Fhub/repository"
	"Fhub/service"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func OrderRoutes(router *mux.Router, db *gorm.DB) {
	repo := repository.NewOrderRepository(db)

	orderService := service.NewOrderService(repo)

	orderHandler := handlers.NewOrderHandler(*orderService)

	// define public routes

	// define protected routes

	_ = orderHandler
}