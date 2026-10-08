package models

import "gorm.io/gorm"

type MenuItem struct {
	gorm.Model
	Name         string
	Description  string
	Price        float64
	ImageURL     string
	Available    bool
	RestaurantID int
	CategoryID   int
}
