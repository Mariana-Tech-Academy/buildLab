package models

import (
	"time"

	"gorm.io/gorm"
)

// Name and UnitPrice are snapshots taken at checkout so later menu edits never change past orders.
type OrderItem struct {
	gorm.Model
	OrderID    uint     `gorm:"not null;index" json:"order_id"`
	MenuItemID uint     `gorm:"not null;index" json:"menu_item_id"`
	MenuItem   MenuItem `json:"-"`
	Name       string   `gorm:"not null" json:"name"`
	UnitPrice  int64    `gorm:"not null" json:"unit_price"`
	Quantity   int      `gorm:"not null;check:quantity > 0" json:"quantity"`
	LineTotal  int64    `gorm:"not null" json:"line_total"`
}

// OrderStatusChange records each time an admin changes an order's status (ADM-07),
// so we can see who made the change, what it changed from and to, and when.
type OrderStatusChange struct {
	ID         uint        `gorm:"primaryKey" json:"id"`
	OrderID    uint        `gorm:"not null;index" json:"order_id"`
	FromStatus OrderStatus `gorm:"type:varchar(30);not null" json:"from_status"`
	ToStatus   OrderStatus `gorm:"type:varchar(30);not null" json:"to_status"`
	ChangedBy  uint        `gorm:"not null;index" json:"changed_by"`
	CreatedAt  time.Time   `json:"created_at"`
}
