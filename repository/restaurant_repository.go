package repository

import (
	"context"
	"Fhub/models"
	"gorm.io/gorm"

)

// RestaurantRepository defines the database operations that other part of our application can use.
type RestaurantRepository interface {
	GetRestaurantByID(ctx context.Context, restaurantID uint) (*models.Restaurant, error)
}

//PostgresRestaurantRepo stores our database connection.
//Then we use this connection to communicate with PostgreSQL.
type PostgresRestaurantRepo struct {
	db *gorm.DB
}

//NewRestaurantRepository creates a new restaurant repository.
//it Receives the database connection and stores it inside our struct.
func NewRestaurantRepository(db *gorm.DB) RestaurantRepository {
	return &PostgresRestaurantRepo{db: db}
}

//GetRestaurantByID searches PostgreSQL for a restaurant using its ID.
func (r *PostgresRestaurantRepo) GetRestaurantByID(ctx context.Context, restaurantID uint) (*models.Restaurant, error){
	
	//Create an empty restaurant variable to hold the database result.
	var restaurant models.Restaurant

	//We use GORM to seach for the restaurant.
	// WithContext(ctx) allows the database operation to follow the request context.
	//First searches for the restaurant using its primary key.
	//Error contains any erro that occured during the search.
	err := r.db.WithContext(ctx).First(&restaurant,restaurantID).Error
	
	// If something goes wrong, return nil and the error.
	if err != nil {
		return nil, err
	}

	// If the restaurant is found, it will return its details and nil means that there was no error.
	return &restaurant, nil
}
