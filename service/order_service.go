package service

import "Fhub/repository"

type OrderService struct {
	OrderRepo repository.OrderRepository
}

func NewOrderService(repo repository.OrderRepository ) *OrderService {
	return &OrderService{OrderRepo: repo}

}