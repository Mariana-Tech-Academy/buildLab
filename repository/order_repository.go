package repository


import (
    "errors"
    "fmt"
    "math"
    "time"

    "Fhub/models"

    "gorm.io/gorm"
    "gorm.io/gorm/clause"
)


var ErrEmptyCart = errors.New("cart is empty")
var ErrUnavailableItem = errors.New("a cart item is unavailable")

type DeliverySnapshot struct {
    Name         string `json:"delivery_name"`
    Phone        string `json:"delivery_phone"`
    Address      string `json:"delivery_address"`
    City         string `json:"delivery_city"`
    State        string `json:"delivery_state"`
    Instructions string `json:"delivery_instructions"`
}


// Checkout must create the order, create its items and clear the cart inside one
// r.db.Transaction(func(tx *gorm.DB) error { ... }) call.
type OrderRepository interface { Checkout(uint, DeliverySnapshot) (*models.Order, error)
List(uint) ([]models.Order, error)
Get(uint, uint) (*models.Order, error)
}

type PostgresOrderRepo struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &PostgresOrderRepo{db: db}
}


func (r *PostgresOrderRepo) Checkout(
    customerID uint,
    d DeliverySnapshot,
) (*models.Order, error) {

    var order models.Order

    err := r.db.Transaction(func(tx *gorm.DB) error {
        var cart models.Cart

        if err := tx.Clauses(
            clause.Locking{Strength: "UPDATE"},
        ).Where("user_id = ?", customerID).
            First(&cart).Error; err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                return ErrEmptyCart
            }
            return err
        }

        var cartItems []models.CartItem

        if err := tx.Where("cart_id = ?", cart.ID).
            Find(&cartItems).Error; err != nil {
            return err
        }

        if len(cartItems) == 0 || cart.RestaurantID == nil {
            return ErrEmptyCart
        }

        order = models.Order{
            CustomerID: customerID,
            RestaurantID: *cart.RestaurantID,
            Status: models.OrderStatusPending,

            DeliveryName: d.Name,
            DeliveryPhone: d.Phone,
            DeliveryAddress: d.Address,
            DeliveryCity: d.City,
            DeliveryState: d.State,
            DeliveryInstructions: d.Instructions,

            OrderNumber: fmt.Sprintf(
                "ORD-%s-%d",
                time.Now().UTC().Format("20060102"),
                time.Now().UTC().UnixNano(),
            ),
        }

        for _, ci := range cartItems {
            if ci.Quantity <= 0 {
                return errors.New("invalid quantity")
            }

            var menu models.MenuItem

            err := tx.Clauses(
                clause.Locking{Strength: "SHARE"},
            ).First(&menu, ci.MenuItemID).Error

            if errors.Is(err, gorm.ErrRecordNotFound) {
                return ErrUnavailableItem
            }
            if err != nil {
                return err
            }

            if !menu.Available ||
                menu.RestaurantID != *cart.RestaurantID {
                return ErrUnavailableItem
            }

            if menu.Price < 0 ||
                (menu.Price > 0 &&
                    int64(ci.Quantity) > math.MaxInt64/menu.Price) {
                return errors.New("invalid item price")
            }

            lineTotal := menu.Price * int64(ci.Quantity)

            if order.Subtotal > math.MaxInt64-lineTotal {
                return errors.New("order total overflow")
            }

            menuID := menu.ID

            order.Items = append(order.Items, models.OrderItem{
                MenuItemID: &menuID,
                Name: menu.Name,
                UnitPrice: menu.Price,
                Quantity: ci.Quantity,
                LineTotal: lineTotal,
            })

            order.Subtotal += lineTotal
        }

        order.DeliveryFee = 0
        order.Total = order.Subtotal + order.DeliveryFee

        if err := tx.Create(&order).Error; err != nil {
            return err
        }

        if err := tx.Where("cart_id = ?", cart.ID).
            Delete(&models.CartItem{}).Error; err != nil {
            return err
        }

        return tx.Model(&models.Cart{}).
            Where("id = ?", cart.ID).
            Update("restaurant_id", nil).Error
    })

    if err != nil {
        return nil, err
    }

    return &order, nil
}


func (r *PostgresOrderRepo) List(
    customerID uint,
) ([]models.Order, error) {
    var orders []models.Order

    err := r.db.Where("customer_id = ?", customerID).
        Preload("Items").
        Order("created_at DESC").
        Find(&orders).Error

    return orders, err
}

func (r *PostgresOrderRepo) Get(
    customerID uint,
    orderID uint,
) (*models.Order, error) {
    var order models.Order

    err := r.db.Where(
        "customer_id = ? AND id = ?",
        customerID,
        orderID,
    ).Preload("Items").First(&order).Error

    if err != nil {
        return nil, err
    }

    return &order, nil
}

