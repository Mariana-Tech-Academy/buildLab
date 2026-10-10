package repository

import (
	"Fhub/models"

	"gorm.io/gorm"
)

type UserRepository interface {
	GetUserByID(id int) (models.User, error)
}

type PostgresUserRepo struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &PostgresUserRepo{db: db}
}

func (r *PostgresUserRepo) GetUserByID(id int) (models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	return user, err
}