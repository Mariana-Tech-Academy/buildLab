package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"Fhub/models"
	"Fhub/repository"

	"gorm.io/gorm"
)

var (
	ErrInvalidMenuItemInput = errors.New("invalid menu item input")
	ErrMenuItemNotFound     = errors.New("menu item not found")
	ErrCategoryNotFound     = errors.New("category not found")
	ErrRestaurantNotFound   = errors.New("restaurant not found")
)

// CreateMenuItemInput is the JSON body for creating a menu item.
// Price is minor units (pence or kobo). 500 means 5.00, not 500 pounds.
type CreateMenuItemInput struct {
	RestaurantID uint   `json:"restaurant_id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Price        int64  `json:"price"`
	ImageURL     string `json:"image_url"`
	CategoryID   uint   `json:"category_id"`
}

// UpdateMenuItemInput contains the fields a manager can edit.
// Pointer fields distinguish omitted fields from empty values.
// Price is minor units (pence or kobo). 1250 means 12.50.
type UpdateMenuItemInput struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Price       *int64  `json:"price"`
	ImageURL    *string `json:"image_url"`
	CategoryID  *uint   `json:"category_id"`
	Available   *bool   `json:"available"`
}

type MenuService interface {
	CreateMenuItem(
		ctx context.Context,
		actorID uint,
		role models.Role,
		input CreateMenuItemInput,
	) (*models.MenuItem, error)
	UpdateMenuItem(ctx context.Context, itemID, actorID uint, role models.Role, input UpdateMenuItemInput) (*models.MenuItem, error)
	DeleteMenuItem(ctx context.Context, itemID, actorID uint, role models.Role) error


    ListRestaurantMenu(ctx context.Context,restaurantID uint,categoryID *uint,) ([]models.MenuItem, error)
	ListMenuItem(ctx context.Context,itemID uint,) (*models.MenuItem, error)


}

type menuService struct {
	menuRepo          repository.MenuRepository
	categoryRepo      repository.CategoryRepository
	restaurantService RestaurantService
}

func NewMenuService(
	repo repository.MenuRepository,
	categories repository.CategoryRepository,
	restaurants RestaurantService,
) MenuService {
	return &menuService{
		menuRepo:          repo,
		categoryRepo:      categories,
		restaurantService: restaurants,
	
	
	}
}

func (s *menuService) CreateMenuItem(ctx context.Context, actorID uint, role models.Role,
	input CreateMenuItemInput,
) (*models.MenuItem, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	input.ImageURL = strings.TrimSpace(input.ImageURL)

	if input.Name == "" {
		return nil, fmt.Errorf("%w: menu item name is required", ErrInvalidMenuItemInput)
	}
	if input.Price <= 0 {
		return nil, fmt.Errorf("%w: price must be greater than zero", ErrInvalidMenuItemInput)
	}
	if input.RestaurantID == 0 {
		return nil, fmt.Errorf("%w: restaurant ID is required", ErrInvalidMenuItemInput)
	}
	if input.CategoryID == 0 {
		return nil, fmt.Errorf("%w: category ID is required", ErrInvalidMenuItemInput)
	}

	err := s.restaurantService.CheckOwnership(ctx, input.RestaurantID, actorID, role)
	if err != nil {
		return nil, err
	}

	if err := s.categoryExists(ctx, input.CategoryID, input.RestaurantID); err != nil {
		return nil, err
	}

	item := &models.MenuItem{
		RestaurantID: input.RestaurantID,
		CategoryID:   input.CategoryID,
		Name:         input.Name,
		Description:  input.Description,
		Price:        input.Price,
		ImageURL:     input.ImageURL,
		Available:    true,
	}

	err = s.menuRepo.CreateMenuItem(ctx, item)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (s *menuService) UpdateMenuItem(ctx context.Context, itemID, actorID uint, role models.Role, input UpdateMenuItemInput) (*models.MenuItem, error) {
	if role != models.RoleManager && role != models.RoleAdmin {
		return nil, ErrRestaurantForbidden
	}
	if itemID == 0 {
		return nil, fmt.Errorf("%w: item ID must be greater than zero", ErrInvalidMenuItemInput)
	}

	item, err := s.menuRepo.GetMenuItemByID(ctx, itemID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrMenuItemNotFound
	}
	if err != nil {
		return nil, err
	}

	err = s.restaurantService.CheckOwnership(ctx, item.RestaurantID, actorID, role)
	if err != nil {
		return nil, err
	}

	if input.Name == nil && input.Description == nil && input.Price == nil && input.ImageURL == nil && input.CategoryID == nil && input.Available == nil {
		return nil, fmt.Errorf("%w: at least one field must be provided", ErrInvalidMenuItemInput)
	}

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return nil, fmt.Errorf("%w: name cannot be empty", ErrInvalidMenuItemInput)
		}
		item.Name = name
	}
	if input.Description != nil {
		item.Description = strings.TrimSpace(*input.Description)
	}
	if input.Price != nil {
		if *input.Price <= 0 {
			return nil, fmt.Errorf("%w: the price must be greater than zero", ErrInvalidMenuItemInput)
		}
		item.Price = *input.Price
	}
	if input.ImageURL != nil {
		item.ImageURL = strings.TrimSpace(*input.ImageURL)
	}
	if input.CategoryID != nil {
		if *input.CategoryID == 0 {
			return nil, fmt.Errorf("%w: category ID must be greater than zero", ErrInvalidMenuItemInput)
		}
		if err := s.categoryExists(ctx, *input.CategoryID, item.RestaurantID); err != nil {
			return nil, err
		}
		item.CategoryID = *input.CategoryID
	}
	if input.Available != nil {
		item.Available = *input.Available
	}

	err = s.menuRepo.UpdateMenuItem(ctx, item)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrMenuItemNotFound
	}
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (s *menuService) DeleteMenuItem(ctx context.Context, itemID, actorID uint, role models.Role) error {
	if role != models.RoleManager && role != models.RoleAdmin {
		return ErrRestaurantForbidden
	}
	if itemID == 0 {
		return fmt.Errorf("%w: the item ID must be greater than zero", ErrInvalidMenuItemInput)
	}

	item, err := s.menuRepo.GetMenuItemByID(ctx, itemID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrMenuItemNotFound
	}
	if err != nil {
		return err
	}

	err = s.restaurantService.CheckOwnership(ctx, item.RestaurantID, actorID, role)
	if err != nil {
		return err
	}

	err = s.menuRepo.DeleteMenuItem(ctx, itemID, item.RestaurantID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrMenuItemNotFound
	}
	return err
}

func (s *menuService) categoryExists(ctx context.Context, categoryID, restaurantID uint) error {
	category, err := s.categoryRepo.FindByID(ctx, categoryID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrCategoryNotFound
	}
	if err != nil {
		return err
	}
	if category.RestaurantID != restaurantID {
		return ErrCategoryNotFound
	}
	return nil
}





func (s *menuService) ListRestaurantMenu(ctx context.Context,restaurantID uint,categoryID *uint,) ([]models.MenuItem, error) {
    return s.menuRepo.ListByRestaurant(ctx,restaurantID,categoryID,
    )
}

func (s *menuService) ListMenuItem(ctx context.Context,itemID uint,) (*models.MenuItem, error) {
    return s.menuRepo.GetMenuItemByID(ctx, itemID)
}