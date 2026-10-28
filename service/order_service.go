package service

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"Fhub/models"
	"Fhub/repository"

	"gorm.io/gorm"
)

const (
	maxRecipientNameLength        = 100
	maxDeliveryAddressLength      = 200
	maxDeliveryCityLength         = 100
	maxDeliveryStateLength        = 100
	maxDeliveryInstructionsLength = 500
)

var (
	ErrRestaurantNotActive = errors.New("restaurant is not active")
	ErrEmptyCart           = errors.New("cannot checkout with an empty cart")
	ErrUnavailableItem     = errors.New("cart contains an unavailable item")
)

type OrderService interface {
	CanAcceptOrder(ctx context.Context, restaurantID uint) error
	GetCheckoutSummary(ctx context.Context, customerID uint) (*models.CheckoutSummary, error)
	ValidateDeliveryInfo(input *DeliveryInfoInput) map[string]string
}

type orderService struct {
	orderRepo repository.OrderRepository
}

func NewOrderService(repo repository.OrderRepository) OrderService {
	return &orderService{orderRepo: repo}
}

// CanAcceptOrder is used by checkout (when implemented) to reject orders
// for restaurants that are not active.
func (s *orderService) CanAcceptOrder(ctx context.Context, restaurantID uint) error {
	active, err := s.orderRepo.IsRestaurantActive(ctx, restaurantID)
	if err != nil {
		return err
	}
	if !active {
		return ErrRestaurantNotActive
	}
	return nil
}

func (s *orderService) GetCheckoutSummary(ctx context.Context, customerID uint) (*models.CheckoutSummary, error) {
	summary, err := s.orderRepo.GetCheckoutSummary(ctx, customerID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrEmptyCart):
			return nil, ErrEmptyCart
		case errors.Is(err, repository.ErrUnavailableItem):
			return nil, ErrUnavailableItem
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrRestaurantNotFound
		default:
			return nil, err
		}
	}

	if err := s.CanAcceptOrder(ctx, summary.RestaurantID); err != nil {
		if errors.Is(err, ErrRestaurantNotActive) || errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRestaurantNotFound
		}
		return nil, err
	}

	return summary, nil
}

// DeliveryInfoInput represents the delivery details
// provided by a customer during checkout.
type DeliveryInfoInput struct {
	RecipientName        string `json:"recipient_name"`
	Phone                string `json:"phone"`
	Address              string `json:"address"`
	City                 string `json:"city"`
	State                string `json:"state"`
	DeliveryInstructions string `json:"delivery_instructions"`
}

// deliveryPhoneRegex validates UK phone numbers.
var deliveryPhoneRegex = regexp.MustCompile(`^(0|\+44)\d{10}$`)

// ValidateDeliveryInfo validates customer delivery details.
func (s *orderService) ValidateDeliveryInfo(input *DeliveryInfoInput) map[string]string {
	errs := make(map[string]string)

	// Remove unnecessary spaces.
	input.RecipientName = strings.TrimSpace(input.RecipientName)
	input.Address = strings.TrimSpace(input.Address)
	input.City = strings.TrimSpace(input.City)
	input.State = strings.TrimSpace(input.State)
	input.DeliveryInstructions = strings.TrimSpace(input.DeliveryInstructions)

	// Remove spaces and hyphens from the phone number.
	input.Phone = strings.ReplaceAll(input.Phone, " ", "")
	input.Phone = strings.ReplaceAll(input.Phone, "-", "")

	if input.RecipientName == "" {
		errs["recipient_name"] = "recipient name is required"
	} else if len(input.RecipientName) > maxRecipientNameLength {
		errs["recipient_name"] = "recipient name cannot exceed 100 characters"
	}

	if input.Phone == "" {
		errs["phone"] = "phone number is required"
	} else if !deliveryPhoneRegex.MatchString(input.Phone) {
		errs["phone"] = "phone number must be a valid UK number, e.g. 01131234567 or +441131234567"
	}

	if input.Address == "" {
		errs["address"] = "delivery address is required"
	} else if len(input.Address) > maxDeliveryAddressLength {
		errs["address"] = "delivery address cannot exceed 200 characters"
	}

	if input.City == "" {
		errs["city"] = "city is required"
	} else if len(input.City) > maxDeliveryCityLength {
		errs["city"] = "city cannot exceed 100 characters"
	}

	if input.State == "" {
		errs["state"] = "state is required"
	} else if len(input.State) > maxDeliveryStateLength {
		errs["state"] = "state cannot exceed 100 characters"
	}

	if len(input.DeliveryInstructions) > maxDeliveryInstructionsLength {
		errs["delivery_instructions"] = "delivery instructions cannot exceed 500 characters"
	}

	return errs
}
