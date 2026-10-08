package models 


type MenuItem struct {
	gorm.Model
	ID           int
	Name         string
	Description  string
	Price        float64
	ImageURL     string
	Available    bool
	RestaurantID int
	CategoryID   int
}