package main

import (
	"fmt"
	"log"
)

func main() {
	recService := NewRecommendationService(0.2)
	reviewService := NewReviewService(0.7)
	productService := NewProductService(recService, reviewService)

	for i := 1; i <= 5; i++ {
		fmt.Printf("Example %d:\n", i)

		result, err := productService.GetProduct("product-1")

		if err != nil {
			log.Printf("Failed to get product: %v\n", err)
			fmt.Println(string(make([]byte, 50)))
			fmt.Println()
			continue
		}

		fmt.Printf("Product: %s\n", result.ProductID)
		fmt.Printf("Price: $%.2f\n", result.Price)
		fmt.Printf("Recommendations: %v\n", result.Recommendations)
		fmt.Printf("Review: %s\n", result.Review)

		if result.IsGD {
			fmt.Printf("Status: isGD=true (review from cache)\n")
		} else {
			fmt.Printf("Status: isGD=false (review is live)\n")
		}

		fmt.Println(string(make([]byte, 50)))
		fmt.Println()
	}
}
