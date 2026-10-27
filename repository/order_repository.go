package repository

import (
	"context"

	"Fhub/models"

	"gorm.io/gorm"
)

// OrderRepository defines the database operations for orders.
type OrderRepository interface {

	// IsRestaurantActive checks if a restaurant is active and returns a boolean or an error.
	IsRestaurantActive(ctx context.Context, restaurantID uint) (bool, error)

	// GetRestaurantOrders retrieves a restaurant's orders and optionally filters them by status.
	// []models.Order returns a list of orders, while error reports any database failure.
	GetRestaurantOrders(ctx context.Context, restaurantID uint, status string) ([]models.Order, error)
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

// GetRestaurantOrders retrieves orders belonging to a specific restaurant.
// It allows the restaurant manager to filter orders by status.
// If no status is provided, all orders belonging to that restaurant are returned.
func (r *PostgresOrderRepo) GetRestaurantOrders(ctx context.Context, restaurantID uint, status string) ([]models.Order, error) {

	// Create an empty slice to store the orders retrieved from PostgreSQL.
	var orders []models.Order

	// The dot (.) in r.db means we are accessing the db field belonging to r.
	// The dot (.) before WithContext, Model and Where means we are calling methods on the GORM query object.
	// WithContext(ctx) attaches the request context to the database operation.
	// Model(&models.Order{}) tells GORM that we want to query the orders table.
	// Where("restaurant_id = ?", restaurantID) ensures we retrieve only orders belonging to the specified restaurant.
	// Each GORM method returns a query object, allowing us to chain multiple methods together.
	query := r.db.WithContext(ctx).Model(&models.Order{}).Where("restaurant_id = ?", restaurantID)

	// Check whether the restaurant manager provided an order status.
	if status != "" {

		// The dot (.) before Where means we are calling the Where method on the existing query.
		// This adds another condition so that only orders matching the requested status are retrieved.
		query = query.Where("status = ?", status)
	}

	// The dot (.) is used to access methods or fields belonging to an object or value.
	// Preload("Items") tells GORM to load the items associated with each order.
	// Preload("Customer") tells GORM to load the customer associated with each order.
	// Order("created_at DESC") arranges the orders from newest to oldest.
	// Find(&orders) executes the database query and stores the results inside the orders slice.
	// The final .Error accesses the Error field of the GORM result to check whether the query failed.
	// All these methods are connected using dots, which is called method chaining.
	err := query.Preload("Items").Preload("Customer").Order("created_at DESC").Find(&orders).Error

	// Check whether an error occurred while retrieving the orders from PostgreSQL.
	if err != nil {

		// Return nil instead of orders and return the database error.
		return nil, err
	}

	// Return the retrieved orders and nil because the database query completed successfully.
	return orders, nil
}
