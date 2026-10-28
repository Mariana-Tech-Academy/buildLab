package repository

import (
	"context"
	"errors"

	"Fhub/models"

	"gorm.io/gorm"
)

var ErrEmptyCart = errors.New("cannot checkout with an empty cart")
var ErrUnavailableItem = errors.New("cart contains an unavailable item")

// DeliveryFee is a fixed checkout fee of 2000.00 in minor units.
const DeliveryFee int64 = 200000

// Checkout must create the order, create its items and clear the cart inside one
// r.db.Transaction(func(tx *gorm.DB) error { ... }) call.
type OrderRepository interface {
	IsRestaurantActive(ctx context.Context, restaurantID uint) (bool, error)
	GetCheckoutSummary(ctx context.Context, customerID uint) (*models.CheckoutSummary, error)
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

func (r *PostgresOrderRepo) GetCheckoutSummary(ctx context.Context, customerID uint) (*models.CheckoutSummary, error) {
	var cart models.Cart

	err := r.db.WithContext(ctx).
		Preload("Restaurant").
		Preload("Items.MenuItem").
		Where("user_id = ?", customerID).
		First(&cart).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrEmptyCart
	}
	if err != nil {
		return nil, err
	}

	if len(cart.Items) == 0 || cart.RestaurantID == nil {
		return nil, ErrEmptyCart
	}

	if cart.Restaurant == nil {
		return nil, gorm.ErrRecordNotFound
	}

	summary := &models.CheckoutSummary{
		RestaurantID:   cart.Restaurant.ID,
		RestaurantName: cart.Restaurant.Name,
		Items:          make([]models.CheckoutItem, 0, len(cart.Items)),
		DeliveryFee:    DeliveryFee,
	}
	for _, item := range cart.Items {
		if item.Quantity <= 0 ||
			item.MenuItem.ID == 0 ||
			!item.MenuItem.Available ||
			item.MenuItem.RestaurantID != cart.Restaurant.ID {
			return nil, ErrUnavailableItem
		}

		lineTotal := item.MenuItem.Price * int64(item.Quantity)

		summary.Items = append(summary.Items, models.CheckoutItem{
			MenuItemID: item.MenuItemID,
			Name:       item.MenuItem.Name,
			Quantity:   item.Quantity,
			UnitPrice:  item.MenuItem.Price,
			LineTotal:  lineTotal,
		})

		summary.Subtotal += lineTotal
	}

	summary.Total = summary.Subtotal + summary.DeliveryFee

	return summary, nil
}
