package service

import (
	"Fhub/models"
	"Fhub/repository"
	"strings"
)



type CreateRestaurantInput struct {
	ManagerID   uint
	Name        string
	Description string
	Address     string
	Phone       string
	Email       string
	OpeningTime string
	ClosingTime string
}

type RestaurantService interface {
	Create(input CreateRestaurantInput) (*models.Restaurant, error)

}

type restaurantService struct {
	restaurantRepo repository.RestaurantRepository
}

func NewRestaurantService(repo repository.RestaurantRepository) RestaurantService {
	return &restaurantService{restaurantRepo: repo}
}

func (s *restaurantService) Create(input CreateRestaurantInput) (*models.Restaurant, error) {  // Create a new restaurant with the provided input and return the created restaurant or an error
 restaurant := &models.Restaurant{  // Create a new Restaurant model instance with the provided input
  ManagerID:   input.ManagerID,
  Name:        input.Name,
  Description: input.Description,
  Address:     input.Address,
  Phone:       input.Phone,
  Email:       strings.ToLower(input.Email),
  OpeningTime: input.OpeningTime,
  ClosingTime: input.ClosingTime,
  Status:      models.RestaurantStatusPending,    // Set the initial status to pending approval
 }

 if err := s.restaurantRepo.Create(restaurant); err != nil {
  return nil, err
 }

 return restaurant, nil
}




















