package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"Fhub/models"
	"Fhub/repository"
)

var ErrInvalidCategory = errors.New("invalid category details")

type CreateCategoryInput struct {
	Name string `json:"name"`
}

type CategoryService interface {
	Create(
		ctx context.Context,
		restaurantID, actorID uint,
		role models.Role,
		input CreateCategoryInput,
	) (*models.Category, error)

	GetByRestaurantID(
		ctx context.Context,
		restaurantID uint,
	) ([]models.Category, error)
}

type categoryService struct {
	categoryRepo      repository.CategoryRepository
	restaurantService RestaurantService
}

func NewCategoryService(
	repo repository.CategoryRepository,
	restaurantService RestaurantService,
) CategoryService {
	return &categoryService{
		categoryRepo:      repo,
		restaurantService: restaurantService,
	}
}

// Create adds a new category to a restaurant.

func (s *categoryService) Create(
	ctx context.Context,
	restaurantID, actorID uint,
	role models.Role,
	input CreateCategoryInput,
) (*models.Category, error) {

	// Check restaurant ownership
	if err := s.restaurantService.CheckOwnership(
		ctx,
		restaurantID,
		actorID,
		role,
	); err != nil {
		return nil, err
	}

	// Validate category name
	name := strings.TrimSpace(input.Name)

	if name == "" {
		return nil, fmt.Errorf(
			"%w: category name is required",
			ErrInvalidCategory,
		)
	}

	taken, err := s.categoryRepo.NameExists(ctx, restaurantID, name)
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, fmt.Errorf("%w: category name already exists", ErrInvalidCategory)
	}

	// Create category
	category := &models.Category{
		RestaurantID: restaurantID,
		Name:         name,
	}

	if err := s.categoryRepo.Create(ctx, category); err != nil {
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("%w: category name already exists", ErrInvalidCategory)
		}
		return nil, err
	}

	return category, nil
}

// GetByRestaurantID retrieves categories for an active restaurant.
func (s *categoryService) GetByRestaurantID(
	ctx context.Context,
	restaurantID uint,
) ([]models.Category, error) {
	if _, err := s.restaurantService.GetActiveRestaurant(ctx, restaurantID); err != nil {
		return nil, err
	}
	return s.categoryRepo.FindByRestaurantID(ctx, restaurantID)
}
