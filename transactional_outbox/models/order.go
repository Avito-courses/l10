package models

import (
	"time"
)

type Order struct {
	ID          int64     `json:"id"`
	CustomerID  string    `json:"customer_id"`
	ProductName string    `json:"product_name"`
	Quantity    int       `json:"quantity"`
	TotalPrice  float64   `json:"total_price"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

const (
	OrderStatusPending   = "PENDING"
	OrderStatusConfirmed = "CONFIRMED"
)

type CreateOrderRequest struct {
	CustomerID  string  `json:"customer_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	TotalPrice  float64 `json:"total_price"`
}

func (r *CreateOrderRequest) Validate() error {
	if r.CustomerID == "" {
		return ErrInvalidCustomerID
	}
	if r.ProductName == "" {
		return ErrInvalidProductName
	}
	if r.Quantity <= 0 {
		return ErrInvalidQuantity
	}
	if r.TotalPrice <= 0 {
		return ErrInvalidPrice
	}
	return nil
}
