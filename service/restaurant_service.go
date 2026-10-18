package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"regexp"
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

var restaurantPhoneRegex = regexp.MustCompile(`^(0|\+44)\d{10}$`)

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
	Create(ctx context.Context, input CreateRestaurantInput) (*models.Restaurant, error)
	UpdateRestaurant(ctx context.Context, restaurantID, actorID uint, role models.Role, input UpdateRestaurantInput) (*models.Restaurant, error)
	CheckOwnership(ctx context.Context, restaurantID, actorID uint, role models.Role) error
	UpdateRestaurantImage(ctx context.Context, restaurantID, actorID uint, role models.Role, imageURL string) (previousURL string, err error)
	SetRestaurantStatus(ctx context.Context, restaurantID, actorID uint, role models.Role, status models.RestaurantStatus) (*models.Restaurant, error)
	ListActive(ctx context.Context) ([]models.Restaurant, error)
}

type restaurantService struct {
	restaurantRepo repository.RestaurantRepository
}

func NewRestaurantService(repo repository.RestaurantRepository) RestaurantService {
	return &restaurantService{restaurantRepo: repo}
}
func (s *restaurantService) ListActive(ctx context.Context) ([]models.Restaurant, error) {
	return s.restaurantRepo.ListActive(ctx)
}

func (s *restaurantService) Create(ctx context.Context, input CreateRestaurantInput) (*models.Restaurant, error) {
	restaurant := &models.Restaurant{
		ManagerID:   input.ManagerID,
		Name:        input.Name,
		Description: input.Description,
		Address:     input.Address,
		Phone:       input.Phone,
		Email:       strings.ToLower(input.Email),
		OpeningTime: input.OpeningTime,
		ClosingTime: input.ClosingTime,
		Status:      models.RestaurantStatusPending,
	}

	if err := s.restaurantRepo.Create(ctx, restaurant); err != nil {
		return nil, err
	}
	return restaurant, nil
}

// CheckOwnership allows the restaurant's manager, or any admin.
func (s *restaurantService) CheckOwnership(ctx context.Context, restaurantID, actorID uint, role models.Role) error {
	restaurant, err := s.restaurantRepo.FindByID(ctx, restaurantID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrRestaurantNotFound
	}
	if err != nil {
		return err
	}
	if role == models.RoleAdmin {
		return nil
	}
	if restaurant.ManagerID != actorID {
		return ErrRestaurantForbidden
	}
	return nil
}

func (s *restaurantService) UpdateRestaurant(
	ctx context.Context,
	restaurantID, actorID uint,
	role models.Role,
	input UpdateRestaurantInput,
) (*models.Restaurant, error) {
	restaurant, err := s.restaurantRepo.FindByID(ctx, restaurantID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRestaurantNotFound
	}
	if err != nil {
		return nil, err
	}

	if role != models.RoleAdmin && restaurant.ManagerID != actorID {
		return nil, ErrRestaurantForbidden
	}

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return nil, fmt.Errorf("%w: name is required", ErrInvalidRestaurant)
		}
		restaurant.Name = name
	}

	if input.Description != nil {
		restaurant.Description = strings.TrimSpace(*input.Description)
	}

	if input.Address != nil {
		address := strings.TrimSpace(*input.Address)
		if address == "" {
			return nil, fmt.Errorf("%w: address is required", ErrInvalidRestaurant)
		}
		restaurant.Address = address
	}

	if input.City != nil {
		restaurant.City = strings.TrimSpace(*input.City)
	}

	if input.State != nil {
		restaurant.State = strings.TrimSpace(*input.State)
	}

	if input.Phone != nil {
		phone := strings.NewReplacer(" ", "", "-", "").Replace(*input.Phone)
		if phone == "" {
			return nil, fmt.Errorf("%w: phone is required", ErrInvalidRestaurant)
		}
		if !restaurantPhoneRegex.MatchString(phone) {
			return nil, fmt.Errorf("%w: phone must be a valid UK number, e.g. 01131234567 or +441131234567", ErrInvalidRestaurant)
		}
		restaurant.Phone = phone
	}

	if input.Email != nil {
		email := strings.TrimSpace(*input.Email)
		if email == "" {
			return nil, fmt.Errorf("%w: email is required", ErrInvalidRestaurant)
		}
		addr, err := mail.ParseAddress(email)
		if err != nil || addr.Address != email {
			return nil, fmt.Errorf("%w: email must be a valid email address", ErrInvalidRestaurant)
		}
		restaurant.Email = strings.ToLower(email)
	}

	if input.OpeningTime != nil {
		restaurant.OpeningTime = strings.TrimSpace(*input.OpeningTime)
	}
	if input.ClosingTime != nil {
		restaurant.ClosingTime = strings.TrimSpace(*input.ClosingTime)
	}

	if input.ImageURL != nil {
		restaurant.ImageURL = *input.ImageURL
	}

	// Only validate hours when the client sent at least one of them.
	if input.OpeningTime != nil || input.ClosingTime != nil {
		if err := validateOpeningHours(restaurant.OpeningTime, restaurant.ClosingTime); err != nil {
			return nil, err
		}
	}

	if err := s.restaurantRepo.Update(ctx, restaurant); err != nil {
		return nil, err
	}
	return restaurant, nil
}
func (s *restaurantService) SetRestaurantStatus(ctx context.Context, restaurantID, actorID uint, role models.Role, status models.RestaurantStatus) (*models.Restaurant, error) {
	restaurant, err := s.restaurantRepo.FindByID(ctx, restaurantID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRestaurantNotFound
	}
	if err != nil {
		return nil, err
	}

	if role != models.RoleAdmin && restaurant.ManagerID != actorID {
		return nil, ErrRestaurantForbidden
	}

	if role != models.RoleAdmin {
		if restaurant.Status != models.RestaurantStatusActive &&
			restaurant.Status != models.RestaurantStatusInactive {
			return nil, fmt.Errorf("%w: only an admin can change this status", ErrInvalidRestaurant)
		}
		if status != models.RestaurantStatusActive &&
			status != models.RestaurantStatusInactive {
			return nil, fmt.Errorf("%w: managers can only activate or deactivate", ErrInvalidRestaurant)
		}
	}

	restaurant.Status = status

	if err := s.restaurantRepo.Update(ctx, restaurant); err != nil {
		return nil, err
	}
	return restaurant, nil
}

func (s *restaurantService) UpdateRestaurantImage(
	ctx context.Context,
	restaurantID, actorID uint,
	role models.Role,
	imageURL string,
) (string, error) {
	restaurant, err := s.restaurantRepo.FindByID(ctx, restaurantID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", ErrRestaurantNotFound
	}
	if err != nil {
		return "", err
	}

	if role != models.RoleAdmin && restaurant.ManagerID != actorID {
		return "", ErrRestaurantForbidden
	}

	previousURL := restaurant.ImageURL
	if err := s.restaurantRepo.UpdateImageURL(ctx, restaurantID, imageURL); err != nil {
		return "", err
	}
	return previousURL, nil
}

func validateOpeningHours(openingTime, closingTime string) error {
	opening, err := time.Parse("15:04", openingTime)
	if err != nil {
		return fmt.Errorf("%w: opening time must be HH:MM", ErrInvalidRestaurant)
	}
	closing, err := time.Parse("15:04", closingTime)
	if err != nil {
		return fmt.Errorf("%w: closing time must be HH:MM", ErrInvalidRestaurant)
	}
	// Same-day hours only; overnight (e.g. 22:00–02:00) is not supported yet.
	if !closing.After(opening) {
		return fmt.Errorf("%w: closing time must be after opening time", ErrInvalidRestaurant)
	}
	return nil
}
