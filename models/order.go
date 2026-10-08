package models 

type Order struct  {
	gorm.Model
OrderNumber string
CustomerID  int
Customer    User 
RestaurantID  int 
Status        string
SubTotal     float64
DeliveryFee  float64
Total        float64

DeliveryName string
DeliveryPhone string 
DeliveryAddress string
DeliveryCity string
DeliveryState string
DeliveryNotes string

OrderItems [] OrderItem


}