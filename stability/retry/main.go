package main

import (
	"context"
	"fmt"
)

func main() {
	fmt.Println("Retry Pattern Demo")
	fmt.Println("======================\n")

	ctx := context.Background()

	apiClient := NewExternalAPIClient(0.7)

	dataService := NewDataService(apiClient)

	fmt.Println("Example 1:")
	dataService.GetData(ctx, "user-123")

	fmt.Println("\n" + string(make([]byte, 50)))

	fmt.Println("\nExample 2:")
	dataService.GetData(ctx, "order-456")

	fmt.Println("\n" + string(make([]byte, 50)))

	fmt.Println("\nExample 3:")
	dataService.GetData(ctx, "product-789")
}
