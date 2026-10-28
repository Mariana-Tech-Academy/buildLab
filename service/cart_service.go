package service

import (
	"context"
	"errors"
	"fmt"

	"Fhub/models"
	"Fhub/repository"

	"gorm.io/gorm"
)

const maxCartItemQuantity = 99

var (
	ErrInvalidCartInput       = errors.New("invalid cart item")
	ErrCartRestaurantConflict = errors.New("cannot add items from different restaurants")
	ErrCartItemNotFound       = errors.New("cart item not found")
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
// CartLine represents an item displayed in the customer's cart.
type CartLine struct {
    ID              uint   `json:"id"`
    MenuItemID      uint   `json:"menu_item_id"`
    Name            string `json:"name"`
    UnitPrice       int64  `json:"unit_price"`
    PriceAtAddition int64  `json:"price_at_addition"`
    Quantity        int    `json:"quantity"`
    LineTotal       int64  `json:"line_total"`
    Available       bool   `json:"available"`
    PriceChanged    bool   `json:"price_changed"`
}

// CartSummary represents the customer's cart and subtotal.
type CartSummary struct {
    ID           uint       `json:"id"`
    RestaurantID *uint      `json:"restaurant_id"`
    Items        []CartLine `json:"items"`
    Subtotal     int64      `json:"subtotal"`
}
// CartService defines the shopping cart operations available to handlers.
type CartService interface {
	AddItemToCart(
		ctx context.Context,
		userID uint,
		menuItemID uint,
		quantity int,
	) (*AddedCartItem, error)
	UpdateQuantity(ctx context.Context, userID, cartItemID uint, input UpdateCartQuantityInput) (*models.Cart, error)
	GetCart(ctx context.Context, userID uint) (*CartSummary, error)
}

// UpdateCartQuantityInput holds the new quantity sent by the customer.
type UpdateCartQuantityInput struct {
	Quantity int
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
		menuItem.Price,
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

// UpdateQuantity replaces the quantity of a line in the customer's own cart.
func (s *cartService) UpdateQuantity(ctx context.Context, userID, cartItemID uint, input UpdateCartQuantityInput) (*models.Cart, error) {
	if cartItemID == 0 {
		return nil, fmt.Errorf("%w: cart item ID is required", ErrInvalidCartInput)
	}
	if input.Quantity < 1 || input.Quantity > maxCartItemQuantity {
		return nil, fmt.Errorf("%w: quantity must be between 1 and %d", ErrInvalidCartInput, maxCartItemQuantity)
	}

	cart, err := s.cartRepo.GetCartWithItems(ctx, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrCartItemNotFound
	}
	if err != nil {
		return nil, err
	}

	var line *models.CartItem
	for i := range cart.Items {
		if cart.Items[i].ID == cartItemID {
			line = &cart.Items[i]
			break
		}
	}
	if line == nil {
		return nil, ErrCartItemNotFound
	}
	if line.MenuItem.ID == 0 || !line.MenuItem.Available {
		return nil, ErrMenuItemNotFound
	}
	if _, err := s.restaurantService.GetActiveRestaurant(ctx, line.MenuItem.RestaurantID); err != nil {
		if errors.Is(err, ErrRestaurantNotFound) {
			return nil, ErrMenuItemNotFound
		}
		return nil, err
	}

	err = s.cartRepo.SetItemQuantity(ctx, cart.ID, cartItemID, input.Quantity)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrCartItemNotFound
	}
	if errors.Is(err, repository.ErrQuantityLimit) {
		return nil, fmt.Errorf("%w: quantity must be between 1 and %d", ErrInvalidCartInput, maxCartItemQuantity)
	}
	if err != nil {
		return nil, err
	}

	return s.cartRepo.GetCartWithItems(ctx, userID)
}
// GetCart retrieves the customer's cart and calculates its subtotal.
func (s *cartService) GetCart(
    ctx context.Context,
    userID uint,
) (*CartSummary, error) {

    cart, err := s.cartRepo.GetCartWithItems(ctx, userID)

    // A customer who has never added an item has an empty cart.
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return &CartSummary{
            Items: []CartLine{},
        }, nil
    }

    if err != nil {
        return nil, err
    }

    result := &CartSummary{
        ID:           cart.ID,
        RestaurantID: cart.RestaurantID,
        Items:        make([]CartLine, 0, len(cart.Items)),
    }

    for _, item := range cart.Items {
        currentPrice := item.MenuItem.Price
        available := item.MenuItem.ID != 0 && item.MenuItem.Available

        // A zero snapshot indicates a legacy item without a known original price.
        priceChanged := item.PriceAtAddition != 0 &&
            item.PriceAtAddition != currentPrice

        lineTotal := currentPrice * int64(item.Quantity)

        result.Items = append(result.Items, CartLine{
            ID:              item.ID,
            MenuItemID:      item.MenuItemID,
            Name:            item.MenuItem.Name,
            UnitPrice:       currentPrice,
            PriceAtAddition: item.PriceAtAddition,
            Quantity:        item.Quantity,
            LineTotal:       lineTotal,
            Available:       available,
            PriceChanged:    priceChanged,
        })

        if available {
            result.Subtotal += lineTotal
        }
    }

    return result, nil
}