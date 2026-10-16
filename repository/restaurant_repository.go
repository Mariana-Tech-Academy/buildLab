package repository

import (
	"Fhub/models"

	"gorm.io/gorm"
)

type RestaurantRepository interface {
	Create(restaurant *models.Restaurant) error
}

type PostgresRestaurantRepo struct {
	db *gorm.DB
}

func NewRestaurantRepository(db *gorm.DB) RestaurantRepository {
	return &PostgresRestaurantRepo{db: db}
}

func (r *PostgresRestaurantRepo) Create(restaurant *models.Restaurant) error {
	return r.db.Create(restaurant).Error
}
