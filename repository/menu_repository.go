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

	ListByRestaurant(ctx context.Context, restaurantID uint, categoryID *uint) ([]models.MenuItem, error)
	Search(ctx context.Context, filters MenuSearchFilter) ([]models.MenuItem, int64, error)
}

// MenuSearchFilter is the database query for menu search.
// Prices are minor units.
type MenuSearchFilter struct {
	Name         string
	CategoryID   *uint
	RestaurantID *uint
	MinPrice     *int64
	MaxPrice     *int64
	Available    *bool
	Limit        int
	Offset       int
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

func (r *PostgresMenuRepo) ListByRestaurant(ctx context.Context, restaurantID uint, categoryID *uint) ([]models.MenuItem, error) {
	var items []models.MenuItem

	query := r.db.WithContext(ctx).
		Where("restaurant_id = ?", restaurantID)

	if categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	}

	err := query.Order("name ASC, id ASC").Find(&items).Error
	return items, err
}

// Search returns one page of menu items from active restaurants.
func (r *PostgresMenuRepo) Search(ctx context.Context, filters MenuSearchFilter) ([]models.MenuItem, int64, error) {
	itemsQuery := r.menuSearchQuery(ctx, filters)
	var total int64
	if err := itemsQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []models.MenuItem
	err := r.menuSearchQuery(ctx, filters).
		Select("menu_items.*").
		Order("menu_items.name ASC, menu_items.id ASC").
		Limit(filters.Limit).
		Offset(filters.Offset).
		Find(&items).Error
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *PostgresMenuRepo) menuSearchQuery(ctx context.Context, filters MenuSearchFilter) *gorm.DB {
	query := r.db.WithContext(ctx).Model(&models.MenuItem{}).
		Joins("JOIN restaurants ON restaurants.id = menu_items.restaurant_id AND restaurants.deleted_at IS NULL").
		Where("restaurants.status = ?", models.RestaurantStatusActive)
	if filters.Name != "" {
		query = query.Where("LOWER(menu_items.name) LIKE LOWER(?)", likeContains(filters.Name))
	}
	if filters.CategoryID != nil {
		query = query.Where("menu_items.category_id = ?", *filters.CategoryID)
	}
	if filters.RestaurantID != nil {
		query = query.Where("menu_items.restaurant_id = ?", *filters.RestaurantID)
	}
	if filters.MinPrice != nil {
		query = query.Where("menu_items.price >= ?", *filters.MinPrice)
	}
	if filters.MaxPrice != nil {
		query = query.Where("menu_items.price <= ?", *filters.MaxPrice)
	}
	if filters.Available != nil {
		query = query.Where("menu_items.available = ?", *filters.Available)
	}
	return query
}
