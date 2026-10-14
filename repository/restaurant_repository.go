
package repository

import (
    "Fhub/models"

    "gorm.io/gorm"
)

// RestaurantRepository defines database operations.
type RestaurantRepository interface {
    FindByID(id uint) (*models.Restaurant, error)
    Update(restaurant *models.Restaurant) error
    UpdateImageURL(
    restaurantID uint,
    imageURL string,
) error
}

// PostgresRestaurantRepo implements RestaurantRepository.
type PostgresRestaurantRepo struct {
    db *gorm.DB
}

// NewRestaurantRepository creates a repository instance.
func NewRestaurantRepository(
    db *gorm.DB,
) RestaurantRepository {
    return &PostgresRestaurantRepo{db: db}
}

// FindByID retrieves a restaurant using its ID.
func (r *PostgresRestaurantRepo) FindByID(
    id uint,
) (*models.Restaurant, error) {

    var restaurant models.Restaurant

    err := r.db.First(&restaurant, id).Error

    if err != nil {
        return nil, err
    }

    return &restaurant, nil
}

// Update saves restaurant changes to the database.
func (r *PostgresRestaurantRepo) Update(
    restaurant *models.Restaurant,
) error {
    return r.db.Save(restaurant).Error
}


func (r *PostgresRestaurantRepo) UpdateImageURL(
    restaurantID uint,
    imageURL string,
) error {
    return r.db.Model(&models.Restaurant{}).
        Where("id = ?", restaurantID).
        Update("image_url", imageURL).Error
}
