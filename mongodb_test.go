package mongodb

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cloudresty/go-mongodb/filter"
	"github.com/cloudresty/go-mongodb/update"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func TestClientCreation(t *testing.T) {
	// Skip if no MongoDB available
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := NewClient(FromEnv())
	if err != nil {
		t.Skipf("Could not connect to MongoDB: %v", err)
	}
	defer func() {
		_ = client.Close() // Ignore error during cleanup
	}()

	// Verify connectivity using Ping instead of IsConnected
	if err := client.Ping(context.Background()); err != nil {
		t.Errorf("Expected client to be connected, ping failed: %v", err)
	}
}

func TestDatabaseCreation(t *testing.T) {
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

	db := client.Database("test")
	if db == nil {
		t.Error("Expected database to be created")
	}
}

func TestCollectionOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := NewClient(FromEnv())
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

	if result.InsertedID == "" {
		t.Error("Expected inserted ID to be set")
	}

	// Test Find
	var foundDoc bson.M
	err = collection.FindOne(ctx, filter.Eq("name", "John Doe")).Decode(&foundDoc)
	if err != nil {
		t.Fatalf("Failed to find document: %v", err)
	}

	if foundDoc["name"] != "John Doe" {
		t.Errorf("Expected name 'John Doe', got %v", foundDoc["name"])
	}

	// Test Update
	updateResult, err := collection.UpdateOne(ctx, filter.Eq("name", "John Doe"), update.Set("age", 31))
	if err != nil {
		t.Fatalf("Failed to update document: %v", err)
	}

	if updateResult.ModifiedCount != 1 {
		t.Errorf("Expected 1 document to be modified, got %d", updateResult.ModifiedCount)
	}

	// Test Delete
	deleteResult, err := collection.DeleteOne(ctx, filter.Eq("name", "John Doe"))
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

	client, err := NewClient(FromEnv())
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
	indexModel := mongo.IndexModel{
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
	defer func() { _ = cursor.Close(ctx) }()

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

	client, err := NewClient(FromEnv())
	if err != nil {
		t.Skipf("Could not create client: %v", err)
	}
	defer func() {
		_ = client.Close() // Ignore error during cleanup
	}()

	// Test health check via Ping
	if err := client.Ping(context.Background()); err != nil {
		t.Errorf("Expected client to be connected, ping failed: %v", err)
	}

	// Test new HealthCheck method
	health := client.HealthCheck()
	if health == nil {
		t.Fatal("Expected health status to be returned")
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

	client, err := NewClient(FromEnv())
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
		// Skip if transactions are not supported (e.g., standalone MongoDB)
		if strings.Contains(err.Error(), "Transaction numbers are only allowed on a replica set member or mongos") {
			t.Skip("Skipping transaction test: MongoDB is not running as a replica set")
		}
		t.Fatalf("Transaction failed: %v", err)
	}
}

func TestBulkOperations(t *testing.T) {
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

	// Setup test collection
	db := client.Database("test")
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
	_, err = collection.DeleteMany(ctx, filter.Eq("type", "bulk"))
	if err != nil {
		t.Logf("Failed to cleanup bulk test documents: %v", err)
	}
}

func TestShutdownManagerNewInterface(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

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
	client, err := NewClient(FromEnv())
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

// Direct Connection tests

func TestDirectConnectionConfiguration(t *testing.T) {
	tests := []struct {
		name          string
		config        *Config
		expectedInURI bool
		expectedValue bool
	}{
		{
			name: "Direct connection enabled",
			config: &Config{
				Hosts:            "localhost:27017",
				Database:         "test",
				DirectConnection: true,
			},
			expectedInURI: true,
			expectedValue: true,
		},
		{
			name: "Direct connection disabled (default)",
			config: &Config{
				Hosts:            "localhost:27017",
				Database:         "test",
				DirectConnection: false,
			},
			expectedInURI: false,
			expectedValue: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uri := tt.config.BuildConnectionURI()

			if tt.expectedInURI {
				if !contains(uri, "directConnection=true") {
					t.Errorf("Expected URI to contain 'directConnection=true', got: %s", uri)
				}
			} else {
				if contains(uri, "directConnection") {
					t.Errorf("Expected URI to not contain 'directConnection', got: %s", uri)
				}
			}

			// Verify config value
			if tt.config.DirectConnection != tt.expectedValue {
				t.Errorf("Expected DirectConnection to be %v, got %v", tt.expectedValue, tt.config.DirectConnection)
			}
		})
	}
}

func TestWithDirectConnectionOption(t *testing.T) {
	tests := []struct {
		name     string
		enabled  bool
		expected bool
	}{
		{
			name:     "Enable direct connection",
			enabled:  true,
			expected: true,
		},
		{
			name:     "Disable direct connection",
			enabled:  false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{}
			option := WithDirectConnection(tt.enabled)
			option(config)

			if config.DirectConnection != tt.expected {
				t.Errorf("Expected DirectConnection to be %v, got %v", tt.expected, config.DirectConnection)
			}
		})
	}
}

func TestDirectConnectionEnvironmentVariable(t *testing.T) {
	// Save original environment
	originalValue := os.Getenv("MONGODB_DIRECT_CONNECTION")
	defer func() {
		if originalValue == "" {
			_ = os.Unsetenv("MONGODB_DIRECT_CONNECTION")
		} else {
			_ = os.Setenv("MONGODB_DIRECT_CONNECTION", originalValue)
		}
	}()

	tests := []struct {
		name        string
		envValue    string
		expected    bool
		shouldError bool
	}{
		{
			name:     "DirectConnection enabled via env",
			envValue: "true",
			expected: true,
		},
		{
			name:     "DirectConnection disabled via env",
			envValue: "false",
			expected: false,
		},
		{
			name:     "DirectConnection not set (default)",
			envValue: "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envValue == "" {
				_ = os.Unsetenv("MONGODB_DIRECT_CONNECTION")
			} else {
				_ = os.Setenv("MONGODB_DIRECT_CONNECTION", tt.envValue)
			}

			// Load config from environment
			config, err := loadConfigFromEnv("")
			if tt.shouldError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Check DirectConnection value
			if config.DirectConnection != tt.expected {
				t.Errorf("Expected DirectConnection to be %v, got %v", tt.expected, config.DirectConnection)
			}

			// Check URI generation
			uri := config.BuildConnectionURI()
			if tt.expected {
				if !contains(uri, "directConnection=true") {
					t.Errorf("Expected URI to contain 'directConnection=true', got: %s", uri)
				}
			} else {
				if contains(uri, "directConnection") {
					t.Errorf("Expected URI to not contain 'directConnection', got: %s", uri)
				}
			}
		})
	}
}

// Error handling tests

func TestIsNotFoundError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "mongo.ErrNoDocuments",
			err:      mongo.ErrNoDocuments,
			expected: true,
		},
		{
			name:     "NamespaceNotFound command error",
			err:      mongo.CommandError{Code: 26, Message: "namespace not found"},
			expected: true,
		},
		{
			name:     "InvalidNamespace command error",
			err:      mongo.CommandError{Code: 73, Message: "invalid namespace"},
			expected: true,
		},
		{
			name:     "Other command error",
			err:      mongo.CommandError{Code: 11000, Message: "duplicate key"},
			expected: false,
		},
		{
			name:     "Generic not found error message",
			err:      &customError{msg: "document not found"},
			expected: true,
		},
		{
			name:     "Does not exist error message",
			err:      &customError{msg: "collection does not exist"},
			expected: true,
		},
		{
			name:     "No documents error message",
			err:      &customError{msg: "no documents found"},
			expected: true,
		},
		{
			name:     "Namespace not found error message",
			err:      &customError{msg: "namespace not found"},
			expected: true,
		},
		{
			name:     "Unrelated error",
			err:      &customError{msg: "connection timeout"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNotFoundError(tt.err)
			if result != tt.expected {
				t.Errorf("IsNotFoundError(%v) = %v, expected %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestIsNotFoundErrorIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := NewClient(FromEnv())
	if err != nil {
		t.Skipf("Could not create client: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()
	// Test actual not found error from FindOne
	collection := client.Database("testdb").Collection("nonexistent")

	var result bson.M
	err = collection.FindOne(context.Background(), filter.Eq("_id", "nonexistent")).Decode(&result)

	if err == nil {
		t.Error("Expected error when finding nonexistent document")
		return
	}

	if !IsNotFoundError(err) {
		t.Errorf("Expected IsNotFoundError to return true for FindOne error: %v", err)
	}

	// Verify it's specifically mongo.ErrNoDocuments
	if err != mongo.ErrNoDocuments {
		t.Errorf("Expected mongo.ErrNoDocuments, got: %v", err)
	}
}

// Helper type for testing custom error messages
type customError struct {
	msg string
}

func (e *customError) Error() string {
	return e.msg
}

func TestDirectConnectionWithExistingAuth(t *testing.T) {
	tests := []struct {
		name                          string
		hosts                         string
		username                      string
		password                      string
		database                      string
		authDatabase                  string
		directConn                    bool
		shouldContainDirectConnection bool
	}{
		{
			name:                          "Single host with auth and direct connection enabled",
			hosts:                         "localhost:27017",
			username:                      "admin",
			password:                      "password",
			database:                      "media_agenda",
			authDatabase:                  "admin",
			directConn:                    true,
			shouldContainDirectConnection: true,
		},
		{
			name:                          "Multiple hosts with auth and direct connection enabled (should not add directConnection)",
			hosts:                         "localhost:27017,localhost:27018",
			username:                      "admin",
			password:                      "password",
			database:                      "media_agenda",
			authDatabase:                  "admin",
			directConn:                    true,
			shouldContainDirectConnection: false,
		},
		{
			name:                          "Single host with auth and direct connection disabled",
			hosts:                         "localhost:27017",
			username:                      "admin",
			password:                      "password",
			database:                      "media_agenda",
			authDatabase:                  "admin",
			directConn:                    false,
			shouldContainDirectConnection: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Hosts:            tt.hosts,
				Username:         tt.username,
				Password:         tt.password,
				Database:         tt.database,
				AuthDatabase:     tt.authDatabase,
				DirectConnection: tt.directConn,
			}

			actualURI := config.BuildConnectionURI()

			// Check if directConnection parameter is present when expected
			containsDirectConnection := contains(actualURI, "directConnection=true")

			if containsDirectConnection != tt.shouldContainDirectConnection {
				t.Errorf("Expected directConnection=true in URI: %v, got: %v\nURI: %s",
					tt.shouldContainDirectConnection, containsDirectConnection, actualURI)
			}

			// Should always contain auth parameters when credentials are provided
			if !contains(actualURI, "authSource="+tt.authDatabase) {
				t.Errorf("Expected URI to contain authSource=%s, got: %s", tt.authDatabase, actualURI)
			}
		})
	}
}

func TestDirectConnectionEnvironmentVariableWithAuth(t *testing.T) {
	// Test with individual environment variables instead of full URI
	t.Setenv("MONGODB_HOSTS", "localhost:27017")
	t.Setenv("MONGODB_USERNAME", "admin")
	t.Setenv("MONGODB_PASSWORD", "password")
	t.Setenv("MONGODB_DATABASE", "media_agenda")
	t.Setenv("MONGODB_AUTH_DATABASE", "admin")
	t.Setenv("MONGODB_DIRECT_CONNECTION", "true")

	config, err := loadConfigFromEnv("")
	if err != nil {
		t.Fatalf("Failed to load config from env: %v", err)
	}

	actualURI := config.BuildConnectionURI()

	// Check that directConnection=true is present for single host
	if !contains(actualURI, "directConnection=true") {
		t.Errorf("Expected URI to contain 'directConnection=true', got: %s", actualURI)
	}

	// Check that auth parameters are present
	if !contains(actualURI, "authSource=admin") {
		t.Errorf("Expected URI to contain 'authSource=admin', got: %s", actualURI)
	}

	// Check that credentials are present
	if !contains(actualURI, "admin:password@") {
		t.Errorf("Expected URI to contain credentials, got: %s", actualURI)
	}
}
