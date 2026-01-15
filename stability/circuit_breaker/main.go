package main

import (
	"errors"
	"fmt"
	"log"
	"time"
)

var failCount = 0

func unreliableService() error {
	failCount++
	if failCount%3 == 0 {
		return nil
	}
	return errors.New("service failed")
}

func main() {
	fmt.Println("Circuit Breaker Demo")

	cb := NewCircuitBreaker(3, 5*time.Second)

	for i := 1; i <= 10; i++ {
		fmt.Printf("Request %d: ", i)

		err := cb.Call(func() error {
			return unreliableService()
		})

		if err == ErrCircuitOpen {
			log.Printf("Circuit is OPEN, request blocked\n")
		} else if err != nil {
			log.Printf("Request failed: %v\n", err)
		} else {
			log.Printf("Request succeeded\n")
		}

		time.Sleep(1 * time.Second)
	}
}
