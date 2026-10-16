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
