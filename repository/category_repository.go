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
	NameExists(ctx context.Context, restaurantID uint, name string) (bool, error)
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

	if err := r.db.WithContext(ctx).First(&category, id).Error; err != nil {
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
		Preload("MenuItems", func(db *gorm.DB) *gorm.DB {
			return db.Where("available = ?", true).Order("name ASC, id ASC")
		}).
		Order("name ASC, id ASC").
		Find(&categories).Error

	return categories, err
}

// NameExists reports whether this restaurant already has a category with the same name, ignoring case.
func (r *PostgresCategoryRepo) NameExists(ctx context.Context, restaurantID uint, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Category{}).
		Where("restaurant_id = ? AND LOWER(name) = LOWER(?)", restaurantID, name).
		Count(&count).Error
	return count > 0, err
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
