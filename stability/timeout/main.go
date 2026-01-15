package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	fmt.Println("Timeout Pattern Demo")
	fmt.Println("====================\n")

	ctx := context.Background()

	dbClient := NewDatabaseClient(2 * time.Second)

	userService := NewUserService(dbClient, 3*time.Second)

	fmt.Println("Example 1: Fast query (should succeed)")
	userService.GetUser(ctx, "123")

	fmt.Println(string(make([]byte, 50)))

	fmt.Println("\nExample 2: Medium query")
	userService.GetUser(ctx, "456")

	fmt.Println(string(make([]byte, 50)))

	fastService := NewUserService(dbClient, 1*time.Second)

	fmt.Println("\nExample 3: Slow query with short timeout (should timeout)")
	fastService.GetUser(ctx, "789")
}
