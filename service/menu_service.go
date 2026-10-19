package service

import (
	"Fhub/models"
	"Fhub/repository"
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// Defines ErrMenuItemNotFound and ErrInvalidMenuItem. The handler recognises these and chooses a HTTP response.
var (
	ErrMenuItemNotFound = errors.New("menu item not found")
	ErrInvalidMenuItem  = errors.New("invalid menu item detaisl")
)

// UpdateMenuItemInput contains the fields a manager can edit. Pointer fields distinguish between not supplied/supplied from empty value
type UpdateMenuItemInput struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Price       *int64  `json:"price"`
	ImageURL    *string `json:"image_url"`
	CategoryID  *uint   `json:"category_id"`
}

// MenuService lists the update and delete operations available to the handler.
type MenuService interface {
	UpdateMenuItem(ctx context.Context, itemID, actorID uint, role models.Role, input UpdateMenuItemInput) (*models.MenuItem, error)
	DeleteMenuItem(ctx context.Context, itemID, actorID uint, role models.Role) error
}

// Stores the dependencies needed to perform menu item operations.
type menuService struct {
	menuRepo repository.MenuRepository
	// Provided the existing restaurant owenership check
	restaurantService RestaurantService
}

// Receives dependencies and stores them in a new menuService.
func NewMenuService(repo repository.MenuRepository, restaurants RestaurantService) MenuService {

	return &menuService{
		menuRepo:          repo,
		restaurantService: restaurants,
	}
}

// Allows only managers and admins. Other roles receive ErrRestaurantForbidden.
func (s *menuService) UpdateMenuItem(ctx context.Context, itemID, actorID uint, role models.Role, input UpdateMenuItemInput) (*models.MenuItem, error) {
	if role != models.RoleManager && role != models.RoleAdmin {
		return nil, ErrRestaurantForbidden
	}
	//Rejects item ID 0
	if itemID == 0 {
		return nil, fmt.Errorf("%w item ID must be greater than zero", ErrInvalidMenuItem)
	}
	//Find the existing item through the repository. If it does not exist, return an error.
	item, err := s.menuRepo.GetMenuItemByID(ctx, itemID)

	//Translate the databaase error into our service error.
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrMenuItemNotFound
	}
	// Stop if another database error occured
	if err != nil {
		return nil, err
	}
	// Check owenership using the restaurantID from the menu item and the actorID. If the actor is not the owner, return an error.
	err = s.restaurantService.CheckOwnership(ctx, item.RestaurantID, actorID, role)
	if err != nil {
		return nil, err
	}
	// Rejects requests containing none of the editable fields.
	if input.Name == nil && input.Description == nil && input.Price == nil && input.ImageURL == nil && input.CategoryID == nil {
		return nil, fmt.Errorf("%w: at least one field must be provided", ErrInvalidMenuItem)
	}
	// Trim surrounding spaces from a supplied name and reject a blank name.
	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return nil, fmt.Errorf("%w: name cannot be empty", ErrInvalidMenuItem)
		}
		item.Name = name
	}
	// Updates a supplied description. An empty description is allowed.
	if input.Description != nil {
		item.Description = strings.TrimSpace(*input.Description)
	}
	// Rejects prices of zero or less, then assigns a valid price. 1250 represents £12.50.
	if input.Price != nil {
		if *input.Price <= 0 {
			return nil, fmt.Errorf("%w: the price must be greater than zero", ErrInvalidMenuItem)
		}
		item.Price = *input.Price
	}
	// Updatesor clears the image URL. This stores a string
	if input.ImageURL != nil {
		item.ImageURL = strings.TrimSpace(*input.ImageURL)
	}
	// Rejects category ID 0, then assigns a positive category ID.
	if input.CategoryID != nil {
		if *input.CategoryID == 0 {
			return nil, fmt.Errorf("%w: category ID must be greater than zero", ErrInvalidMenuItem)
		}
		item.CategoryID = *input.CategoryID
	}
	// Asks the repository to save, handles errors and returns the updated item.
	err = s.menuRepo.UpdateMenuItem(ctx, item)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrMenuItemNotFound
	}
	if err != nil {
		return nil, err
	}
	return item, nil
}

// Allows only managers and admins.
func (s *menuService) DeleteMenuItem(ctx context.Context, itemID, actorID uint, role models.Role) error {
	if role != models.RoleManager && role != models.RoleAdmin {
		return ErrRestaurantForbidden
	}
	if itemID == 0 { // Rejects item ID 0.
		return fmt.Errorf("%w: the item ID must be greater than zero", ErrInvalidMenuItem)
	}
	// Finds the item and handles lookup errors
	item, err := s.menuRepo.GetMenuItemByID(ctx, itemID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrMenuItemNotFound
	}
	if err != nil {
		return err
	}
	// Check that the user may manage this restaurant
	err = s.restaurantService.CheckOwnership(ctx, item.RestaurantID, actorID, role)
	if err != nil {
		return err
	}
	// Calls the repository’s soft-delete method with the item ID and its stored restaurant ID. Converts not-found errors.
	err = s.menuRepo.DeleteMenuItem(ctx, itemID, item.RestaurantID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrMenuItemNotFound
	}
	// nil means success otherwise return the database error
	return err
}
