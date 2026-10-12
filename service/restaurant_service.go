package service

import "Fhub/repository"

type RestaurantService struct {
	restaurantRepo repository.RestaurantRepository
}

func NewRestaurantService(repo repository.RestaurantRepository) *RestaurantService {
	return &RestaurantService{restaurantRepo: repo}
}
