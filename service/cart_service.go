package service

import (
	"context"
	"errors"
	"fmt"

	"Fhub/repository"

	"gorm.io/gorm"
)

const maxCartItemQuantity = 99

var (
	ErrInvalidCartInput       = errors.New("invalid cart item")
	ErrCartRestaurantConflict = errors.New("cannot add items from different restaurants")
)

// AddedCartItem is the line stored for the customer.
type AddedCartItem struct {
	CartID       uint
	RestaurantID uint
	MenuItemID   uint
	Name         string
	Quantity     int
	Price        int64
}

// CartService defines the shopping cart operations available to handlers.
type CartService interface {
	AddItemToCart(
		ctx context.Context,
		userID uint,
		menuItemID uint,
		quantity int,
	) (*AddedCartItem, error)
}

// cartService contains the repositories needed for cart operations.
type cartService struct {
	cartRepo          repository.CartRepository
	menuRepo          repository.MenuRepository
	restaurantService RestaurantService
}

// NewCartService creates a cart service with access to cart, menu, and restaurant data.
func NewCartService(
	cartRepo repository.CartRepository,
	menuRepo repository.MenuRepository,
	restaurants RestaurantService,
) CartService {
	return &cartService{
		cartRepo:          cartRepo,
		menuRepo:          menuRepo,
		restaurantService: restaurants,
	}
}

// AddItemToCart adds a menu item to the authenticated customer's cart.
func (s *cartService) AddItemToCart(
	ctx context.Context,
	userID uint,
	menuItemID uint,
	quantity int,
) (*AddedCartItem, error) {
	if menuItemID == 0 {
		return nil, fmt.Errorf("%w: menu item ID is required", ErrInvalidCartInput)
	}
	if quantity < 1 || quantity > maxCartItemQuantity {
		return nil, fmt.Errorf("%w: quantity must be between 1 and %d", ErrInvalidCartInput, maxCartItemQuantity)
	}

	menuItem, err := s.menuRepo.GetMenuItemByID(ctx, menuItemID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrMenuItemNotFound
	}
	if err != nil {
		return nil, err
	}
	if !menuItem.Available {
		return nil, ErrMenuItemNotFound
	}
	if _, err := s.restaurantService.GetActiveRestaurant(ctx, menuItem.RestaurantID); err != nil {
		if errors.Is(err, ErrRestaurantNotFound) {
			return nil, ErrMenuItemNotFound
		}
		return nil, err
	}

	cartID, newQuantity, err := s.cartRepo.AddItemForUser(
		ctx,
		userID,
		menuItem.RestaurantID,
		menuItem.ID,
		quantity,
		maxCartItemQuantity,
	)
	if errors.Is(err, repository.ErrDifferentRestaurant) {
		return nil, ErrCartRestaurantConflict
	}
	if errors.Is(err, repository.ErrQuantityLimit) {
		return nil, fmt.Errorf("%w: quantity cannot exceed %d", ErrInvalidCartInput, maxCartItemQuantity)
	}
	if err != nil {
		return nil, err
	}

	return &AddedCartItem{
		CartID:       cartID,
		RestaurantID: menuItem.RestaurantID,
		MenuItemID:   menuItem.ID,
		Name:         menuItem.Name,
		Quantity:     newQuantity,
		Price:        menuItem.Price,
	}, nil
}
