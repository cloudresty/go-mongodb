package main

import (
	"fmt"
	"log"

	"github.com/cloudresty/go-mongodb/v2"
	"github.com/cloudresty/go-mongodb/v2/filter"
	"github.com/cloudresty/go-mongodb/v2/update"
)

// ActivityLog represents an activity log entry
type ActivityLog struct {
	Action    string `bson:"action" json:"action"`
	Timestamp string `bson:"timestamp" json:"timestamp"`
}

func main() {
	fmt.Println("🚀 go-mongodb Modern API Demo")
	fmt.Println("===============================")

	// 1. Create client using functional options (modern API)
	fmt.Println("\n1. Creating MongoDB client with functional options...")
	client, err := mongodb.NewClient(
		mongodb.WithDatabase("demo_modern"),
		mongodb.WithAppName("modern-api-demo"),
		mongodb.WithMaxPoolSize(10),
		mongodb.WithMinPoolSize(2),
	)
	if err != nil {
		log.Printf("❌ Failed to create client: %v", err)
		fmt.Println("Note: This is expected if MongoDB is not running")
		demoBuilders()
		return
	}

	// Note: In a real app, you would defer closing the client
	// defer client.CloseWithContext(context.Background())

	fmt.Println("✅ Client created successfully")

	// 2. Get database and collection using modern API
	fmt.Println("\n2. Getting database and collection...")
	db := client.Database("demo_modern")
	fmt.Printf("📂 Database: %s\n", db.Name())

	collection := db.Collection("users")
	fmt.Printf("📄 Collection: %s\n", collection.Name())

	// 3. Demonstrate fluent builders even without MongoDB connection
	demoBuilders()

	fmt.Println("\n✨ Modern API demo completed!")
}

func demoBuilders() {
	fmt.Println("\n3. Demonstrating Fluent Filter Builders...")

	// Simple equality filter
	simpleFilter := filter.Eq("status", "active")
	fmt.Printf("Simple filter: %+v\n", simpleFilter.Build())

	// Complex filter with multiple conditions
	complexFilter := filter.And(
		filter.Eq("status", "active"),
		filter.Gte("age", 21),
		filter.In("role", "admin", "moderator"),
	)
	fmt.Printf("Complex filter: %+v\n", complexFilter.Build())

	// Array and text filters
	arrayFilter := filter.ElemMatch("tags", filter.Eq("category", "tech"))
	fmt.Printf("Array filter: %+v\n", arrayFilter.Build())

	textFilter := filter.Regex("name", "^John", "i")
	fmt.Printf("Text filter: %+v\n", textFilter.Build())

	fmt.Println("\n4. Demonstrating Fluent Update Builders...")

	// Simple update
	simpleUpdate := update.Set("last_login", "2024-01-01T00:00:00Z")
	fmt.Printf("Simple update: %+v\n", simpleUpdate.Build())

	// Complex update with multiple operations
	activityEntry := ActivityLog{
		Action:    "login",
		Timestamp: "2024-01-01T00:00:00Z",
	}
	complexUpdate := update.Set("status", "verified").
		And(update.Inc("login_count", 1)).
		And(update.Push("activity_log", activityEntry))
	fmt.Printf("Complex update: %+v\n", complexUpdate.Build())

	// Array operations
	arrayUpdate := update.PushEach("tags", "new", "trending", "popular")
	fmt.Printf("Array update: %+v\n", arrayUpdate.Build())

	fmt.Println("\n5. Demonstrating Query Combinations...")

	// Example: Find active users over 21 who are admins
	userFilter := filter.And(
		filter.Eq("status", "active"),
		filter.Gte("age", 21),
		filter.In("role", "admin", "moderator"),
	)

	// Example: Update their last seen and increment login count
	userUpdate := update.Set("last_seen", "2024-01-01T00:00:00Z").
		And(update.Inc("login_count", 1))

	fmt.Printf("Query filter: %+v\n", userFilter.Build())
	fmt.Printf("Update operation: %+v\n", userUpdate.Build())

	fmt.Println("\n6. Advanced Filter Examples...")

	// Existence and type checking
	existsFilter := filter.And(
		filter.Exists("email", true),
		filter.Type("age", filter.BSONTypeInt32),
	)
	fmt.Printf("Exists filter: %+v\n", existsFilter.Build())

	// Range queries
	rangeFilter := filter.And(
		filter.Gte("created_at", "2024-01-01"),
		filter.Lt("created_at", "2024-02-01"),
	)
	fmt.Printf("Range filter: %+v\n", rangeFilter.Build())

	// Array size and element matching
	arrayAdvancedFilter := filter.And(
		filter.Size("tags", 3),
		filter.All("categories", "tech", "programming"),
	)
	fmt.Printf("Advanced array filter: %+v\n", arrayAdvancedFilter.Build())
}
