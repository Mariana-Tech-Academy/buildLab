
package repository

import (
    "errors"

    "Fhub/models"

    "gorm.io/gorm"
    "gorm.io/gorm/clause"
)

var ErrCartItemNotFound = errors.New("cart item not found")

// CartRepository defines the database operations.
type CartRepository interface {
    RemoveItem(userID, itemID uint) (*models.Cart, error)
    Clear(userID uint) (*models.Cart, error)
}

// PostgresCartRepo connects to PostgreSQL.
type PostgresCartRepo struct {
    db *gorm.DB
}

// NewCartRepository creates a repository.
func NewCartRepository(db *gorm.DB) CartRepository {
    return &PostgresCartRepo{db: db}
}

// Load the cart with its items and menu item details.
func cartWithItems(tx *gorm.DB, cartID uint) (*models.Cart, error) {
    var cart models.Cart

    err := tx.Preload("Items.MenuItem").
        First(&cart, cartID).Error

    if err != nil {
        return nil, err
    }

    return &cart, nil
}

// Remove one item from the customer's cart.
func (r *PostgresCartRepo) RemoveItem(
    userID, itemID uint,
) (*models.Cart, error) {

    var result *models.Cart

    err := r.db.Transaction(func(tx *gorm.DB) error {
        var cart models.Cart

        err := tx.Clauses(
            clause.Locking{Strength: "UPDATE"},
        ).Where("user_id = ?", userID).
            First(&cart).Error

        if errors.Is(err, gorm.ErrRecordNotFound) {
            return ErrCartItemNotFound
        }

        if err != nil {
            return err
        }

        deleted := tx.Unscoped().
            Where("id = ? AND cart_id = ?", itemID, cart.ID).
            Delete(&models.CartItem{})

        if deleted.Error != nil {
            return deleted.Error
        }

        if deleted.RowsAffected == 0 {
            return ErrCartItemNotFound
        }

        var count int64

        err = tx.Model(&models.CartItem{}).
            Where("cart_id = ?", cart.ID).
            Count(&count).Error

        if err != nil {
            return err
        }

        // Reset restaurant when cart becomes empty.
        if count == 0 {
            err = tx.Model(&cart).
                Update("restaurant_id", nil).Error

            if err != nil {
                return err
            }
        }

        result, err = cartWithItems(tx, cart.ID)
        return err
    })

    return result, err
}

// Clear all items from the customer's cart.
func (r *PostgresCartRepo) Clear(
    userID uint,
) (*models.Cart, error) {

    var result *models.Cart

    err := r.db.Transaction(func(tx *gorm.DB) error {
        var cart models.Cart

        err := tx.Clauses(
            clause.Locking{Strength: "UPDATE"},
        ).Where("user_id = ?", userID).
            First(&cart).Error

        if errors.Is(err, gorm.ErrRecordNotFound) {
            result = &models.Cart{
                UserID: userID,
                Items:  []models.CartItem{},
            }
            return nil
        }

        if err != nil {
            return err
        }

        // Remove every cart item.
        err = tx.Unscoped().
            Where("cart_id = ?", cart.ID).
            Delete(&models.CartItem{}).Error

        if err != nil {
            return err
        }

        // Remove restaurant association.
        err = tx.Model(&cart).
            Update("restaurant_id", nil).Error

        if err != nil {
            return err
        }

        result, err = cartWithItems(tx, cart.ID)
        return err
    })

    return result, err
}
