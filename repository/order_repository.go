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
