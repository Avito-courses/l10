package main

import (
	"errors"
	"log"
	"math/rand"
)

type RecommendationService struct {
	failureRate float32
}

func NewRecommendationService(failureRate float32) *RecommendationService {
	return &RecommendationService{
		failureRate: failureRate,
	}
}

func (r *RecommendationService) GetRecommendations(productID string) ([]string, error) {
	if rand.Float32() < r.failureRate {
		return nil, errors.New("recommendation service unavailable")
	}
	return []string{"Product A", "Product B", "Product C"}, nil
}

type ReviewService struct {
	failureRate float32
}

func NewReviewService(failureRate float32) *ReviewService {
	return &ReviewService{
		failureRate: failureRate,
	}
}

func (r *ReviewService) GetReviews(productID string) (string, error) {
	if rand.Float32() < r.failureRate {
		return "", errors.New("review service unavailable")
	}
	return "Excellent product!", nil
}

type ProductService struct {
	recommendationService *RecommendationService
	reviewService         *ReviewService
	cachedReviews         map[string]string
}

func NewProductService(recService *RecommendationService, revService *ReviewService) *ProductService {
	return &ProductService{
		recommendationService: recService,
		reviewService:         revService,
		cachedReviews: map[string]string{
			"product-1": "Good (cached)",
			"product-2": "Very good (cached)",
			"product-3": "Excellent (cached)",
		},
	}
}

type ProductResult struct {
	ProductID       string
	Price           float64
	Recommendations []string
	Review          string
	IsGD            bool
}

func (s *ProductService) GetProduct(productID string) (ProductResult, error) {
	log.Printf("Getting product: %s", productID)

	price := 99.99

	log.Printf("  Getting recommendations...")
	recommendations, err := s.recommendationService.GetRecommendations(productID)
	if err != nil {
		log.Printf("  Recommendation service failed: %v", err)
		return ProductResult{}, err
	}
	log.Printf("  Recommendations OK")

	log.Printf("  Getting reviews...")
	review, err := s.reviewService.GetReviews(productID)
	isGD := false

	if err != nil {
		log.Printf("  Review service failed: %v", err)
		log.Printf("  Using cached review (fallback, isGD=true)")

		if cachedReview, ok := s.cachedReviews[productID]; ok {
			review = cachedReview
		} else {
			review = "No reviews available"
		}
		isGD = true
	} else {
		log.Printf("  Review OK (isGD=false)")
	}

	log.Println()

	return ProductResult{
		ProductID:       productID,
		Price:           price,
		Recommendations: recommendations,
		Review:          review,
		IsGD:            isGD,
	}, nil
}
