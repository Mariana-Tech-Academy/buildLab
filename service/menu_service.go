package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"Fhub/models"
	"Fhub/repository"
)

var ErrInvalidMenuItemInput = errors.New("invalid menu item input")

// when a restaurant manager creates a menu item.
type CreateMenuItemInput struct {
	RestaurantID uint   `json:"restaurant_id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Price        int64  `json:"price"`
	ImageURL     string `json:"image_url"`
	CategoryID   uint   `json:"category_id"`
}

// Define the create operation that menu service must provide.
type MenuService interface {
	CreateMenuItem(
		ctx context.Context,
		actorID uint,
		role models.Role,
		input CreateMenuItemInput,
	) (*models.MenuItem, error)
}

// Hold the dependencies used by the menu service.
type menuService struct {
	menuRepo          repository.MenuRepository
	restaurantService RestaurantService
}

func NewMenuService(repo repository.MenuRepository, restaurants RestaurantService) MenuService {
	return &menuService{menuRepo: repo, restaurantService: restaurants}
}

// Creating a new menu item after validating the input and restaurant ownership.
func (s *menuService) CreateMenuItem(ctx context.Context, actorID uint, role models.Role,
	input CreateMenuItemInput,
) (*models.MenuItem, error) {
	//First step, is to remove spaces around the name.
	input.Name = strings.TrimSpace(input.Name)
	//Second step: Check that a name was provided
	if input.Name == "" {
		return nil, fmt.Errorf(
			"%w: menu item name is required", ErrInvalidMenuItemInput)
	}
	//Third Step is to check that price is positive
	if input.Price <= 0 {
		return nil, fmt.Errorf("%w: price must be greater than zero", ErrInvalidMenuItemInput)
	}

	// Fourth step is to check the restaurant ID
	if input.RestaurantID == 0 {
		return nil, fmt.Errorf("%w: restaurant ID is required", ErrInvalidMenuItemInput)
	}

	// Fifth step requires a category ID.
	if input.CategoryID == 0 {
		return nil, fmt.Errorf("%w: category ID is required", ErrInvalidMenuItemInput)
	}

	// Sixth step is to check that the authenticated user is allowed to manage this restaurant
	err := s.restaurantService.CheckOwnership(ctx, input.RestaurantID, actorID, role)
	if err != nil {
		return nil, err
	}

	// Seventh step is to prepare the menu item for saving
	item := &models.MenuItem{
		RestaurantID: input.RestaurantID,
		CategoryID:   input.CategoryID,
		Name:         input.Name,
		Description:  input.Description,
		Price:        input.Price,
		ImageURL:     input.ImageURL,
		Available:    true,
	}

	// Eighth step is to save the item in PostgreSQL
	err = s.menuRepo.CreateMenuItem(ctx, item)
	if err != nil {
		return nil, err
	}
	// Ninth step is return the newly created item.
	return item, nil
}
