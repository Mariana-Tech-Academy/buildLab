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
	GetMenuItemByID(ctx context.Context, itemID uint) (*models.MenuItem, error)
	ListRestaurantMenu(ctx context.Context, restaurantID uint, categoryID *uint) ([]models.MenuItem, error)
	SetAvailability(ctx context.Context, itemID, actorID uint, role models.Role, available bool) (*models.MenuItem, error)
	Search(ctx context.Context, input MenuSearchInput) (*MenuSearchPage, error)
}

const (
	defaultMenuPageSize = 10
	maxMenuPageSize     = 50
)

// MenuSearchInput is a customer menu search.
// MinPrice and MaxPrice are minor units: 500 means 5.00.
type MenuSearchInput struct {
	Name         string
	CategoryID   *uint
	RestaurantID *uint
	MinPrice     *int64
	MaxPrice     *int64
	Available    *bool
	Page         int
	Limit        int
}

// MenuSearchPage is one page of menu search results.
type MenuSearchPage struct {
	Items []models.MenuItem
	Page  int
	Limit int
	Total int64
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

// GetMenuItemByID retrieves a menu item by its ID.
func (s *menuService) GetMenuItemByID(
	ctx context.Context,
	itemID uint,
) (*models.MenuItem, error) {
	if itemID == 0 {
		return nil, fmt.Errorf(
			"%w: item ID must be greater than zero",
			ErrInvalidMenuItemInput,
		)
	}

	item, err := s.menuRepo.GetMenuItemByID(ctx, itemID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrMenuItemNotFound
	}
	if err != nil {
		return nil, err
	}
	if _, err := s.restaurantService.GetActiveRestaurant(ctx, item.RestaurantID); err != nil {
		if errors.Is(err, ErrRestaurantNotFound) {
			return nil, ErrMenuItemNotFound
		}
		return nil, err
	}
	return item, nil
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

func (s *menuService) ListRestaurantMenu(ctx context.Context, restaurantID uint, categoryID *uint) ([]models.MenuItem, error) {
	if _, err := s.restaurantService.GetActiveRestaurant(ctx, restaurantID); err != nil {
		return nil, err
	}
	return s.menuRepo.ListByRestaurant(ctx, restaurantID, categoryID)
}

// SetAvailability marks one menu item available or unavailable for the restaurant's manager or an admin.
func (s *menuService) SetAvailability(ctx context.Context, itemID, actorID uint, role models.Role, available bool) (*models.MenuItem, error) {
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
	if err := s.restaurantService.CheckOwnership(ctx, item.RestaurantID, actorID, role); err != nil {
		return nil, err
	}

	item.Available = available
	if err := s.menuRepo.UpdateMenuItem(ctx, item); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMenuItemNotFound
		}
		return nil, err
	}
	return item, nil
}

// Search returns one page of menu items from active restaurants.
func (s *menuService) Search(ctx context.Context, input MenuSearchInput) (*MenuSearchPage, error) {
	if input.MinPrice != nil && *input.MinPrice < 0 {
		return nil, fmt.Errorf("%w: min_price cannot be negative", ErrInvalidMenuItemInput)
	}
	if input.MaxPrice != nil && *input.MaxPrice < 0 {
		return nil, fmt.Errorf("%w: max_price cannot be negative", ErrInvalidMenuItemInput)
	}
	if input.MinPrice != nil && input.MaxPrice != nil && *input.MinPrice > *input.MaxPrice {
		return nil, fmt.Errorf("%w: min_price cannot exceed max_price", ErrInvalidMenuItemInput)
	}

	page, limit, offset, err := menuPageBounds(input.Page, input.Limit)
	if err != nil {
		return nil, err
	}
	items, total, err := s.menuRepo.Search(ctx, repository.MenuSearchFilter{
		Name:         strings.TrimSpace(input.Name),
		CategoryID:   input.CategoryID,
		RestaurantID: input.RestaurantID,
		MinPrice:     input.MinPrice,
		MaxPrice:     input.MaxPrice,
		Available:    input.Available,
		Limit:        limit,
		Offset:       offset,
	})
	if err != nil {
		return nil, err
	}
	return &MenuSearchPage{Items: items, Page: page, Limit: limit, Total: total}, nil
}

func menuPageBounds(page, limit int) (int, int, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = defaultMenuPageSize
	}
	if limit > maxMenuPageSize {
		return 0, 0, 0, fmt.Errorf("%w: limit must be between 1 and %d", ErrInvalidMenuItemInput, maxMenuPageSize)
	}
	return page, limit, (page - 1) * limit, nil
}
