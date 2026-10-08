package models

import (
	"time"

	"gorm.io/gorm"
)
type Cart struct {
	gorm.Model
ID           int
UserID       int
User         User
RestaurantID int
Restaurant   Restaurant
Items        []CartItem
CreatedAt    time.Time
UpdatedAt    time.Time
}