package storage

import "errors"

var (
	ErrOrderAlreadyExists = errors.New("ORDER_ALREADY_EXISTS")
	ErrOrderNotFound      = errors.New("ORDER_NOT_FOUND")
)
