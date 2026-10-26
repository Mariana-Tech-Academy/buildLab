package repository

import (
	"Fhub/models"
	"errors"

	"gorm.io/gorm"
)
// 1. ErrCartRestaurantMismatch is returned when a user tries to add an item from a different restaurant to their cart.
var ErrCartRestaurantMismatch = errors.New(
	"cart contains items from another restaurant; clear the cart before adding this item",
)

type CartRepository interface {
	AddItem(userID uint, menuItemID uint, quantity int) error
}

type PostgresCartRepo struct {
	db *gorm.DB
}

func NewCartRepository(db *gorm.DB) CartRepository {
	return &PostgresCartRepo{db: db}
}

// 2.AddItem adds an item to the user's cart. If the user does not have a cart,
//  it creates one. If the cart already contains items from a different restaurant,
// it returns ErrCartRestaurantMismatch.
func (r *PostgresCartRepo) AddItem(userID uint, menuItemID uint, quantity int) error {
	var item models.MenuItem
	if err := r.db.First(&item, menuItemID).Error; err != nil {
		return err
	}
// Check if the user already has a cart
	var cart models.Cart
	err := r.db.Where("user_id = ?", userID).First(&cart).Error

// If the cart does not exist, create a new one with the restaurant ID of the item being added.
	if errors.Is(err, gorm.ErrRecordNotFound) {
		cart = models.Cart{
			UserID:       userID,
			RestaurantID: &item.RestaurantID,
		}
		if err := r.db.Create(&cart).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		if cart.RestaurantID != nil && *cart.RestaurantID != item.RestaurantID {
			return ErrCartRestaurantMismatch
		}
// If the cart exists but has no restaurant ID, set it to the restaurant ID of the item being added.
		if cart.RestaurantID == nil {
			cart.RestaurantID = &item.RestaurantID
			if err := r.db.Save(&cart).Error; err != nil {
				return err
			}
		}
	}
// Add the item to the cart
	return r.db.Create(&models.CartItem{
		CartID:     cart.ID,
		MenuItemID: menuItemID,
		Quantity:   quantity,
	}).Error
}
