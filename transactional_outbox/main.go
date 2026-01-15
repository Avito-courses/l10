package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"transactional_outbox/models"
	"transactional_outbox/publisher"
	"transactional_outbox/repository"
	"transactional_outbox/service"

	_ "github.com/lib/pq"
)

const (
	dbHost     = "localhost"
	dbPort     = 5432
	dbUser     = "outbox_user"
	dbPassword = "outbox_password"
	dbName     = "outbox_db"

	kafkaBroker = "localhost:9092"
	kafkaTopic  = "order-events"

	pollInterval = 2 * time.Second
	batchSize    = 10
)

func main() {
	log.Println("=== Transactional Outbox Pattern Demo ===")
	log.Println()

	db, err := connectToDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	orderRepo := repository.NewOrderRepository(db)
	outboxRepo := repository.NewOutboxRepository(db)

	orderService := service.NewOrderService(db, orderRepo, outboxRepo)

	kafkaPublisher, err := publisher.NewKafkaPublisher([]string{kafkaBroker}, kafkaTopic)
	if err != nil {
		log.Fatalf("Failed to create Kafka publisher: %v", err)
	}
	defer kafkaPublisher.Close()

	outboxProcessor := publisher.NewOutboxProcessor(
		db,
		outboxRepo,
		kafkaPublisher,
		pollInterval,
		batchSize,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go outboxProcessor.Start(ctx)

	time.Sleep(1 * time.Second)

	demonstrateTransactionalOutbox(orderService, orderRepo)

	log.Println()
	log.Println("Waiting for outbox processor to publish messages to Kafka...")
	time.Sleep(5 * time.Second)

	showStatistics(db, orderRepo, outboxRepo)

	log.Println()
	log.Println("Press Ctrl+C to exit...")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println()
	log.Println("Shutting down...")
	cancel()
	time.Sleep(1 * time.Second)
	log.Println("Goodbye!")
}

func connectToDatabase() (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName,
	)

	log.Println("Connecting to PostgreSQL...")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		err = db.Ping()
		if err == nil {
			log.Println("Connected to PostgreSQL successfully")
			return db, nil
		}
		log.Printf("Waiting for PostgreSQL... (attempt %d/%d)", i+1, maxRetries)
		time.Sleep(1 * time.Second)
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
}

func demonstrateTransactionalOutbox(orderService *service.OrderService, orderRepo repository.OrderRepository) {
	log.Println()
	log.Println("=== Демонстрация Transactional Outbox ===")
	log.Println()

	orders := []models.CreateOrderRequest{
		{
			CustomerID:  "customer-001",
			ProductName: "Ноутбук MacBook Pro",
			Quantity:    1,
			TotalPrice:  2500.00,
		},
		{
			CustomerID:  "customer-002",
			ProductName: "iPhone 15 Pro",
			Quantity:    2,
			TotalPrice:  2400.00,
		},
		{
			CustomerID:  "customer-003",
			ProductName: "AirPods Pro",
			Quantity:    3,
			TotalPrice:  750.00,
		},
	}

	log.Println("Создаем заказы (в одной транзакции с outbox)...")
	log.Println()

	var createdOrders []*models.Order
	for i, req := range orders {
		log.Printf("--- Заказ #%d ---", i+1)
		order, err := orderService.CreateOrder(&req)
		if err != nil {
			log.Printf("Failed to create order: %v", err)
			continue
		}

		createdOrders = append(createdOrders, order)
		log.Printf("Order ID=%d создан для клиента %s", order.ID, order.CustomerID)
		log.Println()

		time.Sleep(500 * time.Millisecond)
	}

	if len(createdOrders) > 0 {
		log.Println("--- Подтверждение заказа ---")
		orderID := createdOrders[0].ID
		err := orderService.ConfirmOrder(orderID)
		if err != nil {
			log.Printf("Failed to confirm order: %v", err)
		} else {
			log.Printf("Order ID=%d подтвержден", orderID)
		}
		log.Println()
	}
}

func showStatistics(db *sql.DB, orderRepo repository.OrderRepository, outboxRepo repository.OutboxRepository) {
	log.Println()

	orders, err := orderRepo.List(100, 0)
	if err != nil {
		log.Printf("Failed to get orders: %v", err)
	} else {
		log.Printf("Всего заказов: %d", len(orders))
		for _, order := range orders {
			log.Printf("   - Order #%d: %s, Status: %s, Price: $%.2f",
				order.ID, order.ProductName, order.Status, order.TotalPrice)
		}
	}

	log.Println()

	unprocessed, err := outboxRepo.GetUnprocessed(100)
	if err != nil {
		log.Printf("Failed to get unprocessed messages: %v", err)
	} else {
		log.Printf("Необработанных сообщений в outbox: %d", len(unprocessed))
	}

	var totalOutbox, processedOutbox int
	err = db.QueryRow("SELECT COUNT(*) FROM outbox").Scan(&totalOutbox)
	if err != nil {
		log.Printf("Failed to count outbox messages: %v", err)
	}

	err = db.QueryRow("SELECT COUNT(*) FROM outbox WHERE processed = true").Scan(&processedOutbox)
	if err != nil {
		log.Printf("Failed to count processed messages: %v", err)
	}

	log.Printf("Всего сообщений в outbox: %d", totalOutbox)
	log.Printf("Обработано и отправлено в Kafka: %d", processedOutbox)
	log.Printf("Ожидают обработки: %d", totalOutbox-processedOutbox)
}
