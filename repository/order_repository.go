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
// OrderRepository defines the database operations for orders.
type OrderRepository interface {

	// IsRestaurantActive checks if a restaurant is active and returns a boolean or an error.
	IsRestaurantActive(ctx context.Context, restaurantID uint) (bool, error)
	GetOrderByID(ctx context.Context, orderID uint, customerID uint) (*models.Order, error)

	// GetRestaurantOrders retrieves one page of a restaurant's orders and optionally filters them by status.
	GetRestaurantOrders(ctx context.Context, restaurantID uint, status string, limit, offset int) ([]models.Order, int64, error)

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

	err := r.db.WithContext(ctx).Select("status").First(&restaurant, restaurantID).Error
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
		Preload("Restaurant").
		Where("id = ? AND customer_id = ?", orderID, customerID).
		First(&order).Error
	if err != nil {
		return nil, err
	}

	return &order, nil
}

// GetRestaurantOrders retrieves one page of orders belonging to a specific restaurant.
// If no status is provided, orders of every status are included.
func (r *PostgresOrderRepo) GetRestaurantOrders(ctx context.Context, restaurantID uint, status string, limit, offset int) ([]models.Order, int64, error) {
	var total int64
	if err := r.restaurantOrdersQuery(ctx, restaurantID, status).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var orders []models.Order
	err := r.restaurantOrdersQuery(ctx, restaurantID, status).
		Preload("Items").
		Preload("Customer").
		Order("created_at DESC, id DESC").
		Limit(limit).
		Offset(offset).
		Find(&orders).Error
	if err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}

func (r *PostgresOrderRepo) restaurantOrdersQuery(ctx context.Context, restaurantID uint, status string) *gorm.DB {
	query := r.db.WithContext(ctx).Model(&models.Order{}).Where("restaurant_id = ?", restaurantID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	return query
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
