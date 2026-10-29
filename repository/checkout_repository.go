package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"

	"Fhub/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ErrRestaurantInactive means checkout was asked to charge a restaurant that is not active.
var ErrRestaurantInactive = errors.New("restaurant is not active")

// CheckoutRepository defines the database operations needed for checkout.
type CheckoutRepository interface {
	GetCheckoutCart(ctx context.Context, customerID uint) (*CheckoutCart, error)
	CreateCheckoutOrder(ctx context.Context, input CheckoutInput) (*models.Order, error)
}

// CheckoutCart contains the cart and its items.
type CheckoutCart struct {
	Cart  models.Cart
	Items []models.CartItem
}

// NewCheckoutRepository creates the repository used by checkout.
func NewCheckoutRepository(db *gorm.DB) CheckoutRepository {
	return &PostgresOrderRepo{db: db}
}

// GetCheckoutCart retrieves the customer's cart and its items.
func (r *PostgresOrderRepo) GetCheckoutCart(ctx context.Context, customerID uint) (*CheckoutCart, error) {
	var cart models.Cart
	err := r.db.WithContext(ctx).
		Preload("Items").
		First(&cart, "user_id = ?", customerID).Error
	if err != nil {
		return nil, err
	}
	return &CheckoutCart{
		Cart:  cart,
		Items: cart.Items,
	}, nil
}

// CheckoutInput contains the information needed to create an order.
type CheckoutInput struct {
	CustomerID uint

	DeliveryName         string
	DeliveryPhone        string
	DeliveryAddress      string
	DeliveryCity         string
	DeliveryState        string
	DeliveryInstructions string

	DeliveryFee int64
	Subtotal    func(items []models.CartItem, prices map[uint]int64) (int64, error)
}

type checkoutLine struct {
	name  string
	price int64
}

// CreateCheckoutOrder creates the order, stores item snapshots, and clears the cart.
// If an error occurs, GORM rolls back the transaction.
func (r *PostgresOrderRepo) CreateCheckoutOrder(ctx context.Context, input CheckoutInput) (*models.Order, error) {
	if input.Subtotal == nil {
		return nil, errors.New("checkout subtotal is not configured")
	}

	var order models.Order
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var cart models.Cart
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ?", input.CustomerID).
			First(&cart).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrEmptyCart
		}
		if err != nil {
			return err
		}

		var items []models.CartItem
		if err := tx.Where("cart_id = ?", cart.ID).Find(&items).Error; err != nil {
			return err
		}
		if len(items) == 0 || cart.RestaurantID == nil {
			return ErrEmptyCart
		}

		var restaurant models.Restaurant
		err = tx.Clauses(clause.Locking{Strength: "SHARE"}).
			First(&restaurant, *cart.RestaurantID).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return gorm.ErrRecordNotFound
		}
		if err != nil {
			return err
		}
		if restaurant.Status != models.RestaurantStatusActive {
			return ErrRestaurantInactive
		}

		lines := make(map[uint]checkoutLine, len(items))
		prices := make(map[uint]int64, len(items))
		for _, item := range items {
			var menuItem models.MenuItem
			err = tx.Clauses(clause.Locking{Strength: "SHARE"}).
				First(&menuItem, item.MenuItemID).Error
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			if errors.Is(err, gorm.ErrRecordNotFound) || !menuItem.Available || menuItem.RestaurantID != *cart.RestaurantID || item.Quantity <= 0 {
				return ErrUnavailableItem
			}
			if menuItem.Price < 0 {
				return errors.New("invalid menu item price")
			}
			lines[item.MenuItemID] = checkoutLine{name: menuItem.Name, price: menuItem.Price}
			prices[item.MenuItemID] = menuItem.Price
		}

		subtotal, err := input.Subtotal(items, prices)
		if err != nil {
			return err
		}
		if input.DeliveryFee < 0 {
			return errors.New("invalid delivery fee")
		}
		if subtotal > math.MaxInt64-input.DeliveryFee {
			return errors.New("order total exceeds the allowed amount")
		}

		order = models.Order{
			CustomerID:           input.CustomerID,
			RestaurantID:         *cart.RestaurantID,
			Restaurant:           restaurant,
			Status:               models.OrderStatusPending,
			Subtotal:             subtotal,
			DeliveryFee:          input.DeliveryFee,
			Total:                subtotal + input.DeliveryFee,
			DeliveryName:         input.DeliveryName,
			DeliveryPhone:        input.DeliveryPhone,
			DeliveryAddress:      input.DeliveryAddress,
			DeliveryCity:         input.DeliveryCity,
			DeliveryState:        input.DeliveryState,
			DeliveryInstructions: input.DeliveryInstructions,
		}
		if err := insertOrder(tx, &order); err != nil {
			return err
		}
		order.Restaurant = restaurant

		for _, item := range items {
			line := lines[item.MenuItemID]
			orderItem := models.OrderItem{
				OrderID:    order.ID,
				MenuItemID: item.MenuItemID,
				Name:       line.name,
				UnitPrice:  line.price,
				Quantity:   item.Quantity,
				LineTotal:  line.price * int64(item.Quantity),
			}
			if err := tx.Create(&orderItem).Error; err != nil {
				return err
			}
			order.Items = append(order.Items, orderItem)
		}

		if err := simulateCheckoutPayment(); err != nil {
			return err
		}
		if err := tx.Unscoped().Where("cart_id = ?", cart.ID).Delete(&models.CartItem{}).Error; err != nil {
			return err
		}
		return clearCartRestaurant(tx, cart.ID)
	})
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// simulateCheckoutPayment accepts the payment. A returned error rolls the order back.
func simulateCheckoutPayment() error {
	return nil
}

func insertOrder(tx *gorm.DB, order *models.Order) error {
	const attempts = 5
	for attempt := 0; attempt < attempts; attempt++ {
		number, err := nextOrderNumber(tx)
		if err != nil {
			return err
		}
		order.ID = 0
		order.OrderNumber = number
		if err := tx.SavePoint("create_order").Error; err != nil {
			return err
		}
		if err := tx.Create(order).Error; err != nil {
			if !isUniqueViolation(err) {
				return err
			}
			if err := tx.RollbackTo("create_order").Error; err != nil {
				return err
			}
			continue
		}
		return nil
	}
	return errors.New("could not allocate an order number")
}

func nextOrderNumber(tx *gorm.DB) (string, error) {
	prefix := "ORD-" + time.Now().Format("20060102") + "-"
	var sequence sql.NullInt64
	err := tx.Raw(
		`SELECT MAX(CAST(SPLIT_PART(order_number, '-', 3) AS INTEGER))
		 FROM orders
		 WHERE order_number LIKE ?`,
		prefix+"%",
	).Row().Scan(&sequence)
	if err != nil {
		return "", err
	}
	next := 1
	if sequence.Valid {
		next = int(sequence.Int64) + 1
	}
	return fmt.Sprintf("%s%03d", prefix, next), nil
}
