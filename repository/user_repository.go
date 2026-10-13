package repository

import (
	"context"
	"fmt"

	"Fhub/models"

	"gorm.io/gorm"
)

type UserRepository interface {
	GetUserByID(id int) (models.User, error)

	UpdateProfile(
		ctx context.Context,
		id int64,
		firstName, lastName, phone string,
	) (*models.User, error)
	 
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

func (r *PostgresUserRepo) UpdateProfile(
	ctx context.Context,           //
	id int64,
	firstName, lastName, phone string,
) (*models.User, error) {    // Use context for cancellation and timeouts
	var user models.User
	
	result := r.db.WithContext(ctx).   //passed down the content 
		Model(&models.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"first_name": firstName,
			"last_name":  lastName,
			"phone":      phone,
		})

	if result.Error != nil {
		return nil, fmt.Errorf("update profile: %w", result.Error)
	}

	if result.RowsAffected == 0 {    // Check if the user exists
		var count int64
		err := r.db.WithContext(ctx).    // Use WithContext to ensure the context is passed down
			Model(&models.User{}).      
			Where("id = ?", id).
			Count(&count).Error 

		if err != nil {
			return nil, fmt.Errorf("check user: %w", err)
		}
		if count == 0 {
			return nil, gorm.ErrRecordNotFound
		}
	}

	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {     // Fetch the updated user profile
		return nil, fmt.Errorf("fetch updated profile: %w", err)
	}

	return &user, nil
}





