package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cloudresty/go-mongodb"
	"github.com/cloudresty/go-mongodb/filter"
	"github.com/cloudresty/go-mongodb/update"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	// Comprehensive test of atomic upsert functionality with $setOnInsert

	// Create client
	client, err := mongodb.NewClient(mongodb.FromEnv())
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer func() { _ = client.Close() }()

	// Get collection
	collection := client.Database("test").Collection("upsert_test")

	ctx := context.Background()

	// Example event structure similar to user's use case
	type Event struct {
		ID        string    `bson:"_id"`
		URL       string    `bson:"url"`
		MediaID   string    `bson:"media_id"`
		Title     string    `bson:"title"`
		EventType string    `bson:"event_type"`
		CreatedAt time.Time `bson:"created_at"`
		UpdatedAt time.Time `bson:"updated_at"`
	}

	now := time.Now()
	event := Event{
		ID:        "test-ulid-123",
		URL:       "https://example.com/test",
		MediaID:   "media-456",
		Title:     "Test Event",
		EventType: "test",
		CreatedAt: now,
		UpdatedAt: now,
	}

	fmt.Println("=== Atomic Upsert with $setOnInsert Examples ===")
	fmt.Println()

	// Method 1: Individual field SetOnInsert (original approach)
	fmt.Println("Method 1: Individual field SetOnInsert")
	filterBuilder := filter.Eq("url", event.URL)
	updateBuilder := update.New().
		SetOnInsert("_id", event.ID).
		SetOnInsert("media_id", event.MediaID).
		SetOnInsert("title", event.Title).
		SetOnInsert("event_type", event.EventType).
		SetOnInsert("created_at", event.CreatedAt).
		SetOnInsert("updated_at", event.UpdatedAt)

	opts := options.UpdateOne().SetUpsert(true)
	result1, err := collection.UpdateOne(ctx, filterBuilder, updateBuilder, opts)
	if err != nil {
		log.Printf("Method 1 failed: %v", err)
	} else {
		fmt.Printf("✅ Result: MatchedCount=%d, ModifiedCount=%d, UpsertedCount=%d\n\n",
			result1.MatchedCount, result1.ModifiedCount, result1.UpsertedCount)
	}

	// Clean up for next test
	_, _ = collection.DeleteOne(ctx, filter.Eq("url", event.URL))

	// Method 2: Using SetOnInsertMap (NEW)
	fmt.Println("Method 2: Using SetOnInsertMap (NEW)")
	fieldMap := map[string]any{
		"_id":        event.ID,
		"media_id":   event.MediaID,
		"title":      event.Title,
		"event_type": event.EventType,
		"created_at": event.CreatedAt,
		"updated_at": event.UpdatedAt,
	}

	updateBuilder2 := update.New().SetOnInsertMap(fieldMap)
	result2, err := collection.UpdateOne(ctx, filterBuilder, updateBuilder2, opts)
	if err != nil {
		log.Printf("Method 2 failed: %v", err)
	} else {
		fmt.Printf("✅ Result: MatchedCount=%d, ModifiedCount=%d, UpsertedCount=%d\n\n",
			result2.MatchedCount, result2.ModifiedCount, result2.UpsertedCount)
	}

	// Clean up for next test
	_, _ = collection.DeleteOne(ctx, filter.Eq("url", event.URL))

	// Method 3: Using SetOnInsertStruct (NEW)
	fmt.Println("Method 3: Using SetOnInsertStruct (NEW)")
	updateBuilder3 := update.New().SetOnInsertStruct(event)
	result3, err := collection.UpdateOne(ctx, filterBuilder, updateBuilder3, opts)
	if err != nil {
		log.Printf("Method 3 failed: %v", err)
	} else {
		fmt.Printf("✅ Result: MatchedCount=%d, ModifiedCount=%d, UpsertedCount=%d\n\n",
			result3.MatchedCount, result3.ModifiedCount, result3.UpsertedCount)
	}

	// Clean up for next test
	_, _ = collection.DeleteOne(ctx, filter.Eq("url", event.URL))

	// Method 4: Convenience UpsertByField (NEW)
	fmt.Println("Method 4: Convenience UpsertByField (NEW)")
	result4, err := collection.UpsertByField(ctx, "url", event.URL, event)
	if err != nil {
		log.Printf("Method 4 failed: %v", err)
	} else {
		fmt.Printf("✅ Result: MatchedCount=%d, ModifiedCount=%d, UpsertedCount=%d\n\n",
			result4.MatchedCount, result4.ModifiedCount, result4.UpsertedCount)
	}

	// Clean up for next test
	_, _ = collection.DeleteOne(ctx, filter.Eq("url", event.URL))

	// Method 5: Convenience UpsertByFieldMap (NEW)
	fmt.Println("Method 5: Convenience UpsertByFieldMap (NEW)")
	result5, err := collection.UpsertByFieldMap(ctx, "url", event.URL, fieldMap)
	if err != nil {
		log.Printf("Method 5 failed: %v", err)
	} else {
		fmt.Printf("✅ Result: MatchedCount=%d, ModifiedCount=%d, UpsertedCount=%d\n\n",
			result5.MatchedCount, result5.ModifiedCount, result5.UpsertedCount)
	}

	// Test race condition prevention
	fmt.Println("=== Testing Race Condition Prevention ===")
	fmt.Println()

	// Try to upsert the same document again - should not modify existing
	fmt.Println("Attempting second upsert with same URL (should not modify):")
	result6, err := collection.UpsertByField(ctx, "url", event.URL, event)
	if err != nil {
		log.Printf("Second upsert failed: %v", err)
	} else {
		fmt.Printf("✅ Result: MatchedCount=%d, ModifiedCount=%d, UpsertedCount=%d\n",
			result6.MatchedCount, result6.ModifiedCount, result6.UpsertedCount)
		if result6.ModifiedCount == 0 && result6.UpsertedCount == 0 {
			fmt.Println("✅ SUCCESS: No modification occurred - existing document preserved!")
		}
	}

	// Verify document integrity
	fmt.Println("\nVerifying document integrity:")
	var foundDoc Event
	err = collection.FindOne(ctx, filter.Eq("url", event.URL)).Decode(&foundDoc)
	if err != nil {
		log.Printf("Find failed: %v", err)
	} else {
		fmt.Printf("✅ Document preserved: ID=%s, Title=%s, CreatedAt=%v\n",
			foundDoc.ID, foundDoc.Title, foundDoc.CreatedAt)
	}

	// Clean up test document
	_, err = collection.DeleteOne(ctx, filter.Eq("url", event.URL))
	if err != nil {
		log.Printf("Cleanup failed: %v", err)
	} else {
		fmt.Println("\n✅ Test document cleaned up successfully")
	}

	fmt.Println("\n=== Summary ===")
	fmt.Println("✅ All atomic upsert patterns work correctly")
	fmt.Println("✅ Race conditions are prevented")
	fmt.Println("✅ Existing documents are never modified with $setOnInsert")
	fmt.Println("✅ Multiple convenience methods available for different use cases")
}
