package service

import (
    "errors"
    "strings"

    "Fhub/models"
    "Fhub/repository"
)


var ErrInvalidDelivery = errors.New(
    "delivery name, phone and address are required",
)


type OrderService interface { 
Checkout(uint, repository.DeliverySnapshot) (*models.Order, error)
List(uint) ([]models.Order, error)
Get(uint, uint) (*models.Order, error)

}

type orderService struct {
	orderRepo repository.OrderRepository
}

func NewOrderService(repo repository.OrderRepository) OrderService {
	return &orderService{orderRepo: repo}
}


func (s *orderService) Checkout(
    customerID uint,
    delivery repository.DeliverySnapshot,
) (*models.Order, error) {

    if strings.TrimSpace(delivery.Name) == "" ||
        strings.TrimSpace(delivery.Phone) == "" ||
        strings.TrimSpace(delivery.Address) == "" {
        return nil, ErrInvalidDelivery
    }

    return s.orderRepo.Checkout(customerID, delivery)
}


func (s *orderService) List(
    customerID uint,
) ([]models.Order, error) {
    return s.orderRepo.List(customerID)
}

func (s *orderService) Get(
    customerID uint,
    orderID uint,
) (*models.Order, error) {
    return s.orderRepo.Get(customerID, orderID)
}

