package main

import (
	"fmt"
	"time"

	"github.com/cloudresty/go-mongodb"
	"github.com/cloudresty/go-mongodb/update"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// User represents a user document with custom timestamp fields
type User struct {
	ID          string    `bson:"_id,omitempty"`
	Name        string    `bson:"name"`
	Email       string    `bson:"email"`
	CreatedAt   time.Time `bson:"createdAt"`   // Custom timestamp field name
	CompletedAt time.Time `bson:"completedAt"` // Custom timestamp field name
}

func main() {
	fmt.Println("=== MongoDB Go Library - Timestamp Management Removal Demo ===")
	fmt.Println()

	// This demo shows that automatic timestamp management has been removed
	// Applications now have full control over their timestamp fields

	// Example 1: Manual timestamp management in InsertOne
	fmt.Println("1. InsertOne - Manual timestamp management:")
	user := User{
		Name:        "John Doe",
		Email:       "john@example.com",
		CreatedAt:   time.Now(),
		CompletedAt: time.Now(),
	}

	// Convert to bson.M to show what would be inserted
	userBytes, _ := bson.Marshal(user)
	var userDoc bson.M
	bson.Unmarshal(userBytes, &userDoc)

	fmt.Printf("Document to insert: %+v\n", userDoc)
	fmt.Println("✓ No automatic created_at/updated_at fields added")
	fmt.Println("✓ Application controls timestamp field names and values")
	fmt.Println()

	// Example 2: Manual timestamp management in UpdateOne
	fmt.Println("2. UpdateOne - Manual timestamp management:")
	updateBuilder := update.New().
		Set("name", "Jane Doe").
		Set("completedAt", time.Now()) // Application manually sets timestamp

	updateDoc := updateBuilder.Build()
	fmt.Printf("Update document: %+v\n", updateDoc)
	fmt.Println("✓ No automatic updated_at field added")
	fmt.Println("✓ Application controls when and how timestamps are updated")
	fmt.Println()

	// Example 3: Custom timestamp field names
	fmt.Println("3. Custom timestamp field names:")
	customUpdate := update.New().
		Set("lastModified", time.Now()).
		Set("processedAt", time.Now()).
		Set("syncedAt", time.Now())

	customDoc := customUpdate.Build()
	fmt.Printf("Custom timestamp update: %+v\n", customDoc)
	fmt.Println("✓ Applications can use any timestamp field names")
	fmt.Println("✓ Multiple timestamp fields with different purposes")
	fmt.Println()

	// Example 4: EnhanceDocument function behavior
	fmt.Println("4. EnhanceDocument function - Only adds ULID:")
	testDoc := bson.M{
		"title":  "Test Document",
		"status": "active",
	}

	enhanced := mongodb.EnhanceDocument(testDoc)
	fmt.Printf("Original: %+v\n", testDoc)
	fmt.Printf("Enhanced: %+v\n", enhanced)
	fmt.Println("✓ Only ULID is added, no automatic timestamps")
	fmt.Println()

	fmt.Println("=== Summary ===")
	fmt.Println("✓ Automatic created_at/updated_at injection removed from:")
	fmt.Println("  - InsertOne()")
	fmt.Println("  - InsertMany()")
	fmt.Println("  - UpdateOne()")
	fmt.Println("  - UpdateMany()")
	fmt.Println("  - ReplaceOne()")
	fmt.Println("  - EnhanceDocument()")
	fmt.Println()
	fmt.Println("✓ Applications now have full control over:")
	fmt.Println("  - Timestamp field names (createdAt, completedAt, etc.)")
	fmt.Println("  - Timestamp values and logic")
	fmt.Println("  - When timestamps are added or updated")
	fmt.Println()
	fmt.Println("✓ Removed helper functions:")
	fmt.Println("  - addUpdatedAt()")
	fmt.Println("  - enhanceReplacementDocument()")
	fmt.Println("  - Client.enhanceDocument()")
	fmt.Println()
	fmt.Println("✓ Removed UpsertOptions.SkipTimestamps field")
	fmt.Println()
	fmt.Println("Migration Guide:")
	fmt.Println("- Add timestamp fields manually in your application code")
	fmt.Println("- Use update.Set() to explicitly set timestamp fields")
	fmt.Println("- Define custom timestamp field names in your structs")
}

// Example helper function for applications to manage timestamps
func addTimestamps(doc bson.M) bson.M {
	now := time.Now()
	if _, exists := doc["createdAt"]; !exists {
		doc["createdAt"] = now
	}
	doc["updatedAt"] = now
	return doc
}

// Example update builder with custom timestamps
func updateWithTimestamp(updates *update.Builder) *update.Builder {
	return updates.Set("updatedAt", time.Now())
}
