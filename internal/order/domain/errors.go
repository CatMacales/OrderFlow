package domain

import "errors"

var (
	ErrInvalidArgument   = errors.New("INVALID_ARGUMENT")
	ErrInvalidOrderState = errors.New("INVALID_ORDER_STATE")
)
