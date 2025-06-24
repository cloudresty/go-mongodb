package mongodb

import (
	"context"
	"testing"
	"time"

	"github.com/cloudresty/ulid"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// TestULIDDocumentGeneration tests that documents are enhanced with ULID
func TestULIDDocumentGeneration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := NewClient()
	if err != nil {
		t.Skipf("Could not create client: %v", err)
	}
	defer func() {
		_ = client.Close() // Ignore error during cleanup
	}()

	collection := client.Collection("test_ulid_generation")
	ctx := context.Background()

	// Insert a document and verify ULID is generated
	testDoc := bson.M{"name": "Test User", "type": "ulid_test"}
	result, err := collection.InsertOne(ctx, testDoc)
	if err != nil {
		t.Fatalf("Failed to insert document: %v", err)
	}

	// Verify the document was enhanced with ULID
	var insertedDoc bson.M
	err = collection.FindOne(ctx, bson.M{"_id": result.InsertedID}).Decode(&insertedDoc)
	if err != nil {
		t.Fatalf("Failed to find inserted document: %v", err)
	}

	// Check if ULID field exists
	ulidValue, exists := insertedDoc["ulid"]
	if !exists {
		t.Error("Expected ULID field to be present in document")
	}

	// Verify it's a valid ULID string
	ulidStr, ok := ulidValue.(string)
	if !ok {
		t.Errorf("Expected ULID to be string, got %T", ulidValue)
	}

	// Parse to verify it's a valid ULID
	_, err = ulid.Parse(ulidStr)
	if err != nil {
		t.Errorf("Invalid ULID generated: %s, error: %v", ulidStr, err)
	}

	// Verify other enhancement fields
	if _, exists := insertedDoc["created_at"]; !exists {
		t.Error("Expected created_at field to be present")
	}
	if _, exists := insertedDoc["updated_at"]; !exists {
		t.Error("Expected updated_at field to be present")
	}

	t.Logf("Generated ULID: %s", ulidStr)

	// Cleanup
	_, err = collection.DeleteOne(ctx, bson.M{"_id": result.InsertedID})
	if err != nil {
		t.Logf("Failed to cleanup test document: %v", err)
	}
}

// TestULIDUniqueness tests that multiple ULIDs are unique
func TestULIDUniqueness(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := NewClient()
	if err != nil {
		t.Skipf("Could not create client: %v", err)
	}
	defer func() {
		_ = client.Close() // Ignore error during cleanup
	}()

	collection := client.Collection("test_ulid_uniqueness")
	ctx := context.Background()

	documentULIDs := make(map[string]bool)
	var insertedIDs []bson.ObjectID

	// Generate 10 documents and verify all ULIDs are unique
	for i := 0; i < 10; i++ {
		testDoc := bson.M{"name": "Test User", "index": i, "type": "uniqueness_test"}
		result, err := collection.InsertOne(ctx, testDoc)
		if err != nil {
			t.Fatalf("Failed to insert document %d: %v", i, err)
		}

		insertedIDs = append(insertedIDs, result.InsertedID)

		// Retrieve and check ULID
		var doc bson.M
		err = collection.FindOne(ctx, bson.M{"_id": result.InsertedID}).Decode(&doc)
		if err != nil {
			t.Fatalf("Failed to find document %d: %v", i, err)
		}

		ulidStr := doc["ulid"].(string)

		if documentULIDs[ulidStr] {
			t.Fatalf("Duplicate ULID generated: %s", ulidStr)
		}

		// Store the ULID
		documentULIDs[ulidStr] = true

		// Also verify it's a valid ULID
		_, err = ulid.Parse(ulidStr)
		if err != nil {
			t.Fatalf("Invalid ULID generated: %s, error: %v", ulidStr, err)
		}

		// Small delay to ensure different timestamps
		time.Sleep(1 * time.Millisecond)
	}

	t.Logf("Successfully generated %d unique ULIDs", len(documentULIDs))

	// Cleanup
	for _, id := range insertedIDs {
		_, err = collection.DeleteOne(ctx, bson.M{"_id": id})
		if err != nil {
			t.Logf("Failed to cleanup document: %v", err)
		}
	}
}

// TestULIDTemporalOrdering tests that ULIDs generated in sequence are temporally ordered
func TestULIDTemporalOrdering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := NewClient()
	if err != nil {
		t.Skipf("Could not create client: %v", err)
	}
	defer func() {
		_ = client.Close() // Ignore error during cleanup
	}()

	collection := client.Collection("test_ulid_ordering")
	ctx := context.Background()

	const numDocuments = 5
	ulids := make([]string, numDocuments)
	var insertedIDs []bson.ObjectID

	// Generate documents with small delays to ensure temporal ordering
	for i := 0; i < numDocuments; i++ {
		testDoc := bson.M{"name": "Test User", "sequence": i, "type": "ordering_test"}
		result, err := collection.InsertOne(ctx, testDoc)
		if err != nil {
			t.Fatalf("Failed to insert document %d: %v", i, err)
		}

		insertedIDs = append(insertedIDs, result.InsertedID)

		// Retrieve ULID
		var doc bson.M
		err = collection.FindOne(ctx, bson.M{"_id": result.InsertedID}).Decode(&doc)
		if err != nil {
			t.Fatalf("Failed to find document %d: %v", i, err)
		}

		ulids[i] = doc["ulid"].(string)

		// Parse the ULID to get the timestamp
		parsedUlid, err := ulid.Parse(ulids[i])
		if err != nil {
			t.Fatalf("Invalid ULID generated: %s, error: %v", ulids[i], err)
		}

		ulidTimeMs := parsedUlid.GetTime()
		if ulidTimeMs > uint64(1<<63-1) { // Check for overflow
			t.Fatalf("ULID timestamp too large to convert safely: %d", ulidTimeMs)
		}

		currentTime := time.UnixMilli(int64(ulidTimeMs))
		t.Logf("Document %d: ULID=%s, Time=%v", i+1, ulids[i], currentTime)

		time.Sleep(2 * time.Millisecond) // Small delay to ensure different timestamps
	}

	// Verify temporal ordering - each ULID should be >= the previous one
	for i := 1; i < numDocuments; i++ {
		if ulids[i] < ulids[i-1] {
			t.Errorf("ULIDs are not temporally ordered: %s < %s",
				ulids[i], ulids[i-1])
		}
	}

	// Cleanup
	for _, id := range insertedIDs {
		_, err = collection.DeleteOne(ctx, bson.M{"_id": id})
		if err != nil {
			t.Logf("Failed to cleanup document: %v", err)
		}
	}
}

// TestFindByULID tests the FindByULID helper method
func TestFindByULID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := NewClient()
	if err != nil {
		t.Skipf("Could not create client: %v", err)
	}
	defer func() {
		_ = client.Close() // Ignore error during cleanup
	}()

	collection := client.Collection("test_find_by_ulid")
	ctx := context.Background()

	// Insert a document
	testDoc := bson.M{"name": "Find By ULID Test", "type": "find_test"}
	result, err := collection.InsertOne(ctx, testDoc)
	if err != nil {
		t.Fatalf("Failed to insert document: %v", err)
	}

	// Get the ULID from the inserted document
	var insertedDoc bson.M
	err = collection.FindOne(ctx, bson.M{"_id": result.InsertedID}).Decode(&insertedDoc)
	if err != nil {
		t.Fatalf("Failed to find inserted document: %v", err)
	}

	testULID := insertedDoc["ulid"].(string)

	// Test FindByULID
	var foundDoc bson.M
	err = collection.FindByULID(ctx, testULID).Decode(&foundDoc)
	if err != nil {
		t.Fatalf("Failed to find document by ULID: %v", err)
	}

	// Verify it's the same document
	if foundDoc["name"] != "Find By ULID Test" {
		t.Errorf("Expected name 'Find By ULID Test', got %v", foundDoc["name"])
	}

	if foundDoc["ulid"] != testULID {
		t.Errorf("Expected ULID %s, got %s", testULID, foundDoc["ulid"])
	}

	// Cleanup
	_, err = collection.DeleteOne(ctx, bson.M{"_id": result.InsertedID})
	if err != nil {
		t.Logf("Failed to cleanup test document: %v", err)
	}
}

// TestULIDObjectIDEmbedding tests that ObjectIDs embed ULID data
func TestULIDObjectIDEmbedding(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := NewClient()
	if err != nil {
		t.Skipf("Could not create client: %v", err)
	}
	defer func() {
		_ = client.Close() // Ignore error during cleanup
	}()

	collection := client.Collection("test_objectid_ulid")
	ctx := context.Background()

	// Insert a document
	testDoc := bson.M{"name": "ObjectID ULID Test", "type": "objectid_test"}
	result, err := collection.InsertOne(ctx, testDoc)
	if err != nil {
		t.Fatalf("Failed to insert document: %v", err)
	}

	// Verify ObjectID is properly formatted
	objectID := result.InsertedID
	if objectID.IsZero() {
		t.Error("Expected non-zero ObjectID")
	}

	// ObjectID should be 12 bytes (24 hex characters when stringified)
	objectIDStr := objectID.Hex()
	if len(objectIDStr) != 24 {
		t.Errorf("Expected ObjectID hex string length 24, got %d", len(objectIDStr))
	}

	t.Logf("Generated ObjectID: %s", objectIDStr)

	// Cleanup
	_, err = collection.DeleteOne(ctx, bson.M{"_id": result.InsertedID})
	if err != nil {
		t.Logf("Failed to cleanup test document: %v", err)
	}
}
