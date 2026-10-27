package service

import (
	"Fhub/models"
	"Fhub/repository"
	"context"
	"errors"
	"regexp"
	"strings"
)

const (
	maxRecipientNameLength        = 100
	maxDeliveryAddressLength      = 200
	maxDeliveryCityLength         = 100
	maxDeliveryStateLength        = 100
	maxDeliveryInstructionsLength = 500
)

// Error returned when a restaurant is inactive.
var ErrRestaurantNotActive = errors.New("restaurant is not active")

// Error returned when an invalid order status is provided.
var ErrInvalidOrderStatus = errors.New("invalid order status")

// Error returned when a manager does not own the restaurant.
var ErrOrderAccessDenied = errors.New("you are not authorised to view this restaurant's orders")

// OrderService defines the business operations related to orders.
type OrderService interface {

	// CanAcceptOrder checks whether a restaurant can accept orders.
	CanAcceptOrder(ctx context.Context, restaurantID uint) error

	// ValidateDeliveryInfo checks the customer's delivery details.
	ValidateDeliveryInfo(input *DeliveryInfoInput) map[string]string

	// GetRestaurantOrders retrieves orders after checking the manager's access and requested status.
	GetRestaurantOrders(ctx context.Context, managerID uint, restaurantID uint, status string) ([]models.Order, error)
}

// orderService stores the repositories used to manage orders.
type orderService struct {
	orderRepo      repository.OrderRepository
	restaurantRepo repository.RestaurantRepository
}

// NewOrderService creates an order service with both repositories.
func NewOrderService(orderRepo repository.OrderRepository, restaurantRepo repository.RestaurantRepository) OrderService {
	return &orderService{orderRepo: orderRepo, restaurantRepo: restaurantRepo}
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

// GetRestaurantOrders validates the order status, checks restaurant ownership and retrieves the restaurant's orders.
func (s *orderService) GetRestaurantOrders(ctx context.Context, managerID uint, restaurantID uint, status string) ([]models.Order, error) {

	// Reject invalid order statuses before searching the database.
	if !isValidOrderStatus(status) {
		return nil, ErrInvalidOrderStatus
	}
	// Find the restaurant using its ID.
	restaurant, err := s.restaurantRepo.FindByID(ctx, restaurantID)
	if err != nil {
		return nil, err
	}

	// Reject managers who do not own the restaurant.
	if restaurant.ManagerID != managerID {
		return nil, ErrOrderAccessDenied
	}

	// Retrieve the restaurant's orders, optionally filtered by status.
	return s.orderRepo.GetRestaurantOrders(ctx, restaurantID, status)
}

// isValidOrderStatus checks whether the requested order status exists.
func isValidOrderStatus(status string) bool {

	// An empty status means the manager wants all orders.
	if status == "" {
		return true
	}

	// Check the status against the allowed order statuses.
	switch models.OrderStatus(status) {
	case models.OrderStatusPending,
		models.OrderStatusConfirmed,
		models.OrderStatusPreparing,
		models.OrderStatusReadyForDelivery,
		models.OrderStatusOutForDelivery,
		models.OrderStatusDelivered,
		models.OrderStatusCancelled:
		return true
	default:
		return false
	}
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
