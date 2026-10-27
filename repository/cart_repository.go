package repository

import (
	"context"
	"errors"

	"Fhub/models"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrDifferentRestaurant = errors.New("cannot add items from different restaurants")
	ErrQuantityLimit       = errors.New("quantity exceeds the maximum")
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

	// AddItemForUser stores the cart restaurant and inserts or increments the line in one transaction.
	AddItemForUser(ctx context.Context, userID, restaurantID, menuItemID uint, quantity, maxQuantity int) (cartID uint, newQuantity int, err error)

	// GetCartWithItems finds a customer's cart, including its items and their menu details.
	GetCartWithItems(ctx context.Context, userID uint) (*models.Cart, error)

	// SetItemQuantity replaces a cart item's quantity when the item belongs to the given cart.
	SetItemQuantity(ctx context.Context, cartID, cartItemID uint, quantity int) error
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

// AddItemForUser locks the customer's cart, records its restaurant, and adds the menu item.
func (r *PostgresCartRepo) AddItemForUser(
	ctx context.Context,
	userID uint,
	restaurantID uint,
	menuItemID uint,
	quantity int,
	maxQuantity int,
) (uint, int, error) {
	if quantity < 1 || maxQuantity < 1 || quantity > maxQuantity {
		return 0, 0, ErrQuantityLimit
	}

	var cartID uint
	var newQuantity int
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		cart, err := lockOrCreateCart(tx, userID, restaurantID)
		if err != nil {
			return err
		}
		quantityNow, err := upsertCartItem(tx, cart.ID, menuItemID, quantity, maxQuantity)
		if err != nil {
			return err
		}
		cartID = cart.ID
		newQuantity = quantityNow
		return nil
	})
	if err != nil {
		return 0, 0, err
	}
	return cartID, newQuantity, nil
}

func lockOrCreateCart(tx *gorm.DB, userID, restaurantID uint) (*models.Cart, error) {
	var cart models.Cart
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ?", userID).
		First(&cart).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := tx.SavePoint("create_cart").Error; err != nil {
			return nil, err
		}
		cart = models.Cart{UserID: userID, RestaurantID: &restaurantID}
		if err := tx.Create(&cart).Error; err != nil {
			if !isUniqueViolation(err) {
				return nil, err
			}
			if err := tx.RollbackTo("create_cart").Error; err != nil {
				return nil, err
			}
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("user_id = ?", userID).
				First(&cart).Error; err != nil {
				return nil, err
			}
		} else {
			return &cart, nil
		}
	} else if err != nil {
		return nil, err
	}

	if cart.RestaurantID == nil {
		if err := tx.Model(&models.Cart{}).Where("id = ?", cart.ID).Update("restaurant_id", restaurantID).Error; err != nil {
			return nil, err
		}
		cart.RestaurantID = &restaurantID
		return &cart, nil
	}
	if *cart.RestaurantID != restaurantID {
		return nil, ErrDifferentRestaurant
	}
	return &cart, nil
}

func upsertCartItem(tx *gorm.DB, cartID, menuItemID uint, quantity, maxQuantity int) (int, error) {
	var item models.CartItem
	err := tx.Where("cart_id = ? AND menu_item_id = ?", cartID, menuItemID).First(&item).Error
	if err == nil {
		return setCartItemQuantity(tx, &item, item.Quantity, quantity, maxQuantity)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}

	if err := tx.SavePoint("create_item").Error; err != nil {
		return 0, err
	}
	item = models.CartItem{CartID: cartID, MenuItemID: menuItemID, Quantity: quantity}
	if err := tx.Create(&item).Error; err != nil {
		if !isUniqueViolation(err) {
			return 0, err
		}
		if err := tx.RollbackTo("create_item").Error; err != nil {
			return 0, err
		}
		if err := tx.Where("cart_id = ? AND menu_item_id = ?", cartID, menuItemID).First(&item).Error; err != nil {
			return 0, err
		}
		return setCartItemQuantity(tx, &item, item.Quantity, quantity, maxQuantity)
	}
	return quantity, nil
}

func setCartItemQuantity(tx *gorm.DB, item *models.CartItem, current, add, maxQuantity int) (int, error) {
	if add < 1 || current > maxQuantity-add {
		return 0, ErrQuantityLimit
	}
	updated := current + add
	if err := tx.Model(item).Update("quantity", updated).Error; err != nil {
		return 0, err
	}
	return updated, nil
}

// GetCartWithItems finds a customer's cart, including its items and their menu details.
func (r *PostgresCartRepo) GetCartWithItems(ctx context.Context, userID uint) (*models.Cart, error) {
	var cart models.Cart
	err := r.db.WithContext(ctx).
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("id ASC")
		}).
		Preload("Items.MenuItem").
		Where("user_id = ?", userID).
		First(&cart).Error
	if err != nil {
		return nil, err
	}
	return &cart, nil
}

// SetItemQuantity replaces a cart item's quantity when the item belongs to the given cart.
func (r *PostgresCartRepo) SetItemQuantity(ctx context.Context, cartID, cartItemID uint, quantity int) error {
	if quantity < 1 {
		return ErrQuantityLimit
	}
	result := r.db.WithContext(ctx).
		Model(&models.CartItem{}).
		Where("id = ? AND cart_id = ?", cartItemID, cartID).
		Update("quantity", quantity)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func isUniqueViolation(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
