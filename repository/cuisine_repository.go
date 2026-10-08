package repository 

import "gorm.io/gorm"

type CuisineRepository interface {
}

type PostgresCuisineRepo struct {
	db *gorm.DB
}

func NewCuisineRepository(db *gorm.DB) CuisineRepository {
	return &PostgresCuisineRepo{db: db}
}