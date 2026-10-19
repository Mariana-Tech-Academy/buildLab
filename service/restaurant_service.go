package service

import (
	"Fhub/models"
	"Fhub/repository"
	"context"
	"errors"
)

//ErrRestaurantForbidden names means the user does not manage this restaurant.
var ErrRestaurantForbidden = errors.New("You are not authorised to manage this restaurant")

// RestaurantService defines the restaurant related business operations available to other services.
type RestaurantService interface {
	CheckOwnership(ctx context.Context, restaurantID uint, actorID uint, role models.Role) error
}

// restaurantService holds our restaurant repository.
//This will allow the service to request information from PostgreSQL.
type restaurantService struct {
	restaurantRepo repository.RestaurantRepository
}

// NewRestaurantService creates a new restaurant service.
//It receives the restaurant repository and stores it in the struct.
func NewRestaurantService(repo repository.RestaurantRepository) RestaurantService {
	return &restaurantService{restaurantRepo: repo}
}

//CheckOwnership verifies that the authenticated user manages the restaurant they want to modify.
func (s *restaurantService) CheckOwnership(ctx context.Context, restaurantID uint, actorID uint, role models.Role) error {
	
	//We ask the repository to find the restaurant by its ID.
	restaurant, err := s.restaurantRepo.GetRestaurantByID(ctx, restaurantID)
	
	// If the restaurant cannot be found or the database fails, return the error and stop the operation.
	// Return the erro and stop the operation.
	if err != nil {
		return err
	}
	//We compare the restaurant maanger's ID with the logged-in user's ID.
	//If they are different, the user is not authorised.
	if restaurant.ManagerID != actorID {
		return ErrRestaurantForbidden
	}
	// If both IDs match, the user is allowed to manage the restaurant.
	// The return nil means the ownership check succeded.
	return nil
}


