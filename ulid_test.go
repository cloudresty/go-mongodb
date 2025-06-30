package mongodb

import (
	"context"
	"testing"
	"time"

	"github.com/cloudresty/go-mongodb/filter"
	"github.com/cloudresty/ulid"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// TestULIDDocumentGeneration tests that documents are enhanced with ULID
func TestULIDDocumentGeneration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := NewClient(FromEnv())
	if err != nil {
		t.Skipf("Could not create client: %v", err)
	}
	defer func() {
		_ = client.Close() // Ignore error during cleanup
	}()

	collection := client.Collection("test_ulid_generation")
	ctx := context.Background()

	// Cleanup collection before test
	cleanupTestCollection(t, client, "test_ulid_generation")

	// Insert a document and verify ULID is generated
	testDoc := bson.M{"name": "Test User", "type": "ulid_test"}
	result, err := collection.InsertOne(ctx, testDoc)
	if err != nil {
		t.Fatalf("Failed to insert document: %v", err)
	}

	// Verify the document was enhanced with ULID
	var insertedDoc bson.M
	err = collection.FindOne(ctx, filter.Eq("_id", result.InsertedID)).Decode(&insertedDoc)
	if err != nil {
		t.Fatalf("Failed to find inserted document: %v", err)
	}

	// Check if _id field exists and is a ULID
	ulidValue, exists := insertedDoc["_id"]
	if !exists {
		t.Error("Expected _id field to be present in document")
	}

	// Verify it's a valid ULID string
	ulidStr, ok := ulidValue.(string)
	if !ok {
		t.Errorf("Expected _id to be string ULID, got %T", ulidValue)
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
	_, err = collection.DeleteOne(ctx, filter.Eq("_id", result.InsertedID))
	if err != nil {
		t.Logf("Failed to cleanup test document: %v", err)
	}
}

// TestULIDUniqueness tests that multiple ULIDs are unique
func TestULIDUniqueness(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := NewClient(FromEnv())
	if err != nil {
		t.Skipf("Could not create client: %v", err)
	}
	defer func() {
		_ = client.Close() // Ignore error during cleanup
	}()

	collection := client.Collection("test_ulid_uniqueness")
	ctx := context.Background()

	// Cleanup collection before test
	cleanupTestCollection(t, client, "test_ulid_uniqueness")

	documentULIDs := make(map[string]bool)
	var insertedIDs []string

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
		err = collection.FindOne(ctx, filter.Eq("_id", result.InsertedID)).Decode(&doc)
		if err != nil {
			t.Fatalf("Failed to find document %d: %v", i, err)
		}

		ulidStr := doc["_id"].(string)

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
		_, err = collection.DeleteOne(ctx, filter.Eq("_id", id))
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

	client, err := NewClient(FromEnv())
	if err != nil {
		t.Skipf("Could not create client: %v", err)
	}
	defer func() {
		_ = client.Close() // Ignore error during cleanup
	}()

	collection := client.Collection("test_ulid_ordering")
	ctx := context.Background()

	// Cleanup collection before test
	cleanupTestCollection(t, client, "test_ulid_ordering")

	const numDocuments = 5
	ulids := make([]string, numDocuments)
	var insertedIDs []string

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
		err = collection.FindOne(ctx, filter.Eq("_id", result.InsertedID)).Decode(&doc)
		if err != nil {
			t.Fatalf("Failed to find document %d: %v", i, err)
		}

		ulids[i] = doc["_id"].(string)

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
		_, err = collection.DeleteOne(ctx, filter.Eq("_id", id))
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

	client, err := NewClient(FromEnv())
	if err != nil {
		t.Skipf("Could not create client: %v", err)
	}
	defer func() {
		_ = client.Close() // Ignore error during cleanup
	}()

	collection := client.Collection("test_find_by_ulid")
	ctx := context.Background()

	// Cleanup collection before test
	cleanupTestCollection(t, client, "test_find_by_ulid")

	// Insert a document
	testDoc := bson.M{"name": "Find By ULID Test", "type": "find_test"}
	result, err := collection.InsertOne(ctx, testDoc)
	if err != nil {
		t.Fatalf("Failed to insert document: %v", err)
	}

	// Get the ULID from the inserted document
	var insertedDoc bson.M
	err = collection.FindOne(ctx, filter.Eq("_id", result.InsertedID)).Decode(&insertedDoc)
	if err != nil {
		t.Fatalf("Failed to find inserted document: %v", err)
	}

	testULID := insertedDoc["_id"].(string)

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

	if foundDoc["_id"] != testULID {
		t.Errorf("Expected ULID %s, got %s", testULID, foundDoc["_id"])
	}

	// Cleanup
	_, err = collection.DeleteOne(ctx, filter.Eq("_id", result.InsertedID))
	if err != nil {
		t.Logf("Failed to cleanup test document: %v", err)
	}
}

// TestULIDDocumentFormatting tests that ULID formatting works properly
func TestULIDDocumentFormatting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := NewClient(FromEnv())
	if err != nil {
		t.Skipf("Could not create client: %v", err)
	}
	defer func() {
		_ = client.Close() // Ignore error during cleanup
	}()

	collection := client.Collection("test_ulid_operations")
	ctx := context.Background()

	// Cleanup collection before test
	cleanupTestCollection(t, client, "test_ulid_operations")

	// Insert a document
	testDoc := bson.M{"name": "ULID Test", "type": "ulid_test"}
	result, err := collection.InsertOne(ctx, testDoc)
	if err != nil {
		t.Fatalf("Failed to insert document: %v", err)
	}

	// Verify ULID is properly formatted
	ulidID := result.InsertedID
	if ulidID == "" {
		t.Error("Expected non-empty ULID")
	}

	// ULID should be 26 characters
	if len(ulidID) != 26 {
		t.Errorf("Expected ULID length 26, got %d", len(ulidID))
	}

	t.Logf("Generated ULID: %s", ulidID)

	// Cleanup
	_, err = collection.DeleteOne(ctx, filter.Eq("_id", result.InsertedID))
	if err != nil {
		t.Logf("Failed to cleanup test document: %v", err)
	}
}

// cleanupTestCollection removes all documents from a test collection
func cleanupTestCollection(t *testing.T, client *Client, collectionName string) {
	collection := client.Collection(collectionName)
	ctx := context.Background()

	_, err := collection.DeleteMany(ctx, filter.New())
	if err != nil {
		t.Logf("Warning: Failed to cleanup collection %s: %v", collectionName, err)
	}
}
