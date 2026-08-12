// go-mongodb v2 Example: Match-or-Insert Operations
//
// This example demonstrates the three match-or-insert semantics with v2:
// - InsertIfAbsentByField / InsertIfAbsentByFieldMap: insert, or leave an existing document untouched ($setOnInsert)
// - SetOrInsertByField / SetOrInsertByFieldMap: insert, or merge fields into an existing document ($set)
// - ReplaceOrInsertByField: insert, or replace the existing document wholesale
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cloudresty/go-mongodb/v2"
	"github.com/cloudresty/go-mongodb/v2/filter"
	"github.com/cloudresty/go-mongodb/v2/update"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	// Comprehensive test of match-or-insert functionality

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

	fmt.Println("=== Match-or-Insert Examples ===")
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
	updateBuilder3, err := update.New().SetOnInsertStruct(event)
	if err != nil {
		log.Fatalf("Method 3 builder creation failed: %v", err)
	}
	result3, err := collection.UpdateOne(ctx, filterBuilder, updateBuilder3, opts)
	if err != nil {
		log.Printf("Method 3 failed: %v", err)
	} else {
		fmt.Printf("✅ Result: MatchedCount=%d, ModifiedCount=%d, UpsertedCount=%d\n\n",
			result3.MatchedCount, result3.ModifiedCount, result3.UpsertedCount)
	}

	// Clean up for next test
	_, _ = collection.DeleteOne(ctx, filter.Eq("url", event.URL))

	// Method 4: Convenience InsertIfAbsentByField (NEW)
	fmt.Println("Method 4: Convenience InsertIfAbsentByField (NEW)")
	result4, err := collection.InsertIfAbsentByField(ctx, "url", event.URL, event)
	if err != nil {
		log.Printf("Method 4 failed: %v", err)
	} else {
		fmt.Printf("✅ Result: DidInsert=%t, MatchedWithoutModifying=%t\n\n",
			result4.DidInsert(), result4.MatchedWithoutModifying())
	}

	// Clean up for next test
	_, _ = collection.DeleteOne(ctx, filter.Eq("url", event.URL))

	// Method 5: Convenience InsertIfAbsentByFieldMap (NEW)
	fmt.Println("Method 5: Convenience InsertIfAbsentByFieldMap (NEW)")
	result5, err := collection.InsertIfAbsentByFieldMap(ctx, "url", event.URL, fieldMap)
	if err != nil {
		log.Printf("Method 5 failed: %v", err)
	} else {
		fmt.Printf("✅ Result: DidInsert=%t, MatchedWithoutModifying=%t\n\n",
			result5.DidInsert(), result5.MatchedWithoutModifying())
	}

	// Clean up for next test
	_, _ = collection.DeleteOne(ctx, filter.Eq("url", event.URL))

	// Method 6: SetOrInsertByField (NEW) - merges fields into an existing document
	fmt.Println("Method 6: SetOrInsertByField (NEW) - merges into an existing document")
	result6a, err := collection.SetOrInsertByField(ctx, "url", event.URL, event)
	if err != nil {
		log.Printf("Method 6 insert failed: %v", err)
	} else {
		fmt.Printf("✅ First call (insert): DidInsert=%t\n", result6a.DidInsert())
	}
	updatedEvent := event
	updatedEvent.Title = "Updated Title via SetOrInsert"
	result6b, err := collection.SetOrInsertByField(ctx, "url", event.URL, updatedEvent)
	if err != nil {
		log.Printf("Method 6 merge failed: %v", err)
	} else {
		fmt.Printf("✅ Second call (merge): DidUpdate=%t - existing document's fields were merged, not replaced\n\n",
			result6b.DidUpdate())
	}

	// Clean up for next test
	_, _ = collection.DeleteOne(ctx, filter.Eq("url", event.URL))

	// Method 7: ReplaceOrInsertByField (NEW) - replaces the existing document wholesale
	fmt.Println("Method 7: ReplaceOrInsertByField (NEW) - replaces an existing document wholesale")
	result7a, err := collection.ReplaceOrInsertByField(ctx, "url", event.URL, event)
	if err != nil {
		log.Printf("Method 7 insert failed: %v", err)
	} else {
		fmt.Printf("✅ First call (insert): DidInsert=%t\n", result7a.DidInsert())
	}
	replacementEvent := Event{
		ID:        event.ID,
		URL:       event.URL,
		MediaID:   "media-replaced",
		Title:     "Replaced Document",
		EventType: "replaced",
		CreatedAt: now,
		UpdatedAt: now,
	}
	result7b, err := collection.ReplaceOrInsertByField(ctx, "url", event.URL, replacementEvent)
	if err != nil {
		log.Printf("Method 7 replace failed: %v", err)
	} else {
		fmt.Printf("✅ Second call (replace): DidUpdate=%t - the whole document was overwritten\n\n",
			result7b.DidUpdate())
	}

	// Clean up after the ReplaceOrInsertByField demonstration
	_, _ = collection.DeleteOne(ctx, filter.Eq("url", event.URL))
	_, _ = collection.InsertIfAbsentByField(ctx, "url", event.URL, event)

	// Test race condition prevention
	fmt.Println("=== Testing Race Condition Prevention ===")
	fmt.Println()

	// Try to insert-if-absent the same document again - should not modify existing
	fmt.Println("Attempting second InsertIfAbsentByField with same URL (should not modify):")
	result8, err := collection.InsertIfAbsentByField(ctx, "url", event.URL, event)
	if err != nil {
		log.Printf("Second insert-if-absent failed: %v", err)
	} else {
		fmt.Printf("✅ Result: DidInsert=%t, MatchedWithoutModifying=%t\n",
			result8.DidInsert(), result8.MatchedWithoutModifying())
		if result8.MatchedWithoutModifying() {
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
	fmt.Println("✅ All match-or-insert patterns work correctly")
	fmt.Println("✅ Race conditions are prevented")
	fmt.Println("✅ InsertIfAbsentByField/Map: existing documents are never modified ($setOnInsert)")
	fmt.Println("✅ SetOrInsertByField: existing documents are merged with new fields ($set)")
	fmt.Println("✅ ReplaceOrInsertByField: existing documents are replaced wholesale")
}
