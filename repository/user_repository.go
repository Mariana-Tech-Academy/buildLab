package repository

import (
	"context"

	"Fhub/models"

	"gorm.io/gorm"
)

type UserRepository interface {
	GetUserByID(ctx context.Context, id uint) (*models.User, error)
	UpdateProfile(ctx context.Context, id uint, firstName, lastName, phone string) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	EmailExists(ctx context.Context, email string) (bool, error)
	CreateWithCart(ctx context.Context, user *models.User) error
	UpdatePassword(ctx context.Context, id uint, newHash string) error
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

func (r *PostgresUserRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *PostgresUserRepo) EmailExists(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Unscoped().Model(&models.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

// CreateWithCart creates the user and their empty cart in one transaction.
func (r *PostgresUserRepo) CreateWithCart(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		return tx.Create(&models.Cart{UserID: user.ID}).Error
	})
}

func (r *PostgresUserRepo) UpdatePassword(ctx context.Context, id uint, newHash string) error {
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", id).
		Update("password", newHash).Error
}
