package models

import "gorm.io/gorm"

type Category struct {
    gorm.Model

    RestaurantID uint `gorm:"not null;uniqueIndex:idx_restaurant_category" json:"restaurant_id"`

    Name string `gorm:"not null;uniqueIndex:idx_restaurant_category" json:"name"`

    MenuItems []MenuItem `json:"menu_items,omitempty"`
}