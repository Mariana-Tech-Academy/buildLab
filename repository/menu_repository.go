
package repository

import (
	"errors"

	"Fhub/models"
	"gorm.io/gorm"
)

var ErrMenuItemNotFound = errors.New("menu item not found")
var ErrMenuItemUnavailable = errors.New("this item is currently unavailable")
var ErrMenuItemForbidden = errors.New("you do not manage this restaurant")

type MenuSearchFilters struct {
    Name         string
    CategoryID   *uint
    RestaurantID *uint
    MinPrice     *int64
    MaxPrice     *int64
    Available    *bool
}

type MenuRepository interface {
	Get(id uint) (*models.MenuItem, error)
	List(restaurantID uint) ([]models.MenuItem, error)
	SetAvailability(id, managerID uint, isAdmin, available bool) (*models.MenuItem, error)
	CheckAvailable(id uint) error
	Search(filters MenuSearchFilters) ([]models.MenuItem, error)
}

type PostgresMenuRepo struct {
	db *gorm.DB
}

func NewMenuRepository(db *gorm.DB) MenuRepository {
	return &PostgresMenuRepo{db: db}
}

func (r *PostgresMenuRepo) Get(id uint) (*models.MenuItem, error) {
	var item models.MenuItem

	if err := r.db.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMenuItemNotFound
		}
		return nil, err
	}

	return &item, nil
}

func (r *PostgresMenuRepo) List(restaurantID uint) ([]models.MenuItem, error) {
	var items []models.MenuItem

	err := r.db.Where("restaurant_id = ?", restaurantID).
		Find(&items).Error

	if err != nil {
		return nil, err
	}

	return items, nil
}

func (r *PostgresMenuRepo) SetAvailability(
	id, managerID uint,
	isAdmin, available bool,
) (*models.MenuItem, error) {

	var item models.MenuItem

	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Raw(
			"SELECT * FROM menu_items WHERE id = ? AND deleted_at IS NULL FOR UPDATE",
			id,
		).Scan(&item).Error; err != nil {
			return err
		}

		if item.ID == 0 {
			return ErrMenuItemNotFound
		}

		if !isAdmin {
			var count int64

			if err := tx.Model(&models.Restaurant{}).
				Where("id = ? AND manager_id = ?", item.RestaurantID, managerID).
				Count(&count).Error; err != nil {
				return err
			}

			if count == 0 {
				return ErrMenuItemForbidden
			}
		}

		if err := tx.Model(&item).
			Update("available", available).Error; err != nil {
			return err
		}

		item.Available = available
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *PostgresMenuRepo) CheckAvailable(id uint) error {
	item, err := r.Get(id)

	if err != nil {
		return err
	}

	if !item.Available {
		return ErrMenuItemUnavailable
	}

	return nil
}


func (r *PostgresMenuRepo) Search(filters MenuSearchFilters) ([]models.MenuItem, error) {
    var items []models.MenuItem

    query := r.db.Model(&models.MenuItem{})

    if filters.Name != "" {
        query = query.Where("name ILIKE ?", "%"+filters.Name+"%")
    }

    if filters.CategoryID != nil {
        query = query.Where("category_id = ?", *filters.CategoryID)
    }

    if filters.RestaurantID != nil {
        query = query.Where("restaurant_id = ?", *filters.RestaurantID)
    }

    if filters.MinPrice != nil {
        query = query.Where("price >= ?", *filters.MinPrice)
    }

    if filters.MaxPrice != nil {
        query = query.Where("price <= ?", *filters.MaxPrice)
    }

    if filters.Available != nil {
        query = query.Where("available = ?", *filters.Available)
    }

    err := query.Find(&items).Error

    return items, err
}


