package repository

import (
	"context"

	"Fhub/models"

	"gorm.io/gorm"
)

// RestaurantRepository defines database operations.
type RestaurantRepository interface {
	Create(ctx context.Context, restaurant *models.Restaurant) error
	FindByID(ctx context.Context, id uint) (*models.Restaurant, error)
	Update(ctx context.Context, restaurant *models.Restaurant) error
	UpdateImageURL(ctx context.Context, restaurantID uint, imageURL string) error
	ListActive(ctx context.Context, limit, offset int) ([]models.Restaurant, int64, error)
}

// PostgresRestaurantRepo implements RestaurantRepository.
type PostgresRestaurantRepo struct {
	db *gorm.DB
}

// NewRestaurantRepository creates a repository instance.
func NewRestaurantRepository(db *gorm.DB) RestaurantRepository {
	return &PostgresRestaurantRepo{db: db}
}

func (r *PostgresRestaurantRepo) Create(ctx context.Context, restaurant *models.Restaurant) error {
	return r.db.WithContext(ctx).Create(restaurant).Error
}

// FindByID retrieves a restaurant using its ID.
func (r *PostgresRestaurantRepo) FindByID(ctx context.Context, id uint) (*models.Restaurant, error) {
	var restaurant models.Restaurant
	if err := r.db.WithContext(ctx).First(&restaurant, id).Error; err != nil {
		return nil, err
	}
	return &restaurant, nil
}

// Update saves restaurant changes to the database.
func (r *PostgresRestaurantRepo) Update(ctx context.Context, restaurant *models.Restaurant) error {
	return r.db.WithContext(ctx).Save(restaurant).Error
}

func (r *PostgresRestaurantRepo) UpdateImageURL(ctx context.Context, restaurantID uint, imageURL string) error {
	return r.db.WithContext(ctx).Model(&models.Restaurant{}).
		Where("id = ?", restaurantID).
		Update("image_url", imageURL).Error
}

// ListActive returns one page of active restaurants incl. the total no. of active restaurants
func (r *PostgresRestaurantRepo) ListActive(ctx context.Context, limit, offset int) ([]models.Restaurant, int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&models.Restaurant{}).
		Where("status = ?", models.RestaurantStatusActive).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	var restaurants []models.Restaurant
	err = r.db.WithContext(ctx).
		Where("status = ?", models.RestaurantStatusActive).
		Order("name ASC, id ASC").
		Limit(limit).
		Offset(offset).
		Find(&restaurants).Error
	if err != nil {
		return nil, 0, err
	}

	return restaurants, total, nil

}
