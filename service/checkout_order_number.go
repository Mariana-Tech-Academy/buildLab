package service

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

// GenerateCheckoutOrderNumber creates a temporary order number using the team's required date and number format.
func GenerateCheckoutOrderNumber() (string, error) {

	// Generate a random number between 001 and 999.
	number, err := rand.Int(rand.Reader, big.NewInt(999))
	if err != nil {
		return "", err
	}

	// Add one so the number never starts at 000.
	sequence := number.Int64() + 1

	// Format the order number using today's date.
	return fmt.Sprintf(
		"ORD-%s-%03d",
		time.Now().Format("20060102"),
		sequence,
	), nil
}
