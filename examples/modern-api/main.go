// go-mongodb v2 Example: Modern API
//
// This example demonstrates the v2 API with:
// - Fluent filter and update builders
// - ULID-based document IDs
// - Type-safe operations
//
// v2 Changes:
// - Use col.FindByID(ctx, id) instead of mongoid.FindByULID()
// - mongoid.NewULID() panics on entropy failure (use NewULIDWithError() for explicit handling)
// - Strict type validation rejects incompatible ID field types in ULID mode
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/cloudresty/go-mongodb/v2"
	"github.com/cloudresty/go-mongodb/v2/filter"
	"github.com/cloudresty/go-mongodb/v2/mongoid"
	"github.com/cloudresty/go-mongodb/v2/update"
)

// User represents a user document
type User struct {
	ID         string    `bson:"_id"`
	Name       string    `bson:"name"`
	Email      string    `bson:"email"`
	Status     string    `bson:"status"`
	Age        int       `bson:"age"`
	Tags       []string  `bson:"tags"`
	CreatedAt  time.Time `bson:"created_at"`
	UpdatedAt  time.Time `bson:"updated_at"`
	LoginCount int       `bson:"login_count"`
}

func main() {
	// Demo 1: Modern Client Creation with Functional Options
	fmt.Println("=== Demo 1: Modern Client Creation (Target API) ===")

	fmt.Println("Target API (to be implemented):")
	fmt.Println("// Method 1: Environment-first configuration")
	fmt.Println("client, err := mongodb.NewClient(")
	fmt.Println("    mongodb.FromEnv(),                    // Load from environment")
	fmt.Println("    mongodb.WithAppName(\"modern-demo\"),   // Override/supplement")
	fmt.Println("    mongodb.WithTimeout(30*time.Second),  // Additional config")
	fmt.Println(")")
	fmt.Println("")
	fmt.Println("// Method 2: Pure code configuration")
	fmt.Println("client, err := mongodb.NewClient(")
	fmt.Println("    mongodb.WithHosts(\"localhost:27017\"),")
	fmt.Println("    mongodb.WithCredentials(\"user\", \"pass\"),")
	fmt.Println("    mongodb.WithDatabase(\"myapp\"),")
	fmt.Println("    mongodb.WithMaxPoolSize(100),")
	fmt.Println(")")

	// For now, create client using existing API
	client, err := mongodb.NewClient()
	if err != nil {
		log.Printf("Failed to create client with current API: %v", err)
		fmt.Println("Note: Using existing API for demonstration")
		return
	}
	defer func() { _ = client.Close() }()

	fmt.Println("✓ Client created using current API (will be enhanced)")

	// Demo 2: Fluent Filter Builder
	fmt.Println("\n=== Demo 2: Fluent Filter Builders ===")

	// Simple equality filter
	simpleFilter := filter.Eq("status", "active")
	fmt.Printf("Simple filter: %v\n", simpleFilter.Build())

	// Complex nested filter using fluent methods
	complexFilter := filter.Eq("status", "active").
		And(filter.Gt("age", 21)).
		And(
			filter.In("tags", "golang", "mongodb").
				Or(filter.Regex("name", "^John", "i")),
		)
	fmt.Printf("Complex filter: %v\n", complexFilter.Build())

	// Array operations using fluent methods
	arrayFilter := filter.Size("tags", 3).
		And(filter.ElemMatch("orders",
			filter.Gt("amount", 100).
				And(filter.Eq("status", "completed")),
		))
	fmt.Printf("Array filter: %v\n", arrayFilter.Build())

	// Demo 3: Fluent Update Builder
	fmt.Println("\n=== Demo 3: Fluent Update Builders ===")

	// Simple update
	simpleUpdate := update.Set("status", "updated")
	fmt.Printf("Simple update: %v\n", simpleUpdate.Build())

	// Complex chained update
	complexUpdate := update.New().
		Set("status", "active").
		Set("updated_at", time.Now()).
		Inc("login_count", 1).
		Push("activity_log", "user_updated").
		Unset("temp_field")
	fmt.Printf("Complex update: %v\n", complexUpdate.Build())

	// Array updates
	arrayUpdate := update.New().
		PushEach("tags", "new-tag1", "new-tag2").
		AddToSet("categories", "premium").
		Pull("old_tags", filter.Eq("deprecated", true))
	fmt.Printf("Array update: %v\n", arrayUpdate.Build())

	// Demo 4: ULID Operations
	fmt.Println("\n=== Demo 4: ULID Operations ===")

	// Generate ULIDs
	id1 := mongoid.NewULID()
	id2 := mongoid.NewULID()
	fmt.Printf("Generated ULID 1: %s\n", id1)
	fmt.Printf("Generated ULID 2: %s\n", id2)

	// Parse ULID
	parsedULID, err := mongoid.ParseULID(id1)
	if err != nil {
		log.Printf("Failed to parse ULID: %v", err)
	} else {
		fmt.Printf("Parsed ULID: %s, Time: %v\n", parsedULID.String(), parsedULID.Time())
	}

	// Demo 5: Complete CRUD Example
	fmt.Println("\n=== Demo 5: Complete CRUD Example ===")

	// Note: This is a demonstration of API usage
	// In a real application, you would handle connection errors appropriately

	fmt.Println("API Demonstration (showing intended usage):")

	// Resource-oriented API access
	fmt.Println("1. Accessing collections via resource hierarchy:")
	fmt.Println("   users := client.Database(\"myapp\").Collection(\"users\")")

	// Creating documents with ULID
	user := User{
		ID:         mongoid.NewULID(),
		Name:       "John Doe",
		Email:      "john@example.com",
		Status:     "active",
		Age:        30,
		Tags:       []string{"golang", "mongodb"},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		LoginCount: 0,
	}
	fmt.Printf("2. Created user with ULID: %s\n", user.ID)

	// Insert operation (simulated)
	fmt.Println("3. Insert operation:")
	fmt.Println("   result, err := users.InsertOne(ctx, user)")

	// Find with complex filter using fluent methods (simulated)
	findFilter := filter.Eq("status", "active").
		And(filter.Gt("age", 25))
	fmt.Printf("4. Find with filter: %v\n", findFilter.Build())
	fmt.Println("   cursor, err := users.Find(ctx, findFilter)")

	// Update with fluent builder (simulated)
	updateDoc := update.New().
		Set("last_login", time.Now()).
		Inc("login_count", 1).
		Push("activity", "login")
	fmt.Printf("5. Update document: %v\n", updateDoc.Build())
	fmt.Println("   result, err := users.UpdateOne(ctx, filter.Eq(\"_id\", user.ID), updateDoc)")

	// Transaction example (simulated)
	fmt.Println("6. Transaction example:")
	fmt.Println("   err := client.WithTransaction(ctx, func(ctx context.Context) error {")
	fmt.Println("       // Multiple operations within transaction")
	fmt.Println("       return nil")
	fmt.Println("   })")

	// Demo 6: Modern Error Handling
	fmt.Println("\n=== Demo 6: Modern Error Handling ===")
	fmt.Println("Structured error checking (intended usage):")
	fmt.Println("if mongodb.IsDuplicateKeyError(err) {")
	fmt.Println("    // Handle duplicate key")
	fmt.Println("} else if mongodb.IsNotFoundError(err) {")
	fmt.Println("    // Handle not found")
	fmt.Println("}")

	// Demo 7: Connection Management
	fmt.Println("\n=== Demo 7: Connection Management ===")
	fmt.Println("Connection health and monitoring (intended usage):")
	fmt.Println("err := client.Ping(ctx)")
	fmt.Println("stats := client.Stats()")
	fmt.Println("// Access stats.ActiveConnections, stats.OperationsExecuted, etc.")

	fmt.Println("\n=== Demo Complete ===")
	fmt.Println("This demonstrates the modern, fluent API design for go-mongodb")
	fmt.Println("Key benefits:")
	fmt.Println("✓ Functional options for flexible client configuration")
	fmt.Println("✓ Type-safe, fluent builders eliminate raw bson.M usage")
	fmt.Println("✓ Resource-oriented API follows MongoDB hierarchy")
	fmt.Println("✓ Clean separation of concerns with sub-packages")
	fmt.Println("✓ ULID support for better performance and sorting")
	fmt.Println("✓ Modern error handling with structured types")
}
