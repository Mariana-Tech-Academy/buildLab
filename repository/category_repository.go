package repository

import (
    "context"

    "Fhub/models"

    "gorm.io/gorm"
)

// CategoryRepository defines database operations.
type CategoryRepository interface {
    Create(ctx context.Context, category *models.Category) error
    FindByID(ctx context.Context, id uint) (*models.Category, error)
    FindByRestaurantID(ctx context.Context, restaurantID uint) ([]models.Category, error)
    Update(ctx context.Context, category *models.Category) error
    Delete(ctx context.Context, id uint, restaurantID uint) error
}

// PostgresCategoryRepo implements CategoryRepository.
type PostgresCategoryRepo struct {
    db *gorm.DB
}

// NewCategoryRepository creates a repository instance.
func NewCategoryRepository(db *gorm.DB) CategoryRepository {
    return &PostgresCategoryRepo{db: db}
}

// Create saves a new category.
func (r *PostgresCategoryRepo) Create(
    ctx context.Context,
    category *models.Category,
) error {
    return r.db.WithContext(ctx).Create(category).Error
}

// FindByID retrieves a category using its ID.
func (r *PostgresCategoryRepo) FindByID(
    ctx context.Context,
    id uint,
) (*models.Category, error) {
    var category models.Category

    if err := r.db.WithContext(ctx).
        First(&category, id).Error; err != nil {
        return nil, err
    }

    return &category, nil
}

// FindByRestaurantID retrieves categories for a restaurant.
func (r *PostgresCategoryRepo) FindByRestaurantID(
    ctx context.Context,
    restaurantID uint,
) ([]models.Category, error) {
    var categories []models.Category

    err := r.db.WithContext(ctx).
        Where("restaurant_id = ?", restaurantID).
        Preload("MenuItems").
        Find(&categories).Error

    return categories, err
}

// Update saves changes to an existing category.
func (r *PostgresCategoryRepo) Update(
    ctx context.Context,
    category *models.Category,
) error {
    return r.db.WithContext(ctx).Save(category).Error
}

// Delete removes a category belonging to a restaurant.
func (r *PostgresCategoryRepo) Delete(
    ctx context.Context,
    id uint,
    restaurantID uint,
) error {
    return r.db.WithContext(ctx).
        Where("id = ? AND restaurant_id = ?", id, restaurantID).
        Delete(&models.Category{}).Error
}
