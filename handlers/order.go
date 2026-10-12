package handlers

import "Fhub/service"

type OrderHandler struct {
	OrderService *service.OrderService
}

func NewOrderHandler(service *service.OrderService) *OrderHandler {
	return &OrderHandler{OrderService: service}

}
