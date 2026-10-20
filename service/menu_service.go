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
	ErrCategoryNotFound     = errors.New("category not found")
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

// Creating a new menu item after validating the input and restaurant ownership.
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

	if _, err := s.categoryRepo.FindByID(ctx, input.CategoryID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
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
