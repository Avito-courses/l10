package models

import (
	"encoding/json"
	"time"
)

type OutboxMessage struct {
	ID            int64           `json:"id"`
	AggregateType string          `json:"aggregate_type"`
	AggregateID   string          `json:"aggregate_id"`
	EventType     string          `json:"event_type"`
	Payload       json.RawMessage `json:"payload"`
	CreatedAt     time.Time       `json:"created_at"`
	ProcessedAt   *time.Time      `json:"processed_at,omitempty"`
	Processed     bool            `json:"processed"`
}

const (
	EventTypeOrderCreated   = "OrderCreated"
	EventTypeOrderConfirmed = "OrderConfirmed"
)

const (
	AggregateTypeOrder = "Order"
)

type OrderCreatedEvent struct {
	OrderID     int64   `json:"order_id"`
	CustomerID  string  `json:"customer_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	TotalPrice  float64 `json:"total_price"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"created_at"`
}

type OrderConfirmedEvent struct {
	OrderID     int64  `json:"order_id"`
	CustomerID  string `json:"customer_id"`
	ConfirmedAt string `json:"confirmed_at"`
}

func NewOutboxMessage(aggregateType, aggregateID, eventType string, payload interface{}) (*OutboxMessage, error) {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &OutboxMessage{
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		EventType:     eventType,
		Payload:       payloadJSON,
		Processed:     false,
	}, nil
}
