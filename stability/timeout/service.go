package main

import (
	"context"
	"log"
	"time"
)

type DatabaseClient struct {
	averageDelay time.Duration
}

func NewDatabaseClient(avgDelay time.Duration) *DatabaseClient {
	return &DatabaseClient{
		averageDelay: avgDelay,
	}
}

func (c *DatabaseClient) Query(query string) (string, error) {
	delay := c.averageDelay + time.Duration(time.Now().UnixNano()%int64(c.averageDelay))
	log.Printf("    Query will take %v...", delay)
	time.Sleep(delay)

	return "result for: " + query, nil
}

type UserService struct {
	db           *DatabaseClient
	queryTimeout time.Duration
}

func NewUserService(db *DatabaseClient, timeout time.Duration) *UserService {
	return &UserService{
		db:           db,
		queryTimeout: timeout,
	}
}

func (s *UserService) GetUser(ctx context.Context, userID string) (string, error) {
	log.Printf("Getting user %s (timeout: %v)", userID, s.queryTimeout)

	start := time.Now()

	timeoutCtx, cancel := context.WithTimeout(ctx, s.queryTimeout)
	defer cancel()

	resultChan := make(chan string, 1)
	errChan := make(chan error, 1)

	go func() {
		result, err := s.db.Query("SELECT * FROM users WHERE id = " + userID)
		if err != nil {
			errChan <- err
			return
		}
		resultChan <- result
	}()

	select {
	case result := <-resultChan:
		elapsed := time.Since(start)
		log.Printf("  Success in %v: %s\n", elapsed, result)
		return result, nil

	case err := <-errChan:
		log.Printf("  Error: %v\n", err)
		return "", err

	case <-timeoutCtx.Done():
		elapsed := time.Since(start)
		log.Printf("  Timeout after %v\n", elapsed)
		return "", ErrTimeout
	}
}
