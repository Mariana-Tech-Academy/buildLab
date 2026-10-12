package service

import "Fhub/repository"

type CartService struct {
	cartRepo repository.CartRepository
}

func NewCartService(repo repository.CartRepository) *CartService {
	return &CartService{cartRepo: repo}
}
