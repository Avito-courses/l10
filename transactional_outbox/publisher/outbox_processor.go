package publisher

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"transactional_outbox/models"
	"transactional_outbox/repository"
)

type OutboxProcessor struct {
	db             *sql.DB
	outboxRepo     repository.OutboxRepository
	kafkaPublisher *KafkaPublisher
	pollInterval   time.Duration
	batchSize      int
}

func NewOutboxProcessor(
	db *sql.DB,
	outboxRepo repository.OutboxRepository,
	kafkaPublisher *KafkaPublisher,
	pollInterval time.Duration,
	batchSize int,
) *OutboxProcessor {
	return &OutboxProcessor{
		db:             db,
		outboxRepo:     outboxRepo,
		kafkaPublisher: kafkaPublisher,
		pollInterval:   pollInterval,
		batchSize:      batchSize,
	}
}

func (p *OutboxProcessor) Start(ctx context.Context) {
	log.Println("Outbox processor started")

	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()

	p.processMessages()

	for {
		select {
		case <-ctx.Done():
			log.Println("Outbox processor stopped")
			return
		case <-ticker.C:
			p.processMessages()
		}
	}
}

func (p *OutboxProcessor) processMessages() {
	messages, err := p.outboxRepo.GetUnprocessed(p.batchSize)
	if err != nil {
		log.Printf("Failed to get unprocessed messages: %v", err)
		return
	}

	if len(messages) == 0 {
		return
	}

	log.Printf("Found %d unprocessed messages in outbox", len(messages))

	for _, msg := range messages {
		if err := p.processMessage(msg); err != nil {
			log.Printf("Failed to process message ID=%d: %v", msg.ID, err)
			continue
		}
	}
}

func (p *OutboxProcessor) processMessage(msg *models.OutboxMessage) error {
	log.Printf("Processing outbox message: ID=%d, Type=%s, AggregateID=%s",
		msg.ID, msg.EventType, msg.AggregateID)

	key := fmt.Sprintf("%s-%s", msg.AggregateType, msg.AggregateID)
	err := p.kafkaPublisher.Publish(key, msg.Payload)
	if err != nil {
		return fmt.Errorf("failed to publish to Kafka: %w", err)
	}

	tx, err := p.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	err = p.outboxRepo.MarkAsProcessed(tx, msg.ID)
	if err != nil {
		return fmt.Errorf("failed to mark as processed: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Message ID=%d processed and published successfully", msg.ID)
	return nil
}
