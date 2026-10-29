package repository

import (
	"Fhub/models"
	"context"

	"gorm.io/gorm"
)

type AdminRepository interface {
	//ListAllOrders gets a single page or every order, optionally filtered by status
	ListAllOrders(ctx context.Context, status, search string, limit, offset int) ([]models.Order, int64, error)

	//GetOrderForAdmin gets one order with its customer, restaurant and items
	GetOrderForAdmin(ctx context.Context, orderID uint) (*models.Order, error)
}

type PostgresAdminRepo struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &PostgresAdminRepo{db: db}
}

// ListAllOrders gets one page of orders for the admin, plus thte total no. of matching orders.
func (r *PostgresAdminRepo) ListAllOrders(ctx context.Context, status, search string, limit, offset int) ([]models.Order, int64, error) {
	var total int64
	if err := r.adminOrdersQuery(ctx, status, search).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var orders []models.Order
	err := r.adminOrdersQuery(ctx, status, search).
		Preload("Items").
		Preload("Customer").
		Preload("Restaurant").
		Order("orders.created_at DESC, orders.id DESC").
		Limit(limit).
		Offset(offset).
		Find(&orders).Error
	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (r *PostgresAdminRepo) adminOrdersQuery(ctx context.Context, status, search string) *gorm.DB {
	query := r.db.WithContext(ctx).Model(&models.Order{})

	if status != "" {
		query = query.Where("orders.status = ?", status)
	}

	if search != "" {
		like := "%" + search + "%"
		query = query.
			Joins("JOIN users ON users.id = orders.customer_id").
			Where("orders.order_number ILIKE ? OR users.email ILIKE ?", like, like)
	}

	return query
}

// GetOrderForAdmin gets one order by ID with its customer, restaurant and items
func (r *PostgresAdminRepo) GetOrderForAdmin(ctx context.Context, orderID uint) (*models.Order, error) {
	var order models.Order

	err := r.db.WithContext(ctx).
		Preload("Items").
		Preload("Customer").
		Preload("Restaurant").
		First(&order, orderID).Error
	if err != nil {
		return nil, err
	}

	return &order, nil
}
