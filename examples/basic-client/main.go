// go-mongodb v2 Example: Basic Client
//
// This example demonstrates basic MongoDB client setup with v2.
// The client uses environment variables for configuration.
package main

import (
	"context"
	"log"
	"os"

	"github.com/cloudresty/go-mongodb/v2"
)

func main() {
	log.Println("Starting basic MongoDB client example")

	// Create client from environment variables - uses NopLogger by default (silent)
	client, err := mongodb.NewClient()
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		log.Println("Hint: Set MONGODB_* environment variables or defaults will be used")
		os.Exit(1)
	}
	defer func() { _ = client.Close() }()

	log.Println("MongoDB client connected successfully")

	// Test connection
	ctx := context.Background()
	err = client.Ping(ctx)
	if err != nil {
		log.Printf("Failed to ping MongoDB: %v", err)
		os.Exit(1)
	}

	log.Println("MongoDB ping successful")

	// Get database and collection
	db := client.Database("testdb")
	_ = db.Collection("testcollection") // Just to show usage

	log.Printf("Connected to database=%s and collection=%s", "testdb", "testcollection")

	log.Println("Basic client example completed successfully!")
}
