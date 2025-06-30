package main

import (
	"fmt"
	"log"
	"os"

	"github.com/cloudresty/go-mongodb"
	"github.com/cloudresty/go-mongodb/mongoid"
)

type User struct {
	Name  string `bson:"name" json:"name"`
	Email string `bson:"email" json:"email"`
	Role  string `bson:"role" json:"role"`
}

type Order struct {
	ID     string  `bson:"_id" json:"id"`
	Amount float64 `bson:"amount" json:"amount"`
	Status string  `bson:"status" json:"status"`
}

func main() {
	fmt.Println("MongoDB ID Mode Configuration Demo")
	fmt.Println("==================================")

	// Demo 1: ULID Mode (Default)
	fmt.Println("\n1. ULID Mode (Default)")
	demoULIDMode()

	// Demo 2: Environment Variable Configuration
	fmt.Println("\n2. Environment Variable Configuration")
	demoEnvironmentConfig()

	// Demo 3: ULID Generation
	fmt.Println("\n3. ULID Generation Examples")
	demoULIDGeneration()
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

	// Type-safe struct instead of raw map
	user := User{
		Name:  "John Doe",
		Email: "john@example.com",
		Role:  "user",
	}

	// Simulate document enhancement (ULID generation)
	generatedULID := mongoid.NewULID() // Generate actual ULID
	fmt.Printf("   Generated ID: %s (length: %d)\n", generatedULID, len(generatedULID))
	fmt.Printf("   ID Type: ULID string\n")
	fmt.Printf("   Sample user: %+v\n", user)
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
	fmt.Println("   export MONGODB_ID_MODE=custom    # User-provided IDs")

	// Demo loading client with environment configuration
	client, err := mongodb.NewClient(mongodb.FromEnv())
	if err != nil {
		log.Printf("Error creating client: %v", err)
		return
	}
	defer client.Close()

	fmt.Printf("   Client created with environment-configured ID mode\n")
}

func demoULIDGeneration() {
	fmt.Println("   Generating multiple ULIDs to show uniqueness and sorting:")

	for i := range 5 {
		ulid := mongoid.NewULID()
		fmt.Printf("   ULID %d: %s\n", i+1, ulid)
	}

	// Show parsing
	ulid := mongoid.NewULID()
	parsed, err := mongoid.ParseULID(ulid)
	if err != nil {
		log.Printf("Error parsing ULID: %v", err)
		return
	}

	fmt.Printf("\n   ULID Details for %s:\n", ulid)
	fmt.Printf("   Timestamp: %v\n", parsed.Time())
	fmt.Printf("   String: %s\n", parsed.String())

	// Example with user-provided ID using type-safe struct
	order := Order{
		ID:     "order-" + mongoid.NewULID(),
		Amount: 250.00,
		Status: "pending",
	}

	fmt.Printf("\n   Custom ID with ULID suffix: %s\n", order.ID)
	fmt.Printf("   Sample order: %+v\n", order)
}
