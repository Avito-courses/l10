package main

import (
	"context"
	"errors"
	"log"
	"time"
)

type ExternalAPIClient struct {
	failureRate float32
}

func NewExternalAPIClient(failureRate float32) *ExternalAPIClient {
	return &ExternalAPIClient{
		failureRate: failureRate,
	}
}

func (c *ExternalAPIClient) GetData(id string) (string, error) {
	if time.Now().UnixNano()%100 < int64(c.failureRate*100) {
		return "", errors.New("network error: connection timeout")
	}

	return "data for " + id, nil
}

type DataService struct {
	apiClient *ExternalAPIClient
	retry     *RetryExecutor
}

func NewDataService(apiClient *ExternalAPIClient) *DataService {
	retryConfig := RetryConfig{
		MaxAttempts: 5,
		Strategy:    NewExponentialBackoff(100*time.Millisecond, 5*time.Second, 2.0),
	}

	return &DataService{
		apiClient: apiClient,
		retry:     NewRetryExecutor(retryConfig),
	}
}

func (s *DataService) GetData(ctx context.Context, id string) (string, error) {
	log.Printf("Getting data for ID: %s", id)

	var result string
	var attempts int

	err := s.retry.ExecuteWithCallback(
		func() error {
			attempts++
			log.Printf("  Attempt %d...", attempts)

			data, err := s.apiClient.GetData(id)
			if err != nil {
				log.Printf("    Failed: %v", err)
				return err
			}

			result = data
			log.Printf("    Success!")
			return nil
		},
		func(attempt int, err error, delay time.Duration) {
			log.Printf("  Retrying in %v...", delay)
		},
	)

	if err != nil {
		log.Printf("All attempts failed")
		return "", err
	}

	log.Printf("Data retrieved after %d attempts: %s\n", attempts, result)
	return result, nil
}
