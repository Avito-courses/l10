package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	fmt.Println("Fixed Window Rate Limiter Demo")
	fmt.Println("===============================\n")

	limiter := NewFixedWindow(10, 10*time.Second)

	mux := http.NewServeMux()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Request processed successfully"))
	})

	mux.Handle("/api", RateLimitMiddleware(limiter)(handler))

	port := ":8083"
	fmt.Printf("Server starting on http://localhost%s\n", port)
	fmt.Println("Limit: 10 requests per 10 seconds")
	fmt.Printf("Try: curl http://localhost%s/api\n", port)

	log.Fatal(http.ListenAndServe(port, mux))
}
