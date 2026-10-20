package repository

import (
	"context"

	"Fhub/models"

	"gorm.io/gorm"
)

type CategoryRepository interface {
	FindByID(ctx context.Context, id uint) (*models.Category, error)
}

type PostgresCategoryRepo struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &PostgresCategoryRepo{db: db}
}

func (r *PostgresCategoryRepo) FindByID(ctx context.Context, id uint) (*models.Category, error) {
	var category models.Category
	if err := r.db.WithContext(ctx).First(&category, id).Error; err != nil {
		return nil, err
	}
	return &category, nil
}
