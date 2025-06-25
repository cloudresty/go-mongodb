package main

import (
	"fmt"
	"log"
	"os"

	"github.com/cloudresty/go-mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func main() {
	fmt.Println("MongoDB ID Mode Configuration Demo")
	fmt.Println("==================================")

	// Demo 1: ULID Mode (Default)
	fmt.Println("\n1. ULID Mode (Default)")
	demoULIDMode()

	// Demo 2: ObjectID Mode
	fmt.Println("\n2. ObjectID Mode")
	demoObjectIDMode()

	// Demo 3: Custom Mode (User-provided IDs)
	fmt.Println("\n3. Custom Mode (User-provided IDs)")
	demoCustomMode()

	// Demo 4: Environment Variable Configuration
	fmt.Println("\n4. Environment Variable Configuration")
	demoEnvironmentConfig()
}

func demoULIDMode() {
	// Default configuration uses ULID
	client, err := mongodb.NewClient()
	if err != nil {
		log.Printf("Error creating client: %v", err)
		return
	}
	defer client.Close()

	fmt.Println("   Configuration: Default (ULID)")
	fmt.Println("   Document ID will be a 26-character ULID string")

	// This would normally insert to MongoDB, but we'll just show the enhanced document
	doc := map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
		"role":  "user",
	}

	enhanced := mongodb.EnhanceDocument(doc)
	fmt.Printf("   Generated ID: %s (length: %d)\n", enhanced["_id"], len(enhanced["_id"].(string)))
	fmt.Printf("   ID Type: ULID string\n")
}

func demoObjectIDMode() {
	config := &mongodb.Config{
		Host:     "localhost",
		Port:     27017,
		Database: "demo",
		IDMode:   mongodb.IDModeObjectID,
	}

	client, err := mongodb.ConnectWithConfig(config)
	if err != nil {
		log.Printf("Error creating client: %v", err)
		return
	}
	defer client.Close()

	fmt.Println("   Configuration: ObjectID Mode")
	fmt.Println("   Document ID will be a MongoDB ObjectID")

	// Create a sample ObjectID to show the format
	objectID := bson.NewObjectID()
	fmt.Printf("   Generated ID: %s (12 bytes)\n", objectID.Hex())
	fmt.Printf("   ID Type: MongoDB ObjectID\n")
}

func demoCustomMode() {
	config := &mongodb.Config{
		Host:     "localhost",
		Port:     27017,
		Database: "demo",
		IDMode:   mongodb.IDModeCustom,
	}

	client, err := mongodb.ConnectWithConfig(config)
	if err != nil {
		log.Printf("Error creating client: %v", err)
		return
	}
	defer client.Close()

	fmt.Println("   Configuration: Custom Mode")
	fmt.Println("   No automatic ID generation - user must provide _id")

	// Example with user-provided ID
	doc := map[string]interface{}{
		"_id":    "custom-order-2023-001",
		"amount": 250.00,
		"status": "pending",
	}

	fmt.Printf("   User-provided ID: %s\n", doc["_id"])
	fmt.Printf("   ID Type: User-defined (string)\n")
}

func demoEnvironmentConfig() {
	fmt.Println("   Environment Variable: MONGODB_ID_MODE")
	fmt.Println("   Available values: ulid (default), objectid, custom")

	// Show current environment setting
	currentMode := os.Getenv("MONGODB_ID_MODE")
	if currentMode == "" {
		currentMode = "ulid (default)"
	}
	fmt.Printf("   Current setting: %s\n", currentMode)

	fmt.Println("\n   Example environment configurations:")
	fmt.Println("   export MONGODB_ID_MODE=ulid      # Use ULID strings")
	fmt.Println("   export MONGODB_ID_MODE=objectid  # Use MongoDB ObjectIDs")
	fmt.Println("   export MONGODB_ID_MODE=custom     # User-provided IDs")

	// Demo loading config from environment
	config, err := mongodb.LoadConfig()
	if err != nil {
		log.Printf("Error loading config: %v", err)
		return
	}

	fmt.Printf("   Loaded ID Mode: %s\n", config.IDMode)
}
