package repository

import (
	"context"
	"errors"
	"math"
	"Fhub/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)
// CheckoutRepository defines the database operations needed for checkout.
type CheckoutRepository interface {

	// GetCheckoutCart retrieves a customer's cart and it's items.
	GetCheckoutCart(
		ctx context.Context,
		customerID uint,
	) (*CheckoutCart, error)

	// CreateCheckoutOrder processes an order inside one database transaction.
	CreateCheckoutOrder(
		ctx context.Context,
		input CheckoutInput,
	) (*models.Order, error)
}
// CheckoutCart contains the cart and it's items.
type CheckoutCart struct {
	Cart  models.Cart
	Items []models.CartItem
}

// GetCheckoutCart retrieves the customer's cart and it's items.
func (r *PostgresOrderRepo) GetCheckoutCart(
	ctx context.Context,
	customerID uint,
) (*CheckoutCart, error) {

	var cart models.Cart

	// Find the cart belonging to the customer.
	err := r.db.WithContext(ctx).
		Preload("Items").
		First(&cart, "user_id = ?", customerID).Error

	if err != nil {
		return nil, err
	}

	// Return the cart and it's items.
	return &CheckoutCart{
		Cart:  cart,
		Items: cart.Items,
	}, nil
}

// CheckoutInput contains the information needed to create an order.
type CheckoutInput struct {
	CustomerID uint

	// Delivery details provided by the customer.
	DeliveryName         string
	DeliveryPhone        string
	DeliveryAddress      string
	DeliveryCity         string
	DeliveryState        string
	DeliveryInstructions string

	OrderNumber string
	DeliveryFee int64
}

// CreateCheckoutOrder processes checkout inside one database transaction.
// If an error occurs, GORM rolls back the transaction.
func (r *PostgresOrderRepo) CreateCheckoutOrder(
	ctx context.Context,
	input CheckoutInput,
) (*models.Order, error) {

	var order models.Order

	// Start a database transaction.
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var cart models.Cart

		// Lock the customer's cart while checkout is processing.
		err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ?", input.CustomerID).
			First(&cart).Error

		if err != nil {

			// Reject checkout if the customer has no cart.
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("customer cart not found")
			}
			return err
		}
		// Retrieve the cart items using the same transaction.
		var items []models.CartItem

		if err := tx.
			Where("cart_id = ?", cart.ID).
			Find(&items).Error; err != nil {
			return err
		}

		// Reject checkout if the cart contains no items.
		if len(items) == 0 {
			return errors.New("cannot checkout an empty cart")
		}

		// Check that the cart has a restaurant.
		if cart.RestaurantID == nil {
			return errors.New("cart has no restaurant")
		}

		// Get the restaurant linked to the customer's cart.
		var restaurant models.Restaurant

		err = tx.
			Clauses(clause.Locking{Strength: "SHARE"}).
			First(&restaurant, *cart.RestaurantID).Error

		if err != nil {

			// Return an error if the restaurant cannot be found.
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("restaurant not found")
			}

			return err
		}

		// Check that the restaurant is active.
		if restaurant.Status != models.RestaurantStatusActive {
			return errors.New("restaurant is not active")
		}
		var subtotal int64
		prices := make(map[uint]int64)

		// Check each cart item against the current menu information.
		for _, item := range items {
			var menuItem models.MenuItem

			// Get the menu item and lock it during checkout.
			err = tx.
				Clauses(clause.Locking{Strength: "SHARE"}).
				First(&menuItem, item.MenuItemID).Error

			if err != nil {
				// Return an error if the menu item cannot be found.
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errors.New("menu item not found")
				}
				return err
			}

			// Check that the menu item is available.
			if !menuItem.Available {
				return errors.New("menu item is unavailable")
			}

			// Check that the item belongs to the selected restaurant.
			if menuItem.RestaurantID != *cart.RestaurantID {
				return errors.New("menu item does not belong to this restaurant")
			}

			// Check that the quantity is valid.
			if item.Quantity <= 0 {
				return errors.New("invalid item quantity")
			}

			// Check that the menu price is valid.
			if menuItem.Price < 0 {
				return errors.New("invalid menu item price")
			}

			// Store the current price for this menu item.
			prices[item.MenuItemID] = menuItem.Price

			// Calculate the cost of this item.
			quantity := int64(item.Quantity)

			// Check that multiplying price by quantity will not overflow.
			if menuItem.Price > 0 &&
				quantity > math.MaxInt64/menuItem.Price {
				return errors.New("item total exceeds the allowed amount")
			}

			itemTotal := menuItem.Price * quantity

			// Check that adding the item will not overflow the subtotal.
			if subtotal > math.MaxInt64-itemTotal {
				return errors.New("order subtotal exceeds the allowed amount")
			}

			// Add the item cost to the subtotal.
			subtotal += itemTotal
		}

		// Check that the delivery fee is not negative.
		if input.DeliveryFee < 0 {
			return errors.New("invalid delivery fee")
		}

		// Check that adding the delivery fee will not overflow.
		if subtotal > math.MaxInt64-input.DeliveryFee {
			return errors.New("order total exceeds the allowed amount")
		}

		// Calculate the final total including delivery.
		total := subtotal + input.DeliveryFee

		// Check that the order number has been provided.
		if input.OrderNumber == "" {
			return errors.New("order number is required")
		}

		// Create the order using the customer's delivery details.
		order = models.Order{
			OrderNumber:          input.OrderNumber,
			CustomerID:           input.CustomerID,
			RestaurantID:         *cart.RestaurantID,
			Status:               models.OrderStatusPending,
			Subtotal:             subtotal,
			DeliveryFee:          input.DeliveryFee,
			Total:                total,
			DeliveryName:         input.DeliveryName,
			DeliveryPhone:        input.DeliveryPhone,
			DeliveryAddress:      input.DeliveryAddress,
			DeliveryCity:         input.DeliveryCity,
			DeliveryState:        input.DeliveryState,
			DeliveryInstructions: input.DeliveryInstructions,
		}
		// Save the order to the database.
		if err := tx.Create(&order).Error; err != nil {
			return err
		}
		// Save each purchased item against the new order.
		for _, item := range items {
			// Retrieve the menu item details for the order snapshot.
			var menuItem models.MenuItem
			if err := tx.First(&menuItem, item.MenuItemID).Error; err != nil {
				return err
			}
			// Store the purchased item's name, price and quantity.
			orderItem := models.OrderItem{
				OrderID:    order.ID,
				MenuItemID: item.MenuItemID,
				Name:       menuItem.Name,
				UnitPrice:  prices[item.MenuItemID],
				Quantity:   item.Quantity,
				LineTotal:  prices[item.MenuItemID] * int64(item.Quantity),
			}
			// Save the order item in the same transaction.
			if err := tx.Create(&orderItem).Error; err != nil {
				return err
			}
			// Add the purchased item to the order response.
			order.Items = append(order.Items, orderItem)
		}

		// Simulate a successful payment without using a payment provider.
		paymentSuccessful := true

		// Stop checkout if the simulated payment fails.
		if !paymentSuccessful {
			return errors.New("simulated payment failed")
		}
		// Remove the purchased items from the customer's cart.

		if err := tx.
			Where("cart_id = ?", cart.ID).
			Delete(&models.CartItem{}).Error; err != nil {
			return err
		}
		// Reset the cart so the customer can order from another restaurant.
		if err := tx.
			Model(&cart).
			Update("restaurant_id", nil).Error; err != nil {

			return err
		}
		// Commit the transaction when all checkout operations succeed.
		return nil
	})

	// Return any error from the transaction.
	if err != nil {
		return nil, err
	}

	return &order, nil
}
