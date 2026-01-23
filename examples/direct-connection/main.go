package main

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudresty/go-mongodb/v2"
	"github.com/cloudresty/go-mongodb/v2/filter"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func main() {
	fmt.Println("=== MongoDB Direct Connection Example ===")

	// Example 1: Using functional options with direct connection
	fmt.Println("\n1. Using functional options with direct connection:")

	client1, err := mongodb.NewClient(
		mongodb.WithHosts("localhost:27017"),
		mongodb.WithDatabase("direct_test"),
		mongodb.WithCredentials("admin", "password"),
		mongodb.WithDirectConnection(true), // Enable direct connection
	)
	if err != nil {
		log.Printf("Failed to create client with direct connection: %v", err)
	} else {
		fmt.Println("✓ Client created with direct connection")
		defer func() { _ = client1.Close() }()
	}

	// Example 2: Using environment variables with direct connection
	fmt.Println("\n2. Using environment variables:")
	fmt.Println("Set MONGODB_DIRECT_CONNECTION=true to enable direct connection")

	// This would use environment variables:
	// MONGODB_HOSTS=localhost:27017
	// MONGODB_DATABASE=direct_test
	// MONGODB_USERNAME=admin
	// MONGODB_PASSWORD=password
	// MONGODB_DIRECT_CONNECTION=true
	/*
		client2, err := mongodb.NewClient(mongodb.FromEnv())
		if err != nil {
			log.Printf("Failed to create client from env: %v", err)
		} else {
			fmt.Println("✓ Client created from environment with direct connection")
			defer client2.Close()
		}
	*/

	// Example 3: When to use direct connection
	fmt.Println("\n3. When to use direct connection:")
	fmt.Println("   - Connecting to a standalone MongoDB instance")
	fmt.Println("   - Connecting to a specific replica set member")
	fmt.Println("   - Testing/development environments")
	fmt.Println("   - When you want to bypass replica set discovery")

	// Example 4: Demonstrate the difference in connection URIs
	fmt.Println("\n4. Connection URI examples:")

	// Without direct connection
	config1 := &mongodb.Config{}
	mongodb.WithHosts("localhost:27017")(config1)
	mongodb.WithDatabase("test")(config1)
	mongodb.WithCredentials("admin", "password")(config1)
	fmt.Printf("   Without direct connection: %s\n", config1.BuildConnectionURI())

	// With direct connection
	config2 := &mongodb.Config{}
	mongodb.WithHosts("localhost:27017")(config2)
	mongodb.WithDatabase("test")(config2)
	mongodb.WithCredentials("admin", "password")(config2)
	mongodb.WithDirectConnection(true)(config2)
	fmt.Printf("   With direct connection:    %s\n", config2.BuildConnectionURI())

	// Example 5: Basic CRUD operations with direct connection
	if client1 != nil {
		fmt.Println("\n5. Testing basic operations with direct connection:")

		collection := client1.Collection("test_direct")
		ctx := context.Background()

		// Insert a test document
		testDoc := bson.M{
			"name":       "Direct Connection Test",
			"timestamp":  bson.A{2025, 7, 6},
			"connection": "direct",
		}

		result, err := collection.InsertOne(ctx, testDoc)
		if err != nil {
			log.Printf("Failed to insert document: %v", err)
		} else {
			fmt.Printf("   ✓ Document inserted with ID: %s\n", result.InsertedID)

			// Find the document
			var foundDoc bson.M
			err = collection.FindOne(ctx, filter.Eq("name", "Direct Connection Test")).Decode(&foundDoc)
			if err != nil {
				log.Printf("Failed to find document: %v", err)
			} else {
				fmt.Printf("   ✓ Document found: %v\n", foundDoc["name"])
			}

			// Clean up
			_, err = collection.DeleteOne(ctx, filter.Eq("name", "Direct Connection Test"))
			if err != nil {
				log.Printf("Failed to delete document: %v", err)
			} else {
				fmt.Println("   ✓ Test document cleaned up")
			}
		}
	}

	fmt.Println("\n=== Example completed ===")
}
