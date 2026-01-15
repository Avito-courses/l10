package repository

import (
	"database/sql"
	"fmt"
	"time"

	"transactional_outbox/models"
)

type OutboxRepository interface {
	Create(tx *sql.Tx, message *models.OutboxMessage) error
	GetUnprocessed(limit int) ([]*models.OutboxMessage, error)
	MarkAsProcessed(tx *sql.Tx, id int64) error
	GetByID(id int64) (*models.OutboxMessage, error)
}

type outboxRepository struct {
	db *sql.DB
}

func NewOutboxRepository(db *sql.DB) OutboxRepository {
	return &outboxRepository{db: db}
}

func (r *outboxRepository) Create(tx *sql.Tx, message *models.OutboxMessage) error {
	query := `
		INSERT INTO outbox (aggregate_type, aggregate_id, event_type, payload, processed)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	err := tx.QueryRow(
		query,
		message.AggregateType,
		message.AggregateID,
		message.EventType,
		message.Payload,
		message.Processed,
	).Scan(&message.ID, &message.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create outbox message: %w", err)
	}

	return nil
}

func (r *outboxRepository) GetUnprocessed(limit int) ([]*models.OutboxMessage, error) {
	query := `
		SELECT id, aggregate_type, aggregate_id, event_type, payload, created_at, processed_at, processed
		FROM outbox
		WHERE processed = false
		ORDER BY created_at ASC
		LIMIT $1
	`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get unprocessed messages: %w", err)
	}
	defer rows.Close()

	var messages []*models.OutboxMessage
	for rows.Next() {
		message := &models.OutboxMessage{}
		err := rows.Scan(
			&message.ID,
			&message.AggregateType,
			&message.AggregateID,
			&message.EventType,
			&message.Payload,
			&message.CreatedAt,
			&message.ProcessedAt,
			&message.Processed,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan outbox message: %w", err)
		}
		messages = append(messages, message)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return messages, nil
}

func (r *outboxRepository) MarkAsProcessed(tx *sql.Tx, id int64) error {
	query := `
		UPDATE outbox
		SET processed = true, processed_at = $1
		WHERE id = $2 AND processed = false
	`

	result, err := tx.Exec(query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to mark message as processed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return models.ErrAlreadyProcessed
	}

	return nil
}

func (r *outboxRepository) GetByID(id int64) (*models.OutboxMessage, error) {
	query := `
		SELECT id, aggregate_type, aggregate_id, event_type, payload, created_at, processed_at, processed
		FROM outbox
		WHERE id = $1
	`

	message := &models.OutboxMessage{}
	err := r.db.QueryRow(query, id).Scan(
		&message.ID,
		&message.AggregateType,
		&message.AggregateID,
		&message.EventType,
		&message.Payload,
		&message.CreatedAt,
		&message.ProcessedAt,
		&message.Processed,
	)

	if err == sql.ErrNoRows {
		return nil, models.ErrOutboxMessageNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get outbox message: %w", err)
	}

	return message, nil
}
