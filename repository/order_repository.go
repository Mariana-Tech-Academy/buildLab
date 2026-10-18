package repository

import (
	"Fhub/models"
	"context"
	"errors"

	"gorm.io/gorm"
)

// Checkout must create the order, create its items and clear the cart inside one
// r.db.Transaction(func(tx *gorm.DB) error { ... }) call.
var ErrRestaurantNotActive = errors.New("restaurant is not active")

type OrderRepository interface {
	IsRestaurantActive(ctx context.Context, restaurantID uint) (bool, error)
	Checkout(ctx context.Context, userID uint, order *models.Order) error

}

type PostgresOrderRepo struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &PostgresOrderRepo{db: db}
}

func (r *PostgresOrderRepo) IsRestaurantActive(
	ctx context.Context,
	restaurantID uint,
) (bool, error) {
	var restaurant models.Restaurant

	err := r.db.WithContext(ctx).
		Select("status").
		First(&restaurant, restaurantID).Error
	if err != nil {
		return false, err
	}

	return restaurant.Status == models.RestaurantStatusActive, nil
}
func (r *PostgresOrderRepo) Checkout(
	ctx context.Context,
	userID uint,
	order *models.Order,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var cart models.Cart

		if err := tx.Preload("Items.MenuItem").
			First(&cart, "user_id = ?", userID).Error; err != nil {
			return err
		}

		if cart.RestaurantID == nil {
			return errors.New("cart has no restaurant")
		}

		var restaurant models.Restaurant
		if err := tx.First(&restaurant, *cart.RestaurantID).Error; err != nil {
			return err
		}

		// REST-03: reject orders for inactive restaurants.
		if restaurant.Status != models.RestaurantStatusActive {
			return ErrRestaurantNotActive
		}

		// No order is saved here: the checkout flow must provide the
		// order number and delivery details before creating the order.
		return errors.New("checkout requires order details")
	})
}