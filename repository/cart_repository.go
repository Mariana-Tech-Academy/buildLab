package repository

import (
	"context"
	"errors"

	"Fhub/models"

	"gorm.io/gorm"
)

// CartRepository defines the database operations available for shopping carts.
// Both CART-01 and CART-03 can use these methods.
type CartRepository interface {

	// GetCartByUserID finds the cart belonging to a particular customer.
	GetCartByUserID(ctx context.Context, userID uint) (*models.Cart, error)

	// CreateCart saves a new shopping cart in the database.
	CreateCart(ctx context.Context, cart *models.Cart) error

	// GetCartItem finds a particular menu item inside a shopping cart.
	GetCartItem(ctx context.Context, cartID uint, menuItemID uint) (*models.CartItem, error)

	// CreateCartItem adds a new menu item to the customer's cart.
	CreateCartItem(ctx context.Context, item *models.CartItem) error

	// UpdateCartItemQuantity changes the quantity of an existing cart item.
	// Hizra can reuse this method for CART-03.
	UpdateCartItemQuantity(ctx context.Context, itemID uint, quantity int) error
}

// PostgresCartRepo holds the database connection used by the repository.
type PostgresCartRepo struct {
	db *gorm.DB
}

// NewCartRepository creates a new cart repository using the database connection.
func NewCartRepository(db *gorm.DB) CartRepository {
	return &PostgresCartRepo{db: db}
}

// GetCartByUserID retrieves a customer's shopping cart.
func (r *PostgresCartRepo) GetCartByUserID(
	ctx context.Context,
	userID uint,
) (*models.Cart, error) {

	// Create a variable to store the cart returned from PostgreSQL.
	var cart models.Cart

	// Find the cart where user_id matches the customer's ID.
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&cart).Error

	// Return the error if the cart cannot be retrieved.
	if err != nil {
		return nil, err
	}

	// Return the cart of the customer.
	return &cart, nil
}

// CreateCart saves a new cart for a customer.
func (r *PostgresCartRepo) CreateCart(
	ctx context.Context,
	cart *models.Cart,
) error {

	// Insert the new cart into PostgreSQL.
	return r.db.WithContext(ctx).
		Create(cart).Error
}

// GetCartItem checks whether a particular menu item is already in the cart.
func (r *PostgresCartRepo) GetCartItem(
	ctx context.Context,
	cartID uint,
	menuItemID uint,
) (*models.CartItem, error) {

	// Create a variable to hold the cart item.
	var item models.CartItem

	// Find an item matching both the cart ID and menu item ID.
	err := r.db.WithContext(ctx).
		Where("cart_id = ? AND menu_item_id = ?", cartID, menuItemID).
		First(&item).Error

	// Return an error if the item cannot be retrieved.
	if err != nil {
		return nil, err
	}

	// Return the matching cart item.
	return &item, nil
}

// CreateCartItem inserts a new menu item into a customer's cart.
func (r *PostgresCartRepo) CreateCartItem(
	ctx context.Context,
	item *models.CartItem,
) error {

	// Save the cart item in PostgreSQL.
	return r.db.WithContext(ctx).
		Create(item).Error
}

// UpdateCartItemQuantity updates the quantity of an existing cart item.
func (r *PostgresCartRepo) UpdateCartItemQuantity(
	ctx context.Context,
	itemID uint,
	quantity int,
) error {

	// Reject quantities below 1.
	if quantity < 1 {
		return errors.New("quantity must be at least 1")
	}

	// Update the quantity of the cart item with the matching ID.
	result := r.db.WithContext(ctx).
		Model(&models.CartItem{}).
		Where("id = ?", itemID).
		Update("quantity", quantity)

	// Return any database error.
	if result.Error != nil {
		return result.Error
	}

	// Return an error if no cart item was updated.
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	// Return nil to indicate the update succeeded.
	return nil
}
