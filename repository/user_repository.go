package repository

import (
	"Fhub/models"

	"gorm.io/gorm"
)

type UserRepository interface {
	EmailExists(email string) (bool, error)
	Create(user *models.User) error
}

type PostgresUserRepo struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &PostgresUserRepo{db: db}
}

func (r *PostgresUserRepo) EmailExists(email string) (bool, error) {
	var count int64
	err := r.db.Unscoped().Model(&models.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

func (r *PostgresUserRepo) Create(user *models.User) error {
	return r.db.Create(user).Error
}
