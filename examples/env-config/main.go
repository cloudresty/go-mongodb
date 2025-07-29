package main

import (
	"context"
	"log"
	"os"

	"github.com/cloudresty/go-mongodb"
)

func main() {
	log.Println("Starting environment configuration examples")

	// Example 1: Using default MONGODB_ prefix
	log.Println("Creating client from environment variables (MONGODB_ prefix)")

	client, err := mongodb.NewClient()
	if err != nil {
		log.Printf("Failed to create client from environment: %v", err)
		os.Exit(1)
	}
	defer client.Close()

	log.Println("Client created successfully from environment variables")

	// Example 2: Using custom prefix (e.g., MYAPP_MONGODB_)
	log.Println("Creating client from environment variables with custom prefix")

	clientWithPrefix, err := mongodb.NewClient(mongodb.FromEnvWithPrefix("MYAPP_"))
	if err != nil {
		log.Printf("Failed to create client from environment with prefix: %v", err)
		log.Println("Custom prefix example failed (expected if MYAPP_* vars not set)")
	} else {
		defer clientWithPrefix.Close()
		log.Println("Client with custom prefix created successfully")
	}

	// Example 3: Environment variables with custom config overrides
	log.Println("Creating client from environment with custom overrides")

	clientWithConfig, err := mongodb.NewClient(
		mongodb.FromEnv(), // Load from environment
		mongodb.WithAppName("custom-env-app"),
		mongodb.WithTimeout(5000),
	)
	if err != nil {
		log.Printf("Failed to create client with environment config: %v", err)
		os.Exit(1)
	}
	defer clientWithConfig.Close()

	log.Println("Created client from environment with custom overrides")

	log.Println("Client with customized config created successfully")

	// Test the connections
	if err := client.Ping(context.Background()); err != nil {
		log.Printf("Default client ping failed: %v", err)
	} else {
		log.Println("Default client ping successful")
	}

	if clientWithPrefix != nil {
		if err := clientWithPrefix.Ping(context.Background()); err != nil {
			log.Printf("Custom client ping failed: %v", err)
		} else {
			log.Println("Custom client ping successful")
		}
	}

	if err := clientWithConfig.Ping(context.Background()); err != nil {
		log.Printf("Configured client ping failed: %v", err)
	} else {
		log.Println("Configured client ping successful")
	}

	log.Println("Environment configuration examples completed")
}
