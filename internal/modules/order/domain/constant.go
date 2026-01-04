package domain

import (
	"errors"
)

var ErrInvalidTransition = errors.New("invalid transition")



// Service Types
const (
	TypeFood  = "FOOD"
	TypeRide  = "RIDE"
	TypeHotel = "HOTEL"
)

// Order Status
