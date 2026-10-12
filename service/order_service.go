package service

import "Fhub/repository"

type OrderService struct {
	orderRepo repository.OrderRepository
}

func NewOrderService(repo repository.OrderRepository) *OrderService {
	return &OrderService{orderRepo: repo}
}
