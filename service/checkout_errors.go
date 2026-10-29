package service

import "errors"

// ErrInvalidDelivery is returned when checkout delivery details fail validation.
var ErrInvalidDelivery = errors.New("invalid delivery information")
