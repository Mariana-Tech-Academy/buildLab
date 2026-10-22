package repository

import (
	"context"


	"Fhub/models"

	"gorm.io/gorm"
)


// MenuRepository defines database operations for menu items.
type MenuRepository interface {
	CreateMenuItem(ctx context.Context, item *models.MenuItem) error
	GetMenuItemByID(ctx context.Context, itemID uint) (*models.MenuItem, error)
	UpdateMenuItem(ctx context.Context, item *models.MenuItem) error
	DeleteMenuItem(ctx context.Context, itemID, restaurantID uint) error

	ListByRestaurant(ctx context.Context,restaurantID uint,categoryID *uint,) ([]models.MenuItem, error)
	//GetByID(ctx context.Context,itemID uint,) (*models.MenuItem, error)
}

// PostgresMenuRepo stores the connection of the database.
type PostgresMenuRepo struct {
	db *gorm.DB
}

// NewMenuRepository creates a menu repository.
func NewMenuRepository(db *gorm.DB) MenuRepository {
	return &PostgresMenuRepo{db: db}
}

// CreateMenuItem saves a new menu item.
func (r *PostgresMenuRepo) CreateMenuItem(ctx context.Context, item *models.MenuItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

// GetMenuItemByID finds one existing menu item using its ID.
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
		"Available",
		"UpdatedAt",
	).Updates(item)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeleteMenuItem soft-deletes a menu item belonging to a specific restaurant.
func (r *PostgresMenuRepo) DeleteMenuItem(ctx context.Context, itemID, restaurantID uint) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND restaurant_id = ?", itemID, restaurantID).
		Delete(&models.MenuItem{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}



func (r *PostgresMenuRepo) ListByRestaurant(ctx context.Context,restaurantID uint,categoryID *uint,) ([]models.MenuItem, error) {
    var items []models.MenuItem

    query := r.db.WithContext(ctx).
        Model(&models.MenuItem{}).
        Preload("Category").
        Where("restaurant_id = ?", restaurantID)

    if categoryID != nil {
        query = query.Where("category_id = ?", *categoryID)
    }

    err := query.Order("name ASC").Find(&items).Error

    return items, err
}

