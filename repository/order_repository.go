package repository

import (
	"context"

	"Fhub/models"

	"gorm.io/gorm"
)

// Checkout must create the order, create its items and clear the cart inside one
// r.db.Transaction(func(tx *gorm.DB) error { ... }) call.
type OrderRepository interface {
	IsRestaurantActive(ctx context.Context, restaurantID uint) (bool, error)
	GetOrderByID(ctx context.Context, orderID uint, customerID uint) (*models.Order, error)
}

type PostgresOrderRepo struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &PostgresOrderRepo{db: db}
}

func (r *PostgresOrderRepo) IsRestaurantActive(ctx context.Context, restaurantID uint) (bool, error) {
	var restaurant models.Restaurant

	err := r.db.WithContext(ctx).
		Select("status").
		First(&restaurant, restaurantID).Error
	if err != nil {
		return false, err
	}

	return restaurant.Status == models.RestaurantStatusActive, nil
}

// Loads all items associated with the order so the customer can view their order details.
func (r *PostgresOrderRepo) GetOrderByID(ctx context.Context, orderID uint, customerID uint) (*models.Order, error) {
	var order models.Order

	err := r.db.WithContext(ctx).
		Preload("Items").
		Where("id = ? AND customer_id = ?", orderID, customerID).
		First(&order).Error

	if err != nil {
		return nil, err
	}

	return &order, nil
}
