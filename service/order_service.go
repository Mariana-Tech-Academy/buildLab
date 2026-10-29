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
	defaultOrderPageSize          = 10
	maxOrderPageSize              = 50
)

var (
	ErrRestaurantNotActive = errors.New("restaurant is not active")
	ErrEmptyCart           = errors.New("cannot checkout with an empty cart")
	ErrUnavailableItem     = errors.New("cart contains an unavailable item")
)

// Error returned when an invalid order status is provided.
var ErrInvalidOrderStatus = errors.New("invalid order status")

// Error returned when a manager does not own the restaurant.
var ErrOrderAccessDenied = errors.New("you are not authorised to view this restaurant's orders")

var ErrInvalidOrderPage = errors.New("limit must be between 1 and 50")

var ErrOrderNotFound = errors.New("order not found")

// Error returned when an order has gone too far to be cancelled
var ErrOrderCannotBeCancelled = errors.New("this order can no longer be cancelled")

// OrderService defines the business operations related to orders.
type OrderService interface {

	// CanAcceptOrder checks whether a restaurant can accept orders.
	CanAcceptOrder(ctx context.Context, restaurantID uint) error
	GetOrderByID(ctx context.Context, orderID uint, customerID uint) (*models.Order, error)

	GetCheckoutSummary(ctx context.Context, customerID uint) (*models.CheckoutSummary, error)

	// ValidateDeliveryInfo checks the customer's delivery details.
	ValidateDeliveryInfo(input *DeliveryInfoInput) map[string]string

	// GetRestaurantOrders retrieves one page of orders after checking access and the requested status.
	GetRestaurantOrders(ctx context.Context, managerID uint, role models.Role, restaurantID uint, status string, page, limit int) (*RestaurantOrdersPage, error)
	List(ctx context.Context, customerID uint, status string, page, limit int) (*CustomerOrdersPage, error)

	// CancelOrder cancels an order if the user is allowed to and the order has not gone too far.
	CancelOrder(ctx context.Context, userID uint, role models.Role, orderID uint) (*models.Order, error)
}

// CustomerOrdersPage is one page of the customer's own orders.
type CustomerOrdersPage struct {
	Orders []models.Order
	Page   int
	Limit  int
	Total  int64
}

// RestaurantOrdersPage is one page of orders for a restaurant.
type RestaurantOrdersPage struct {
	Orders []models.Order
	Page   int
	Limit  int
	Total  int64
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

// GetOrderByID retrieves an order using the order ID and customer ID.
// Returns the order details if found, or an error if no matching order exists.
func (s *orderService) GetOrderByID(ctx context.Context, orderID uint, customerID uint) (*models.Order, error) {
	order, err := s.orderRepo.GetOrderByID(ctx, orderID, customerID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	return order, nil
}

// GetRestaurantOrders validates the order status, checks restaurant ownership and retrieves one page of orders.
func (s *orderService) GetRestaurantOrders(ctx context.Context, managerID uint, role models.Role, restaurantID uint, status string, page, limit int) (*RestaurantOrdersPage, error) {
	if !isValidOrderStatus(status) {
		return nil, ErrInvalidOrderStatus
	}

	page, limit, offset, err := orderPageBounds(page, limit)
	if err != nil {
		return nil, err
	}

	restaurant, err := s.restaurantRepo.FindByID(ctx, restaurantID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRestaurantNotFound
	}
	if err != nil {
		return nil, err
	}

	if role != models.RoleAdmin && restaurant.ManagerID != managerID {
		return nil, ErrOrderAccessDenied
	}

	orders, total, err := s.orderRepo.GetRestaurantOrders(ctx, restaurantID, status, limit, offset)
	if err != nil {
		return nil, err
	}
	return &RestaurantOrdersPage{Orders: orders, Page: page, Limit: limit, Total: total}, nil
}

func orderPageBounds(page, limit int) (int, int, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = defaultOrderPageSize
	}
	if limit > maxOrderPageSize {
		return 0, 0, 0, ErrInvalidOrderPage
	}
	return page, limit, (page - 1) * limit, nil
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

func (s *orderService) List(ctx context.Context, customerID uint, status string, page, limit int) (*CustomerOrdersPage, error) {
	status = strings.TrimSpace(status)
	if !isValidOrderStatus(status) {
		return nil, ErrInvalidOrderStatus
	}

	page, limit, offset, err := orderPageBounds(page, limit)
	if err != nil {
		return nil, err
	}

	orders, total, err := s.orderRepo.ListForCustomer(ctx, customerID, status, limit, offset)
	if err != nil {
		return nil, err
	}
	return &CustomerOrdersPage{Orders: orders, Page: page, Limit: limit, Total: total}, nil
}

// cancellableStatuses lists the statuses an order can still be cancelled from.
// Orders can be cancelled while Pending or Confirmed, but not once the restaurant starts Preparing.
var cancellableStatuses = []models.OrderStatus{
	models.OrderStatusPending,
	models.OrderStatusConfirmed,
}

// CanCancel checks whether an order in this status can still be cancelled.
func CanCancel(status models.OrderStatus) bool {
	for _, allowed := range cancellableStatuses {
		if status == allowed {
			return true
		}
	}
	return false
}

// canUserCancel checks whether this user is allowed to cancel this order.
// Customers can cancel their own orders, managers can cancel orders for their
// own restaurant, and admins can cancel any order.
func canUserCancel(order *models.Order, userID uint, role models.Role) bool {
	switch role {
	case models.RoleAdmin:
		return true
	case models.RoleCustomer:
		return order.CustomerID == userID
	case models.RoleManager:
		return order.Restaurant.ID != 0 && order.Restaurant.ManagerID == userID
	default:
		return false
	}
}

// CancelOrder cancels an order after checking who is cancelling it and its current status.
func (s *orderService) CancelOrder(ctx context.Context, userID uint, role models.Role, orderID uint) (*models.Order, error) {
	order, err := s.orderRepo.FindOrderForCancel(ctx, orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}

	// Return not found, rather than forbidden, so callers cannot tell whether someone else's order exists.
	if !canUserCancel(order, userID, role) {
		return nil, ErrOrderNotFound
	}

	if !CanCancel(order.Status) {
		return nil, ErrOrderCannotBeCancelled
	}

	if err := s.orderRepo.CancelOrder(ctx, order.ID, cancellableStatuses); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// The order moved on (for example to preparing) just before we cancelled it.
			return nil, ErrOrderCannotBeCancelled
		}
		return nil, err
	}

	order.Status = models.OrderStatusCancelled
	return order, nil
}
