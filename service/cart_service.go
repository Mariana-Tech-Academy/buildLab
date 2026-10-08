package service

import "Fhub/repository"

type CartService struct {
	CartRepo repository.CartRepository
}

func NewCartService(repo repository.CartRepository) *CartService {
	return &CartService{CartRepo: repo}

}