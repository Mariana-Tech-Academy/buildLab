package repository

import (
	"context"
	"strings"

	"Fhub/models"

	"gorm.io/gorm"
)

// RestaurantRepository defines database operations.
type RestaurantRepository interface {
	Create(ctx context.Context, restaurant *models.Restaurant) error
	FindByID(ctx context.Context, id uint) (*models.Restaurant, error)
	Update(ctx context.Context, restaurant *models.Restaurant) error
	UpdateImageURL(ctx context.Context, restaurantID uint, imageURL string) error
	ListActive(ctx context.Context) ([]models.Restaurant, error)
	SearchActive(ctx context.Context, query string) ([]models.Restaurant, error)
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

// ListActive retrieves only active restaurants for customer listings.
func (r *PostgresRestaurantRepo) ListActive(ctx context.Context) ([]models.Restaurant, error) {
	var restaurants []models.Restaurant

	err := r.db.WithContext(ctx).
		Where("status = ?", models.RestaurantStatusActive).
		Find(&restaurants).Error

	return restaurants, err
}
// SearchActive searches active restaurants by name or location.
func (r *PostgresRestaurantRepo) SearchActive(
	ctx context.Context,
	query string,
) ([]models.Restaurant, error) {
	var restaurants []models.Restaurant
	searchTerm := "%" + strings.TrimSpace(query) + "%"

	err := r.db.WithContext(ctx).
		Where("status = ?", models.RestaurantStatusActive).
		Where(
			"LOWER(name) LIKE LOWER(?) OR LOWER(address) LIKE LOWER(?) OR LOWER(city) LIKE LOWER(?) OR LOWER(state) LIKE LOWER(?)",
			searchTerm,
			searchTerm,
			searchTerm,
			searchTerm,
		).
		Find(&restaurants).Error

	return restaurants, err
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
