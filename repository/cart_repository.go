package repository

import (
	"Fhub/models"
	"context"

	"gorm.io/gorm"
)

type CartRepository interface {
	GetCartWithItems(ctx context.Context, userID uint) (*models.Cart, error)
	SetItemQuantity(ctx context.Context, cartID, cartItemID uint, quantity int) error
}

type PostgresCartRepo struct {
	db *gorm.DB
}

func NewCartRepository(db *gorm.DB) CartRepository {
	return &PostgresCartRepo{db: db}
}

// GetCArtWithItems finds a customer's cart incl. it's items and item's menu details
func (r *PostgresCartRepo) GetCartWithItems(ctx context.Context, userID uint) (*models.Cart, error) {
	var cart models.Cart
	//preload loads the cart items and their menu items (needed for prices and totals)
	err := r.db.WithContext(ctx).
		Preload("Items.MenuItem").
		Where("user_id = ?", userID).
		First(&cart).Error
	if err != nil {
		return nil, err
	}
	return &cart, nil
}

// SetItemQuantity replaces a cart item's quantity, but only if the item is in the given cart
func (r *PostgresCartRepo) SetItemQuantity(ctx context.Context, cartID, cartItemID uint, quantity int) error {
	result := r.db.WithContext(ctx).
		Model(&models.CartItem{}).
		Where("id =? AND cart_id = ?", cartItemID, cartID).
		Update("quantity", quantity)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil

}
