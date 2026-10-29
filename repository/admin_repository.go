package repository

import (
	"context"
	"strings"

	"Fhub/models"

	"gorm.io/gorm"
)

type AdminRepository interface {
	ListUsers(ctx context.Context, search string, limit, offset int) ([]models.User, int64, error)
}

type PostgresAdminRepo struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &PostgresAdminRepo{db: db}
}

// Gets a list of users and allows searching by name or email with pagination.
func (r *PostgresAdminRepo) ListUsers(
	ctx context.Context,
	search string,
	limit, offset int,
) ([]models.User, int64, error) {

	var users []models.User
	var total int64

	query := r.db.WithContext(ctx).Model(&models.User{})
	search = strings.TrimSpace(search)
	// Searches for users by first name, last name or email.
	if search != "" {
		searchTerm := "%" + search + "%"

		query = query.Where(
			"first_name ILIKE ? OR last_name ILIKE ? OR email ILIKE ?",
			searchTerm,
			searchTerm,
			searchTerm,
		)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.
		Order("id ASC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}
