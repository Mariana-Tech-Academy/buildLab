package repository

import (
	"Fhub/models"
	"context"

	"gorm.io/gorm"
)

// MenuRepository defines database operations for menu items.
type MenuRepository interface {
	GetMenuItemByID(ctx context.Context, itemID uint) (*models.MenuItem, error)
	UpdateMenuItem(ctx context.Context, item *models.MenuItem) error
	DeleteMenuItem(ctx context.Context, itemID, restaurantID uint) error
}

// PostgresMenuRepo holds the database connection
type PostgresMenuRepo struct {
	db *gorm.DB
}

// NewMenuRepository creates a menu repository.
func NewMenuRepository(db *gorm.DB) MenuRepository {
	return &PostgresMenuRepo{db: db}
}

// GetMenuItemByID -  Finds one existing menu item using its ID.
func (r *PostgresMenuRepo) GetMenuItemByID(ctx context.Context, itemID uint) (*models.MenuItem, error) {
	var item models.MenuItem

	err := r.db.WithContext(ctx).First(&item, itemID).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// UpdateMenuItem saves only the fields allowed by the edit operation.
func (r *PostgresMenuRepo) UpdateMenuItem(ctx context.Context, item *models.MenuItem) error {
	result := r.db.WithContext(ctx).Model(item).Select(
		"Name",
		"Description",
		"CategoryID",
		"Price",
		"ImageURL",
		"UpdatedAt",
	).Updates(item)

	// Return any database error
	if result.Error != nil {
		return result.Error
	}
	// Report when no existing item was updated
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil

}

// DeleteMenuItem - Soft-deletes a menu item belonging to a specific restaurant.
func (r *PostgresMenuRepo) DeleteMenuItem(ctx context.Context, itemID uint, restuarantID uint) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND restaurant_id = ?",
			itemID,
			restuarantID,
		).
		Delete(&models.MenuItem{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound // This lets the service recognise a missing item and return 404.
	}
	return nil
}
