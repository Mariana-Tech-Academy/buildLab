package models

import "time"

// OrderStatusChange records each admin status change: who changed the order, from which status, and when.
type OrderStatusChange struct {
	ID         uint        `gorm:"primaryKey" json:"id"`
	OrderID    uint        `gorm:"not null;index" json:"order_id"`
	FromStatus OrderStatus `gorm:"type:varchar(30);not null" json:"from_status"`
	ToStatus   OrderStatus `gorm:"type:varchar(30);not null" json:"to_status"`
	ChangedBy  uint        `gorm:"not null;index" json:"changed_by"`
	CreatedAt  time.Time   `json:"created_at"`
}
