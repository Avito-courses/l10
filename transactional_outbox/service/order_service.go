package service

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	"transactional_outbox/models"
	"transactional_outbox/repository"
)

type OrderService struct {
	db         *sql.DB
	orderRepo  repository.OrderRepository
	outboxRepo repository.OutboxRepository
}

func NewOrderService(
	db *sql.DB,
	orderRepo repository.OrderRepository,
	outboxRepo repository.OutboxRepository,
) *OrderService {
	return &OrderService{
		db:         db,
		orderRepo:  orderRepo,
		outboxRepo: outboxRepo,
	}
}

func (s *OrderService) CreateOrder(req *models.CreateOrderRequest) (*models.Order, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Printf("Failed to rollback transaction: %v", rbErr)
			}
		}
	}()

	order, err := s.orderRepo.Create(tx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	log.Printf("Order created in DB: ID=%d, Customer=%s, Product=%s",
		order.ID, order.CustomerID, order.ProductName)

	event := &models.OrderCreatedEvent{
		OrderID:     order.ID,
		CustomerID:  order.CustomerID,
		ProductName: order.ProductName,
		Quantity:    order.Quantity,
		TotalPrice:  order.TotalPrice,
		Status:      order.Status,
		CreatedAt:   order.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	outboxMsg, err := models.NewOutboxMessage(
		models.AggregateTypeOrder,
		strconv.FormatInt(order.ID, 10),
		models.EventTypeOrderCreated,
		event,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create outbox message: %w", err)
	}

	err = s.outboxRepo.Create(tx, outboxMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to save outbox message: %w", err)
	}

	log.Printf("Outbox message created: ID=%d, EventType=%s",
		outboxMsg.ID, outboxMsg.EventType)

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Transaction committed successfully!")

	return order, nil
}

func (s *OrderService) ConfirmOrder(orderID int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Printf("Failed to rollback transaction: %v", rbErr)
			}
		}
	}()

	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	err = s.orderRepo.UpdateStatus(tx, orderID, models.OrderStatusConfirmed)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	log.Printf("Order %d status updated to CONFIRMED", orderID)

	event := &models.OrderConfirmedEvent{
		OrderID:     orderID,
		CustomerID:  order.CustomerID,
		ConfirmedAt: order.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	outboxMsg, err := models.NewOutboxMessage(
		models.AggregateTypeOrder,
		strconv.FormatInt(orderID, 10),
		models.EventTypeOrderConfirmed,
		event,
	)
	if err != nil {
		return fmt.Errorf("failed to create outbox message: %w", err)
	}

	err = s.outboxRepo.Create(tx, outboxMsg)
	if err != nil {
		return fmt.Errorf("failed to save outbox message: %w", err)
	}

	log.Printf("Outbox message created for order confirmation")

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Transaction committed successfully!")

	return nil
}

func (s *OrderService) GetOrder(orderID int64) (*models.Order, error) {
	return s.orderRepo.GetByID(orderID)
}

func (s *OrderService) ListOrders(limit, offset int) ([]*models.Order, error) {
	return s.orderRepo.List(limit, offset)
}
