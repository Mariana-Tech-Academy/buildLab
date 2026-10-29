package service

import (
	"context"
	"errors"
	"regexp"
	"strings"
    "Fhub/models"
	"Fhub/repository"
	"fmt"
    "gorm.io/gorm"
)

const (
	maxRecipientNameLength        = 100
	maxDeliveryAddressLength      = 200
	maxDeliveryCityLength         = 100
	maxDeliveryStateLength        = 100
	maxDeliveryInstructionsLength = 500
)

var ErrRestaurantNotActive = errors.New("restaurant is not active")
var (
    ErrOrderNotFound = errors.New("order not found")
    ErrOrderForbidden = errors.New("you cannot update an order belonging to another restaurant")
    ErrInvalidOrderStatus = errors.New("invalid order status")
    ErrInvalidStatusTransition = errors.New("invalid order status transition")
    ErrOrderStatusFinal = errors.New("delivered or cancelled orders cannot be updated")
)
type OrderService interface {
	CanAcceptOrder(ctx context.Context, restaurantID uint) error
	ValidateDeliveryInfo(input *DeliveryInfoInput) map[string]string
	UpdateOrderStatus(
    ctx context.Context,
    orderID uint,
    restaurantID uint,
    actorID uint,
    role models.Role,
    newStatus models.OrderStatus,
) error
}
// UpdateOrderStatus validates and updates an order's status.
func (s *orderService) UpdateOrderStatus(
    ctx context.Context,
    orderID uint,
    restaurantID uint,
    actorID uint,
    role models.Role,
    newStatus models.OrderStatus,
) error {

    // Retrieve the order from the database.
    order, err := s.orderRepo.GetOrderByID(ctx, orderID)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return ErrOrderNotFound
        }
        return err
    }
    // Verify that the authenticated manager owns the restaurant.
if err := s.restaurantService.CheckOwnership(
    ctx,
    restaurantID,
    actorID,
    role,
); err != nil {
    return err
}
    // Ensure the order belongs to the manager's restaurant.
    if order.RestaurantID != restaurantID {
        return ErrOrderForbidden
    }

    // Delivered and cancelled orders cannot change status.
    if order.Status == models.OrderStatusDelivered ||
        order.Status == models.OrderStatusCancelled {
        return ErrOrderStatusFinal
    }

    // Define the only allowed next status for each stage.
    allowedTransitions := map[models.OrderStatus]models.OrderStatus{
        models.OrderStatusPending:          models.OrderStatusConfirmed,
        models.OrderStatusConfirmed:        models.OrderStatusPreparing,
        models.OrderStatusPreparing:        models.OrderStatusReadyForDelivery,
        models.OrderStatusReadyForDelivery: models.OrderStatusOutForDelivery,
        models.OrderStatusOutForDelivery:   models.OrderStatusDelivered,
    }

    expectedStatus, ok := allowedTransitions[order.Status]
    if !ok {
        return ErrInvalidOrderStatus
    }

    // Prevent skipping stages or moving backwards.
    if newStatus != expectedStatus {
        return fmt.Errorf(
            "%w: allowed next status is %s",
            ErrInvalidStatusTransition,
            expectedStatus,
        )
    }

    // Save the new status and updated timestamp.
    err = s.orderRepo.UpdateOrderStatus(
    ctx,
    orderID,
    restaurantID,
    order.Status,
    newStatus,
)

if errors.Is(err, gorm.ErrRecordNotFound) {
    return fmt.Errorf(
        "%w: order status has changed, please refresh and try again",
        ErrInvalidStatusTransition,
    )
}

return err
}

type orderService struct {
    orderRepo          repository.OrderRepository
    restaurantService  RestaurantService
}

func NewOrderService(
    repo repository.OrderRepository,
    restaurantService RestaurantService,
) OrderService {
    return &orderService{
        orderRepo:         repo,
        restaurantService: restaurantService,
    }
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
