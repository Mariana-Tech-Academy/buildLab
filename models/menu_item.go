package models

type Menu_Item struct {
	gorm.Model
ID int 
Name string 
Description string 
Price float64 
Image string
CategoryID int 
RestaurantID int 
Availability bool
CreatedAt int
UpdatedAt int
}

