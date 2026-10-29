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

    GetOrderByID(ctx context.Context, orderID uint) (*models.Order, error)

    UpdateOrderStatus(
        ctx context.Context,
        orderID uint,
        restaurantID uint,
        currentStatus models.OrderStatus,
        newStatus models.OrderStatus,
    ) error
}
func (r *PostgresOrderRepo) GetOrderByID(
    ctx context.Context,
    orderID uint,
) (*models.Order, error) {
    var order models.Order

    err := r.db.WithContext(ctx).
        First(&order, orderID).Error

    if err != nil {
        return nil, err
    }

    return &order, nil
}

func (r *PostgresOrderRepo) UpdateOrderStatus(
    ctx context.Context,
    orderID uint,
    restaurantID uint,
    currentStatus models.OrderStatus,
    newStatus models.OrderStatus,
) error {

    result := r.db.WithContext(ctx).
        Model(&models.Order{}).
        Where(
            "id = ? AND restaurant_id = ? AND status = ?",
            orderID,
            restaurantID,
            currentStatus,
        ).
        Update("status", newStatus)

    if result.Error != nil {
        return result.Error
    }

    if result.RowsAffected == 0 {
        return gorm.ErrRecordNotFound
    }

    return nil
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
