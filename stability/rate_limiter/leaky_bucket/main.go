package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	fmt.Println("Leaky Bucket Rate Limiter Demo")
	fmt.Println("===============================\n")

	limiter := NewLeakyBucket(10, 5)

	mux := http.NewServeMux()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Request processed successfully"))
	})

	mux.Handle("/api", RateLimitMiddleware(limiter)(handler))

	port := ":8082"
	fmt.Printf("Server starting on http://localhost%s\n", port)
	fmt.Println("Limit: 10 requests queue, leak rate: 5 requests/sec")
	fmt.Printf("Try: curl http://localhost%s/api\n", port)

	log.Fatal(http.ListenAndServe(port, mux))
}
