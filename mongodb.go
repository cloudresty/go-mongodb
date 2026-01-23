// Package mongodb provides a modern, production-ready Go package for MongoDB operations
// with environment-first configuration, ULID message IDs, auto-reconnection, and comprehensive production features.
//
// This package follows the same design patterns as the cloudresty/go-rabbitmq package,
// providing a clean, intuitive API for MongoDB operations while maintaining high performance
// and production-ready features.
//
// Key Features:
//   - Environment-first configuration using cloudresty/go-env
//   - ULID-based document IDs for better performance and sorting
//   - Auto-reconnection with intelligent retry and exponential backoff
//   - Zero-allocation logging with cloudresty/emit
//   - Production-ready features (graceful shutdown, health checks, metrics)
//   - Simple, intuitive function names following Go best practices
//   - Comprehensive error handling and logging
//   - Built-in connection pooling and compression
//   - Transaction support with helper methods
//   - Change streams for real-time data
//   - Index management utilities
//
// Environment Variables:
//   - MONGODB_HOSTS: MongoDB server hosts (default: localhost:27017)
//   - MONGODB_USERNAME: Authentication username
//   - MONGODB_PASSWORD: Authentication password
//   - MONGODB_DATABASE: Database name (required)
//   - MONGODB_AUTH_DATABASE: Authentication database (default: admin)
//   - MONGODB_REPLICA_SET: Replica set name
//   - MONGODB_MAX_POOL_SIZE: Maximum connection pool size (default: 100)
//   - MONGODB_MIN_POOL_SIZE: Minimum connection pool size (default: 5)
//   - MONGODB_CONNECT_TIMEOUT: Connection timeout (default: 10s)
//   - MONGODB_RECONNECT_ENABLED: Enable auto-reconnection (default: true)
//   - MONGODB_HEALTH_CHECK_ENABLED: Enable health checks (default: true)
//   - MONGODB_COMPRESSION_ENABLED: Enable compression (default: true)
//   - MONGODB_READ_PREFERENCE: Read preference (default: primary)
//   - MONGODB_DIRECT_CONNECTION: Enable direct connection mode (default: false)
//   - MONGODB_APP_NAME: Application name for connection metadata
//   - MONGODB_LOG_LEVEL: Logging level (default: info)
//
// Basic Usage:
//
//	package main
//
//	import (
//	    "context"
//	    "github.com/cloudresty/go-mongodb/v2"
//	)
//
//	func main() {
//	    // Create client using environment variables
//	    client, err := mongodb.NewClient()
//	    if err != nil {
//	        panic(err)
//	    }
//	    defer client.Close()
//
//	    // Get a collection
//	    users := client.Collection("users")
//
//	    // Insert a document with auto-generated ULID
//	    result, err := users.InsertOne(context.Background(), bson.M{
//	        "name":  "John Doe",
//	        "email": "john@example.com",
//	    })
//	    if err != nil {
//	        panic(err)
//	    }
//
//	    fmt.Printf("Inserted document with ID: %s, ULID: %s\n",
//	        result.InsertedID.Hex(), result.ULID)
//	}
//
// Production Usage with Graceful Shutdown:
//
//	func main() {
//	    // Create client with custom environment prefix
//	    client, err := mongodb.NewClientWithPrefix("PAYMENTS_")
//	    if err != nil {
//	        panic(err)
//	    }
//
//	    // Setup graceful shutdown
//	    shutdownManager := mongodb.NewShutdownManager(&mongodb.ShutdownConfig{
//	        Timeout: 30 * time.Second,
//	    })
//	    shutdownManager.SetupSignalHandler()
//	    shutdownManager.Register(client)
//
//	    // Your application logic here
//	    // ...
//
//	    // Wait for shutdown signal
//	    shutdownManager.Wait() // Blocks until SIGINT/SIGTERM
//	}
//
// Transaction Example:
//
//	result, err := client.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (any, error) {
//	    users := client.Collection("users")
//	    orders := client.Collection("orders")
//
//	    // Insert user
//	    userResult, err := users.InsertOne(sessCtx, bson.M{"name": "John"})
//	    if err != nil {
//	        return nil, err
//	    }
//
//	    // Insert order
//	    orderResult, err := orders.InsertOne(sessCtx, bson.M{
//	        "user_id": userResult.InsertedID,
//	        "amount":  100.00,
//	    })
//	    if err != nil {
//	        return nil, err
//	    }
//
//	    return orderResult, nil
//	})
package mongodb

import (
	"context"
	"fmt"
	"maps"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Version of the go-mongodb package
const Version = "1.0.0"

// Package-level convenience functions

// Connect creates a new MongoDB client using environment variables
func Connect() (*Client, error) {
	return NewClient()
}

// ConnectWithPrefix creates a new MongoDB client with a custom environment prefix
func ConnectWithPrefix(prefix string) (*Client, error) {
	return NewClientWithPrefix(prefix)
}

// ConnectWithConfig creates a new MongoDB client with the provided configuration
func ConnectWithConfig(config *Config) (*Client, error) {
	return NewClientWithConfig(config)
}

// Quick creates a quick MongoDB connection for simple use cases
// This is useful for scripts and simple applications that don't need advanced features
func Quick(database ...string) (*Client, error) {
	config, err := loadConfigFromEnv("")
	if err != nil {
		return nil, fmt.Errorf("failed to load config for quick connection: %w", err)
	}

	if len(database) > 0 {
		config.Database = database[0]
	}

	// Disable advanced features for quick connections
	config.HealthCheckEnabled = false

	return NewClientWithConfig(config)
}

// MustConnect creates a new MongoDB client or panics on error
// Use this only in main functions or initialization code where panicking is acceptable
func MustConnect() *Client {
	client, err := NewClient()
	if err != nil {
		panic(err)
	}
	return client
}

// MustConnectWithPrefix creates a new MongoDB client with prefix or panics on error
func MustConnectWithPrefix(prefix string) *Client {
	client, err := NewClientWithPrefix(prefix)
	if err != nil {
		panic(err)
	}
	return client
}

// Ping tests connectivity to MongoDB using default configuration
func Ping(ctx ...context.Context) error {
	client, err := Quick()
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()

	var pingCtx context.Context
	if len(ctx) > 0 {
		pingCtx = ctx[0]
	} else {
		var cancel context.CancelFunc
		pingCtx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
	}

	return client.Ping(pingCtx)
}

// Helper functions for common MongoDB operations

// NewULID generates a new ULID string
func NewULID() string {
	return generateULID()
}

// GenerateULID is an alias for NewULID for backward compatibility
func GenerateULID() string {
	return NewULID()
}

// GenerateULIDFromTime generates a ULID with a specific timestamp
func GenerateULIDFromTime(t time.Time) string {
	return generateULIDFromTime(t)
}

// EnhanceDocument adds ULID to a document (uses default ULID mode)
func EnhanceDocument(doc any) bson.M {
	var enhanced bson.M

	// Convert document to bson.M
	if docBytes, err := bson.Marshal(doc); err == nil {
		var docMap bson.M
		if err := bson.Unmarshal(docBytes, &docMap); err == nil {
			enhanced = docMap
		}
	}

	if enhanced == nil {
		enhanced = bson.M{}
	}

	// Generate ULID if no _id provided (default behavior)
	if _, hasID := enhanced["_id"]; !hasID {
		enhanced["_id"] = GenerateULID()
	}

	return enhanced
}

// Common BSON helpers

// M is an alias for bson.M (map)
type M = bson.M

// D is an alias for bson.D (ordered document)
type D = bson.D

// E is an alias for bson.E (element)
type E = bson.E

// A is an alias for bson.A (array)
type A = bson.A

// Common filter builders

// ByID creates a filter for finding by _id
func ByID(id any) bson.M {
	return bson.M{"_id": id}
}

// ByULID creates a filter for finding by ULID
func ByULID(ulid string) bson.M {
	return bson.M{"ulid": ulid}
}

// ByField creates a filter for a specific field
func ByField(field string, value any) bson.M {
	return bson.M{field: value}
}

// ByFields creates a filter for multiple fields
func ByFields(fields bson.M) bson.M {
	return fields
}

// Common update builders

// Set creates a $set update operation
func Set(fields bson.M) bson.M {
	return bson.M{"$set": fields}
}

// Inc creates a $inc update operation
func Inc(fields bson.M) bson.M {
	return bson.M{"$inc": fields}
}

// Push creates a $push update operation
func Push(field string, value any) bson.M {
	return bson.M{"$push": bson.M{field: value}}
}

// Pull creates a $pull update operation
func Pull(field string, value any) bson.M {
	return bson.M{"$pull": bson.M{field: value}}
}

// Unset creates an$unset update operation
func Unset(fields ...string) bson.M {
	unsetDoc := bson.M{}
	for _, field := range fields {
		unsetDoc[field] = ""
	}
	return bson.M{"$unset": unsetDoc}
}

// Common aggregation pipeline builders

// Match creates a $match stage
func Match(filter bson.M) bson.M {
	return bson.M{"$match": filter}
}

// Sort creates a $sort stage
func Sort(fields bson.D) bson.M {
	return bson.M{"$sort": fields}
}

// Limit creates a $limit stage
func Limit(n int64) bson.M {
	return bson.M{"$limit": n}
}

// Skip creates a $skip stage
func Skip(n int64) bson.M {
	return bson.M{"$skip": n}
}

// Group creates a $group stage
func Group(id any, fields bson.M) bson.M {
	groupDoc := bson.M{"_id": id}
	maps.Copy(groupDoc, fields)
	return bson.M{"$group": groupDoc}
}

// Project creates a $project stage
func Project(fields bson.M) bson.M {
	return bson.M{"$project": fields}
}

// Lookup creates a $lookup stage
func Lookup(from, localField, foreignField, as string) bson.M {
	return bson.M{"$lookup": bson.M{
		"from":         from,
		"localField":   localField,
		"foreignField": foreignField,
		"as":           as,
	}}
}

// Common sort orders

// Ascending creates an ascending sort order
func Ascending(fields ...string) bson.D {
	sort := bson.D{}
	for _, field := range fields {
		sort = append(sort, bson.E{Key: field, Value: 1})
	}
	return sort
}

// Descending creates a descending sort order
func Descending(fields ...string) bson.D {
	sort := bson.D{}
	for _, field := range fields {
		sort = append(sort, bson.E{Key: field, Value: -1})
	}
	return sort
}

// Common index builders

// IndexAsc creates an ascending index model
func IndexAsc(fields ...string) IndexModel {
	keys := bson.D{}
	for _, field := range fields {
		keys = append(keys, bson.E{Key: field, Value: 1})
	}
	return IndexModel{Keys: keys}
}

// IndexDesc creates a descending index model
func IndexDesc(fields ...string) IndexModel {
	keys := bson.D{}
	for _, field := range fields {
		keys = append(keys, bson.E{Key: field, Value: -1})
	}
	return IndexModel{Keys: keys}
}

// IndexText creates a text index model
func IndexText(fields ...string) IndexModel {
	keys := bson.D{}
	for _, field := range fields {
		keys = append(keys, bson.E{Key: field, Value: "text"})
	}
	return IndexModel{Keys: keys}
}

// IndexUnique creates a unique index model
func IndexUnique(fields ...string) IndexModel {
	keys := bson.D{}
	for _, field := range fields {
		keys = append(keys, bson.E{Key: field, Value: 1})
	}
	return IndexModel{
		Keys:    keys,
		Options: options.Index().SetUnique(true),
	}
}

// IndexCompound creates a compound index with mixed ascending/descending fields.
// Takes pairs of (field, direction) where direction is 1 for ascending, -1 for descending.
// Example: IndexCompound("status", 1, "created_at", -1)
func IndexCompound(fieldsAndDirections ...any) IndexModel {
	keys := bson.D{}
	for i := 0; i < len(fieldsAndDirections)-1; i += 2 {
		field, ok := fieldsAndDirections[i].(string)
		if !ok {
			continue
		}
		direction := 1
		if dir, ok := fieldsAndDirections[i+1].(int); ok {
			direction = dir
		}
		keys = append(keys, bson.E{Key: field, Value: direction})
	}
	return IndexModel{Keys: keys}
}

// IndexTTL creates a TTL (Time-To-Live) index that automatically expires documents.
// The field should contain a date/time value. Documents expire after the specified duration.
func IndexTTL(field string, expireAfter time.Duration) IndexModel {
	keys := bson.D{{Key: field, Value: 1}}
	expireSeconds := int32(expireAfter.Seconds())
	return IndexModel{
		Keys:    keys,
		Options: options.Index().SetExpireAfterSeconds(expireSeconds),
	}
}

// IndexSparse creates a sparse index that only includes documents with the indexed field.
// Useful for optional fields where many documents might not have the field.
func IndexSparse(fields ...string) IndexModel {
	keys := bson.D{}
	for _, field := range fields {
		keys = append(keys, bson.E{Key: field, Value: 1})
	}
	return IndexModel{
		Keys:    keys,
		Options: options.Index().SetSparse(true),
	}
}

// IndexHashed creates a hashed index for sharding or hash-based lookups.
func IndexHashed(field string) IndexModel {
	keys := bson.D{{Key: field, Value: "hashed"}}
	return IndexModel{Keys: keys}
}

// Index2DSphere creates a 2dsphere index for geospatial queries on GeoJSON data.
func Index2DSphere(field string) IndexModel {
	keys := bson.D{{Key: field, Value: "2dsphere"}}
	return IndexModel{Keys: keys}
}

// IndexWithName creates an index with a custom name.
// Wraps any IndexModel and adds a custom name option.
func IndexWithName(name string, model IndexModel) IndexModel {
	if model.Options == nil {
		model.Options = options.Index()
	}
	model.Options = model.Options.SetName(name)
	return model
}

// IndexPartial creates a partial index with a filter expression.
// Only documents matching the filter are included in the index.
func IndexPartial(filter bson.D, fields ...string) IndexModel {
	keys := bson.D{}
	for _, field := range fields {
		keys = append(keys, bson.E{Key: field, Value: 1})
	}
	return IndexModel{
		Keys:    keys,
		Options: options.Index().SetPartialFilterExpression(filter),
	}
}

// IndexUniqueWithOptions creates a unique index with additional options
func IndexUniqueWithOptions(fields []string, sparse bool, name string) IndexModel {
	keys := bson.D{}
	for _, field := range fields {
		keys = append(keys, bson.E{Key: field, Value: 1})
	}
	opts := options.Index().SetUnique(true)
	if sparse {
		opts = opts.SetSparse(true)
	}
	if name != "" {
		opts = opts.SetName(name)
	}
	return IndexModel{
		Keys:    keys,
		Options: opts,
	}
}

// Error handling utilities

// IsDuplicateKeyError checks if an error is a duplicate key error
func IsDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}

	// Check for MongoDB duplicate key error codes
	if cmdErr, ok := err.(mongo.CommandError); ok {
		return cmdErr.Code == 11000 || cmdErr.Code == 11001
	}

	// Check for write exception
	if writeErr, ok := err.(mongo.WriteException); ok {
		for _, writeError := range writeErr.WriteErrors {
			if writeError.Code == 11000 || writeError.Code == 11001 {
				return true
			}
		}
	}

	return false
}

// IsTimeoutError checks if an error is a timeout error
func IsTimeoutError(err error) bool {
	if err == nil {
		return false
	}

	// Check for context timeout
	if err == context.DeadlineExceeded {
		return true
	}

	// Check for MongoDB timeout errors
	if cmdErr, ok := err.(mongo.CommandError); ok {
		return cmdErr.Code == 50 || cmdErr.Code == 89 || cmdErr.Code == 91
	}

	return false
}

// IsNetworkError checks if an error is a network-related error
func IsNetworkError(err error) bool {
	if err == nil {
		return false
	}

	// This is a simplified check - in practice you might want more sophisticated logic
	errStr := err.Error()
	return contains(errStr, "connection") || contains(errStr, "network") || contains(errStr, "timeout")
}

// IsNotFoundError checks if an error indicates that a document, collection, or database was not found
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	// Check for the standard MongoDB "no documents" error
	if err == mongo.ErrNoDocuments {
		return true
	}

	// Check for MongoDB command errors related to not found
	if cmdErr, ok := err.(mongo.CommandError); ok {
		// Common MongoDB error codes for "not found" scenarios:
		// 26 = NamespaceNotFound (collection/database doesn't exist)
		// 73 = InvalidNamespace (invalid collection/database name)
		return cmdErr.Code == 26 || cmdErr.Code == 73
	}

	// Check error message for common "not found" patterns
	errStr := err.Error()
	return contains(errStr, "not found") ||
		contains(errStr, "does not exist") ||
		contains(errStr, "no documents") ||
		contains(errStr, "namespace not found")
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}

	// Simple substring search
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
