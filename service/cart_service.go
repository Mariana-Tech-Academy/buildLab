package service

import (
	"Fhub/models"
	"Fhub/repository"
	"context"
	"errors"

	"gorm.io/gorm"
)

// errors the cart handler checks for, so it can choose the appropriate status code
var (
	ErrInvalidQuantity = errors.New("quantity must be atleast 1")
	ErrCartItemNotFound = errors.New("cart item not found")
)

//UpdateQuantityInput holds the new quantity which is sent by the customer
type UpdateCartQuantityInput struct {
	Quantity int
}

//CartService lisst the cart actions which the handler can use
type CartService interface {
	UpdateQuantity(ctx context.Context, userID, cartItemID uint, input UpdateCartQuantityInput) (*models.Cart, error)
}

//Holds the cart repository
type cartService struct {
	cartRepo repository.CartRepository
}

//Creates the cart service with its repository 
func NewCartService(repo repository.CartRepository) CartService {
	return &cartService{cartRepo: repo}
}

//UpdateQuantity replaces the quantity of an item in the customer's own cart
//and will return the updated cart
func (s *cartService) UpdateQuantity(ctx context.Context, userID, cartItemID uint, input UpdateCartQuantityInput) (*models.Cart, error) {
	//Reject a quantity of 0 or less 
	if input.Quantity < 1 {
		return nil, ErrInvalidQuantity
	}

	//Find the logged in customer's cart
	cart, err := s.cartRepo.GetCartWithItems(ctx, userID)
	if errors .Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrCartItemNotFound
	}
	if err != nil {
		return nil, err
	}

	err = s.cartRepo.SetItemQuantity(ctx, cart.ID, cartItemID, input.Quantity)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrCartItemNotFound
	}

	if err != nil {
		return nil, err
	}

	//load the cart so the response shows the new quantity and totals
	return s.cartRepo.GetCartWithItems(ctx, userID)

}