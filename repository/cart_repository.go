package repository

import "gorm.io/gorm"

type CartRepository interface {
}

type PostgresCartRepo struct {
	db *gorm.DB
}

func NewCartRepository(db *gorm.DB) AdminRepository {
	return &PostgresAdminRepo{db: db}
}