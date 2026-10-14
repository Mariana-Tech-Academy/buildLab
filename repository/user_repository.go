package repository

import (
	"context"

	"Fhub/models"

	"gorm.io/gorm"
)

type UserRepository interface {
	GetUserByID(ctx context.Context, id uint) (*models.User, error)
	UpdateProfile(ctx context.Context, id uint, firstName, lastName, phone string) (*models.User, error)
}

type PostgresUserRepo struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &PostgresUserRepo{db: db}
}

func (r *PostgresUserRepo) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *PostgresUserRepo) UpdateProfile(ctx context.Context, id uint, firstName, lastName, phone string) (*models.User, error) {
	user, err := r.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	err = r.db.WithContext(ctx).Model(user).Updates(map[string]any{
		"first_name": firstName,
		"last_name":  lastName,
		"phone":      phone,
	}).Error
	if err != nil {
		return nil, err
	}

	return user, nil
}
