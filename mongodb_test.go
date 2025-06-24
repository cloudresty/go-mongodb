package mongodb

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestClientCreation(t *testing.T) {
	// Skip if no MongoDB available
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := &Config{
		Host:           "localhost",
		Port:           27017,
		Database:       "test",
		ConnectTimeout: 5 * time.Second,
	}

	client, err := NewClientWithConfig(config)
	if err != nil {
		t.Skipf("Could not connect to MongoDB: %v", err)
	}
	defer func() {
		_ = client.Close() // Ignore error during cleanup
	}()

	if !client.IsConnected() {
		t.Error("Expected client to be connected")
	}
}

func TestDatabaseCreation(t *testing.T) {
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

	db := client.Database("test_database")
	if db == nil {
		t.Error("Expected database to be created")
	}
}

func TestCollectionOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := NewClient()
	if err != nil {
		t.Skipf("Could not create client: %v", err)
	}
	defer func() {
		// Add a small delay to let any in-flight logging complete
		// This helps avoid race conditions in the emit library's timestamp code
		time.Sleep(10 * time.Millisecond)
		_ = client.Close() // Ignore error during cleanup
	}()

	// Setup test collection
	collection := client.Collection("test_collection")

	// Test document
	testDoc := bson.M{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
	}

	// Test Insert
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := collection.InsertOne(ctx, testDoc)
	if err != nil {
		t.Fatalf("Failed to insert document: %v", err)
	}

	if result.InsertedID.IsZero() {
		t.Error("Expected inserted ID to be set")
	}

	// Test Find
	var foundDoc bson.M
	err = collection.FindOne(ctx, bson.M{"name": "John Doe"}).Decode(&foundDoc)
	if err != nil {
		t.Fatalf("Failed to find document: %v", err)
	}

	if foundDoc["name"] != "John Doe" {
		t.Errorf("Expected name 'John Doe', got %v", foundDoc["name"])
	}

	// Test Update
	updateResult, err := collection.UpdateOne(ctx, bson.M{"name": "John Doe"}, bson.M{"$set": bson.M{"age": 31}})
	if err != nil {
		t.Fatalf("Failed to update document: %v", err)
	}

	if updateResult.ModifiedCount != 1 {
		t.Errorf("Expected 1 document to be modified, got %d", updateResult.ModifiedCount)
	}

	// Test Delete
	deleteResult, err := collection.DeleteOne(ctx, bson.M{"name": "John Doe"})
	if err != nil {
		t.Fatalf("Failed to delete document: %v", err)
	}

	if deleteResult.DeletedCount != 1 {
		t.Errorf("Expected 1 document to be deleted, got %d", deleteResult.DeletedCount)
	}
}

func TestIndexOperations(t *testing.T) {
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

	// Setup test collection
	collection := client.Collection("test_index_collection")

	ctx := context.Background()

	// Create a simple index
	indexModel := IndexModel{
		Keys: bson.D{bson.E{Key: "email", Value: 1}},
	}
	_, err = collection.CreateIndex(ctx, indexModel)
	if err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}

	// List indexes to verify creation
	cursor, err := collection.ListIndexes(ctx)
	if err != nil {
		t.Fatalf("Failed to list indexes: %v", err)
	}
	defer cursor.Close(ctx)

	var indexes []bson.M
	if err = cursor.All(ctx, &indexes); err != nil {
		t.Fatalf("Failed to decode indexes: %v", err)
	}

	// Should have at least 2 indexes: _id (default) and email
	if len(indexes) < 2 {
		t.Errorf("Expected at least 2 indexes, got %d", len(indexes))
	}
}

func TestHealthCheck(t *testing.T) {
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

	// Test health check via IsConnected
	connected := client.IsConnected()
	if !connected {
		t.Error("Expected client to be connected")
	}

	// Test new HealthCheck method
	health := client.HealthCheck()
	if health == nil {
		t.Error("Expected health status to be returned")
	}

	if !health.IsHealthy {
		t.Errorf("Expected client to be healthy, but got error: %s", health.Error)
	}

	if health.Latency <= 0 {
		t.Error("Expected positive latency value")
	}

	if health.CheckedAt.IsZero() {
		t.Error("Expected CheckedAt to be set")
	}
}

func TestTransactionOperations(t *testing.T) {
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

	// Start a session
	session, err := client.StartSession()
	if err != nil {
		t.Fatalf("Failed to start session: %v", err)
	}
	defer session.EndSession(context.Background())

	// Test transaction callback
	ctx := context.Background()
	_, err = client.WithTransaction(ctx, func(ctx context.Context) (any, error) {
		// Perform operations within transaction
		collection := client.Collection("test_transaction_collection")

		doc := bson.M{"test": "transaction", "timestamp": time.Now()}
		result, err := collection.InsertOne(ctx, doc)
		return result, err
	})

	if err != nil {
		t.Fatalf("Transaction failed: %v", err)
	}
}

func TestBulkOperations(t *testing.T) {
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

	// Setup test collection
	db := client.Database("test_db")
	collection := db.Collection("test_bulk_collection")

	ctx := context.Background()

	// Test bulk insert
	docs := []any{
		bson.M{"name": "User1", "type": "bulk"},
		bson.M{"name": "User2", "type": "bulk"},
		bson.M{"name": "User3", "type": "bulk"},
	}

	result, err := collection.InsertMany(ctx, docs)
	if err != nil {
		t.Fatalf("Failed to bulk insert documents: %v", err)
	}

	if len(result.InsertedIDs) != 3 {
		t.Errorf("Expected 3 documents to be inserted, got %d", len(result.InsertedIDs))
	}

	// Cleanup
	_, err = collection.DeleteMany(ctx, bson.M{"type": "bulk"})
	if err != nil {
		t.Logf("Failed to cleanup bulk test documents: %v", err)
	}
}

func TestShutdownManagerNewInterface(t *testing.T) {
	// Test new shutdown manager with ShutdownConfig
	config := &ShutdownConfig{
		Timeout:          10 * time.Second,
		GracePeriod:      2 * time.Second,
		ForceKillTimeout: 5 * time.Second,
	}

	shutdownManager := NewShutdownManager(config)
	if shutdownManager == nil {
		t.Error("Expected shutdown manager to be created")
	}

	// Test Register method
	client, err := NewClient()
	if err != nil {
		t.Skipf("Could not create client: %v", err)
	}

	shutdownManager.Register(client)

	// Test context method
	ctx := shutdownManager.Context()
	if ctx == nil {
		t.Error("Expected context to be returned")
	}

	// Test timeout getter
	timeout := shutdownManager.GetTimeout()
	if timeout != config.Timeout {
		t.Errorf("Expected timeout %v, got %v", config.Timeout, timeout)
	}

	// Test client count
	count := shutdownManager.GetClientCount()
	if count != 1 {
		t.Errorf("Expected 1 client registered, got %d", count)
	}

	// Clean up
	_ = client.Close()
}
