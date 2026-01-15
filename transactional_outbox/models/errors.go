package models

import "errors"

var (
	ErrOrderNotFound      = errors.New("order not found")
	ErrInvalidCustomerID  = errors.New("invalid customer ID")
	ErrInvalidProductName = errors.New("invalid product name")
	ErrInvalidQuantity    = errors.New("invalid quantity")
	ErrInvalidPrice       = errors.New("invalid price")

	ErrOutboxMessageNotFound = errors.New("outbox message not found")
	ErrAlreadyProcessed      = errors.New("message already processed")
)
