package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"Fhub/models"
	"Fhub/repository"

	"gorm.io/gorm"
)

var (
	ErrRestaurantNotFound  = errors.New("restaurant not found")
	ErrRestaurantForbidden = errors.New("you do not own this restaurant")
	ErrInvalidRestaurant   = errors.New("invalid restaurant details")
)

type CreateRestaurantInput struct {
	ManagerID   uint
	Name        string
	Description string
	Address     string
	Phone       string
	Email       string
	OpeningTime string
	ClosingTime string
}

// UpdateRestaurantInput contains only fields managers can update.
// Pointer fields allow us to distinguish omitted fields from empty values.
type UpdateRestaurantInput struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Address     *string `json:"address"`
	City        *string `json:"city"`
	State       *string `json:"state"`
	Phone       *string `json:"phone"`
	Email       *string `json:"email"`
	OpeningTime *string `json:"opening_time"`
	ClosingTime *string `json:"closing_time"`
	ImageURL    *string `json:"-"`
}

type RestaurantService interface {
	Create(input CreateRestaurantInput) (*models.Restaurant, error)

	UpdateRestaurant(
		restaurantID uint,
		managerID uint,
		input UpdateRestaurantInput,
	) (*models.Restaurant, error)

	CheckOwnership(
		restaurantID uint,
		managerID uint,
	) error

	UpdateRestaurantImage(
		restaurantID uint,
		managerID uint,
		imageURL string,
	) error
}

type restaurantService struct {
	restaurantRepo repository.RestaurantRepository
}

func NewRestaurantService(
	repo repository.RestaurantRepository,
) RestaurantService {
	return &restaurantService{restaurantRepo: repo}
}

func (s *restaurantService) Create(input CreateRestaurantInput) (*models.Restaurant, error) { // Create a new restaurant with the provided input and return the created restaurant or an error
	restaurant := &models.Restaurant{ // Create a new Restaurant model instance with the provided input
		ManagerID:   input.ManagerID,
		Name:        input.Name,
		Description: input.Description,
		Address:     input.Address,
		Phone:       input.Phone,
		Email:       strings.ToLower(input.Email),
		OpeningTime: input.OpeningTime,
		ClosingTime: input.ClosingTime,
		Status:      models.RestaurantStatusPending, // Set the initial status to pending approval
	}

	if err := s.restaurantRepo.Create(restaurant); err != nil {
		return nil, err
	}

	return restaurant, nil
}

func (s *restaurantService) CheckOwnership(
	restaurantID uint,
	managerID uint,
) error {

	restaurant, err := s.restaurantRepo.FindByID(restaurantID)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrRestaurantNotFound
	}

	if err != nil {
		return err
	}

	if restaurant.ManagerID != managerID {
		return ErrRestaurantForbidden
	}

	return nil
}

func (s *restaurantService) UpdateRestaurant(
	restaurantID uint,
	managerID uint,
	input UpdateRestaurantInput,
) (*models.Restaurant, error) {

	// Step 1: Find the restaurant.
	restaurant, err := s.restaurantRepo.FindByID(restaurantID)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRestaurantNotFound
	}

	if err != nil {
		return nil, err
	}

	// Step 2: Check ownership.
	if restaurant.ManagerID != managerID {
		return nil, ErrRestaurantForbidden
	}

	// Step 3: Update only the fields supplied.
	if input.Name != nil {
		if *input.Name == "" {
			return nil, fmt.Errorf("%w: name is required", ErrInvalidRestaurant)
		}
		restaurant.Name = *input.Name
	}

	if input.Description != nil {
		restaurant.Description = *input.Description
	}

	if input.Address != nil {
		if *input.Address == "" {
			return nil, fmt.Errorf("%w: address is required", ErrInvalidRestaurant)
		}
		restaurant.Address = *input.Address
	}

	if input.City != nil {
		restaurant.City = *input.City
	}

	if input.State != nil {
		restaurant.State = *input.State
	}

	if input.Phone != nil {
		restaurant.Phone = *input.Phone
	}

	if input.Email != nil {
		restaurant.Email = *input.Email
	}

	if input.OpeningTime != nil {
		restaurant.OpeningTime = *input.OpeningTime
	}

	if input.ClosingTime != nil {
		restaurant.ClosingTime = *input.ClosingTime
	}

	if input.ImageURL != nil {
		restaurant.ImageURL = *input.ImageURL
	}

	// Step 4: Validate opening and closing times.
	opening, err := time.Parse("15:04", restaurant.OpeningTime)
	if err != nil {
		return nil, fmt.Errorf("%w: opening time must be HH:MM", ErrInvalidRestaurant)
	}

	closing, err := time.Parse("15:04", restaurant.ClosingTime)
	if err != nil {
		return nil, fmt.Errorf("%w: closing time must be HH:MM", ErrInvalidRestaurant)
	}

	// This assumes restaurants open and close on the same day.
	// Confirm overnight-hours behaviour with your team.
	if !closing.After(opening) {
		return nil, fmt.Errorf("%w: closing time must be after opening time", ErrInvalidRestaurant)
	}

	// Step 5: Save changes.
	if err := s.restaurantRepo.Update(restaurant); err != nil {
		return nil, err
	}

	return restaurant, nil
}

func (s *restaurantService) UpdateRestaurantImage(
	restaurantID uint,
	managerID uint,
	imageURL string,
) error {

	// Step 1: Confirm the manager owns the restaurant.
	err := s.CheckOwnership(restaurantID, managerID)
	if err != nil {
		return err
	}

	// Step 2: Update the image URL in the database.
	err = s.restaurantRepo.UpdateImageURL(
		restaurantID,
		imageURL,
	)
	if err != nil {
		return err
	}

	return nil
}
