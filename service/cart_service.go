package service

import (
    "Fhub/models"
    "Fhub/repository"
)

type CartResult struct {
    Cart  *models.Cart `json:"cart"`
    Total int64        `json:"total"`
}

type CartService interface {
	RemoveItem(userID, itemID uint) (*CartResult, error)
    Clear(userID uint) (*CartResult, error)
}

type cartService struct {
	cartRepo repository.CartRepository
}

func NewCartService(repo repository.CartRepository) CartService {
	return &cartService{cartRepo: repo}
}

func cartResult(cart *models.Cart) *CartResult {
    if cart.Items == nil {
        cart.Items = []models.CartItem{}
    }

    var total int64

    for _, item := range cart.Items {
        total += item.MenuItem.Price * int64(item.Quantity)
    }

    return &CartResult{
        Cart:  cart,
        Total: total,
    }
}

// Remove one item from the cart.
func (s *cartService) RemoveItem(
    userID, itemID uint,
) (*CartResult, error) {

    cart, err := s.cartRepo.RemoveItem(userID, itemID)

    if err != nil {
        return nil, err
    }

    return cartResult(cart), nil
}

// Clear all items from the cart.
func (s *cartService) Clear(
    userID uint,
) (*CartResult, error) {

    cart, err := s.cartRepo.Clear(userID)

    if err != nil {
        return nil, err
    }

    return cartResult(cart), nil
}
