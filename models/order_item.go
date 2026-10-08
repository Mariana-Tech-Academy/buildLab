package models

import (
	"gorm.io/gorm"
)

type OrderItem struct {
	gorm.Model

	OrderID    uint
	MenuItemID uint
	Name       string
	Quantity   int
	MenuItem   *MenuItem
	Total      int64
}
