package models 
type OrderItem struct {
	gorm.Model
ID unit
OrderID uint
MenuItemID uint
Name string
Quantity int
CreatedAt time.Time
MenuItem *MenuItem
Total    int64

}