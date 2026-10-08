package service

import "Fhub/repository"

type RestaurantService struct {
	RestaurantRepo repository.RestaurantRepository
}

func NewRestaurantService(repo repository.RestaurantRepository ) *RestaurantService {
	return &RestaurantService{RestaurantRepo: repo}

}