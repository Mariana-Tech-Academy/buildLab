package models

import "gorm.io/gorm"

type Category struct {
	gorm.Model
	RestaurantID uint       `gorm:"not null;uniqueIndex:idx_restaurant_category_lower" json:"restaurant_id"`
	Name         string     `gorm:"not null;uniqueIndex:idx_restaurant_category_lower,expression:LOWER(name)" json:"name"`
	MenuItems    []MenuItem `json:"menu_items,omitempty"`
}








// type Category struct {
//     ID   uint   `gorm:"primaryKey" json:"id"`
//     Name string `gorm:"not null" json:"name"`
// }