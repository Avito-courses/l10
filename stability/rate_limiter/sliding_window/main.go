package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	fmt.Println("Sliding Window Rate Limiter Demo")
	fmt.Println("=================================\n")

	limiter := NewSlidingWindow(10, 10*time.Second)

	mux := http.NewServeMux()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Request processed successfully"))
	})

	mux.Handle("/api", RateLimitMiddleware(limiter)(handler))

	port := ":8084"
	fmt.Printf("Server starting on http://localhost%s\n", port)
	fmt.Println("Limit: 10 requests in sliding 10 second window")
	fmt.Printf("Try: curl http://localhost%s/api\n", port)

	log.Fatal(http.ListenAndServe(port, mux))
}
