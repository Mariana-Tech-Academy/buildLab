package service

import "Fhub/repository"

// 1.CartService defines the interface for cart-related operations.
type CartService interface {
	// AddItem adds an item to the user's cart. If the user does not have a cart, it creates one.
	// If the cart already contains items from a different restaurant, it returns ErrCartRestaurantMismatch.
	AddItem(userID uint, menuItemID uint, quantity int) error
}

type cartService struct {
	cartRepo repository.CartRepository
}

func NewCartService(repo repository.CartRepository) CartService {
	return &cartService{cartRepo: repo}
}

// AddItem adds an item to the user's cart. If the user does not have a cart, it creates one.
// If the cart already contains items from a different restaurant, it returns ErrCartRestaurantMismatch.
func (s *cartService) AddItem(
	userID uint,
	menuItemID uint,
	quantity int,
) error {
	return s.cartRepo.AddItem(userID, menuItemID, quantity)
}