package service

import (
	"context"
	"errors"
	"fmt"

	"Fhub/models"
	"Fhub/repository"

	"gorm.io/gorm"
)

// CartService defines the shopping cart operations available to handlers.
type CartService interface {
	AddItemToCart(
		ctx context.Context,
		userID uint,
		menuItemID uint,
		quantity int,
	) (*models.Cart, error)
}

// cartService contains the repositories needed for cart operations.
type cartService struct {
	cartRepo repository.CartRepository
	menuRepo repository.MenuRepository
}

// NewCartService creates a cart service with access to cart and menu data.
func NewCartService(
	cartRepo repository.CartRepository,
	menuRepo repository.MenuRepository,
) CartService {
	return &cartService{
		cartRepo: cartRepo,
		menuRepo: menuRepo,
	}
}

// AddItemToCart adds a menu item to the authenticated customer's cart.
func (s *cartService) AddItemToCart(
	ctx context.Context,
	userID uint,
	menuItemID uint,
	quantity int,
) (*models.Cart, error) {

	// Reject quantities of zero or less.
	if quantity < 1 {
		return nil, errors.New("quantity must be at least 1")
	}

	// Find the requested menu item.
	menuItem, err := s.menuRepo.GetMenuItemByID(ctx, menuItemID)
	if err != nil {
		return nil, fmt.Errorf("menu item not found: %w", err)
	}

	// Customers cannot add unavailable menu items.
	if !menuItem.Available {
		return nil, errors.New("menu item is not available")
	}

	// Find the cart belonging to the authenticated customer.
	cart, err := s.cartRepo.GetCartByUserID(ctx, userID)

	if errors.Is(err, gorm.ErrRecordNotFound) {

		// Create a cart if the customer does not have one.
		cart = &models.Cart{
			UserID:       userID,
			RestaurantID: &menuItem.RestaurantID,
		}

		if err := s.cartRepo.CreateCart(ctx, cart); err != nil {
			return nil, fmt.Errorf("failed to create cart: %w", err)
		}

	} else if err != nil {

		// Return unexpected database errors.
		return nil, fmt.Errorf("failed to retrieve cart: %w", err)

	} else {

		// Set the restaurant when an existing cart has no items.
		if cart.RestaurantID == nil {
			cart.RestaurantID = &menuItem.RestaurantID
		}

		// Prevent mixing food from different restaurants.
		if *cart.RestaurantID != menuItem.RestaurantID {
			return nil, errors.New("cannot add items from different restaurants")
		}
	}

	// Check whether the menu item is already in the cart.
	existingItem, err := s.cartRepo.GetCartItem(
		ctx,
		cart.ID,
		menuItemID,
	)

	if err == nil {

		// Increase the quantity instead of creating a duplicate item.
		newQuantity := existingItem.Quantity + quantity

		if err := s.cartRepo.UpdateCartItemQuantity(
			ctx,
			existingItem.ID,
			newQuantity,
		); err != nil {
			return nil, fmt.Errorf("failed to update quantity: %w", err)
		}

	} else if errors.Is(err, gorm.ErrRecordNotFound) {

		// Create a new cart item when it does not already exist.
		cartItem := &models.CartItem{
			CartID:     cart.ID,
			MenuItemID: menuItemID,
			Quantity:   quantity,
		}

		if err := s.cartRepo.CreateCartItem(ctx, cartItem); err != nil {
			return nil, fmt.Errorf("failed to add cart item: %w", err)
		}

	} else {

		// Return unexpected database errors.
		return nil, fmt.Errorf("failed to check cart item: %w", err)
	}

	// Return the customer's cart.
	return cart, nil
}
