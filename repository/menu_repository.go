package repository

import (
	"context"

	"Fhub/models"

	"gorm.io/gorm"
)

// Menu Repository defines database operation for items of the menu
type MenuRepository interface {
	CreateMenuItem(ctx context.Context, item *models.MenuItem) error
}

// PostgresMenuRepo stores the connection of the database.
type PostgresMenuRepo struct {
	db *gorm.DB
}

// The func name creates a new menu repository
func NewMenuRepository(db *gorm.DB) MenuRepository {
	return &PostgresMenuRepo{db: db}
}

// This will allow to save a menu item in postgreSQL with the use of "Create"
func (r *PostgresMenuRepo) CreateMenuItem(ctx context.Context, item *models.MenuItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}
