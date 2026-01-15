package repository

import (
	"database/sql"
	"fmt"

	"transactional_outbox/models"
)

type OrderRepository interface {
	Create(tx *sql.Tx, order *models.CreateOrderRequest) (*models.Order, error)
	GetByID(id int64) (*models.Order, error)
	UpdateStatus(tx *sql.Tx, orderID int64, status string) error
	List(limit, offset int) ([]*models.Order, error)
}

type orderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(tx *sql.Tx, req *models.CreateOrderRequest) (*models.Order, error) {
	query := `
		INSERT INTO orders (customer_id, product_name, quantity, total_price, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, customer_id, product_name, quantity, total_price, status, created_at, updated_at
	`

	order := &models.Order{}
	err := tx.QueryRow(
		query,
		req.CustomerID,
		req.ProductName,
		req.Quantity,
		req.TotalPrice,
		models.OrderStatusPending,
	).Scan(
		&order.ID,
		&order.CustomerID,
		&order.ProductName,
		&order.Quantity,
		&order.TotalPrice,
		&order.Status,
		&order.CreatedAt,
		&order.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	return order, nil
}

func (r *orderRepository) GetByID(id int64) (*models.Order, error) {
	query := `
		SELECT id, customer_id, product_name, quantity, total_price, status, created_at, updated_at
		FROM orders
		WHERE id = $1
	`

	order := &models.Order{}
	err := r.db.QueryRow(query, id).Scan(
		&order.ID,
		&order.CustomerID,
		&order.ProductName,
		&order.Quantity,
		&order.TotalPrice,
		&order.Status,
		&order.CreatedAt,
		&order.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, models.ErrOrderNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return order, nil
}

func (r *orderRepository) UpdateStatus(tx *sql.Tx, orderID int64, status string) error {
	query := `
		UPDATE orders
		SET status = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`

	result, err := tx.Exec(query, status, orderID)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return models.ErrOrderNotFound
	}

	return nil
}

func (r *orderRepository) List(limit, offset int) ([]*models.Order, error) {
	query := `
		SELECT id, customer_id, product_name, quantity, total_price, status, created_at, updated_at
		FROM orders
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		order := &models.Order{}
		err := rows.Scan(
			&order.ID,
			&order.CustomerID,
			&order.ProductName,
			&order.Quantity,
			&order.TotalPrice,
			&order.Status,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return orders, nil
}
