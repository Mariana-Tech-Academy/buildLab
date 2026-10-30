package repository

import (
	"context"

	"Fhub/models"

	"gorm.io/gorm"
)

type AdminRepository interface {
	// ListAllOrders gets one page of orders, optionally filtered by status and search.
	ListAllOrders(ctx context.Context, status, search string, limit, offset int) ([]models.Order, int64, error)

	// GetOrderForAdmin gets one order with its customer, restaurant and items.
	GetOrderForAdmin(ctx context.Context, orderID uint) (*models.Order, error)

	// AdminUpdateOrderStatus changes an order's status and records which admin did it, in one transaction.
	AdminUpdateOrderStatus(ctx context.Context, orderID uint, fromStatus, toStatus models.OrderStatus, adminID uint) error
}

type PostgresAdminRepo struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &PostgresAdminRepo{db: db}
}

// ListAllOrders gets one page of orders for the admin, plus the total number of matching orders.
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
		pattern := likeContains(search)
		query = query.
			Joins("LEFT JOIN users ON users.id = orders.customer_id").
			Where("orders.order_number ILIKE ? OR users.email ILIKE ?", pattern, pattern)
	}

	return query
}

// GetOrderForAdmin gets one order by ID with its customer, restaurant and items.
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

// AdminUpdateOrderStatus changes the order's status and saves a record of the change.
// Both happen in one transaction, so a status change is never saved without its record, or the other way round.
// The update only works if the order is still in fromStatus, so if someone else changed it at the same moment,
// nothing is changed.
func (r *PostgresAdminRepo) AdminUpdateOrderStatus(ctx context.Context, orderID uint, fromStatus, toStatus models.OrderStatus, adminID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&models.Order{}).
			Where("id = ? AND status = ?", orderID, fromStatus).
			Update("status", toStatus)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		change := models.OrderStatusChange{
			OrderID:    orderID,
			FromStatus: fromStatus,
			ToStatus:   toStatus,
			ChangedBy:  adminID,
		}
		return tx.Create(&change).Error
	})
}
