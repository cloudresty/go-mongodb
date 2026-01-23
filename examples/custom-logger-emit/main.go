package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cloudresty/emit"
	"github.com/cloudresty/go-mongodb/v2"
)

// EmitAdapter adapts the emit logger to satisfy the mongodb.Logger interface
type EmitAdapter struct{}

// Info implements mongodb.Logger.Info using emit's structured logging
func (e EmitAdapter) Info(msg string, fields ...any) {
	e.logWithFields(emit.Info.StructuredFields, emit.Info.Msg, msg, fields...)
}

// Warn implements mongodb.Logger.Warn using emit's structured logging
func (e EmitAdapter) Warn(msg string, fields ...any) {
	e.logWithFields(emit.Warn.StructuredFields, emit.Warn.Msg, msg, fields...)
}

// Error implements mongodb.Logger.Error using emit's structured logging
func (e EmitAdapter) Error(msg string, fields ...any) {
	e.logWithFields(emit.Error.StructuredFields, emit.Error.Msg, msg, fields...)
}

// Debug implements mongodb.Logger.Debug using emit's structured logging
func (e EmitAdapter) Debug(msg string, fields ...any) {
	e.logWithFields(emit.Debug.StructuredFields, emit.Debug.Msg, msg, fields...)
}

// logWithFields is a helper method that converts key-value pairs to emit fields
func (e EmitAdapter) logWithFields(structuredLogger func(string, ...emit.ZField), msgLogger func(string), msg string, fields ...any) {
	if len(fields) == 0 {
		msgLogger(msg)
		return
	}

	// Convert key-value pairs to emit ZFields
	emitFields := make([]emit.ZField, 0, len(fields)/2)
	for i := 0; i < len(fields)-1; i += 2 {
		key, ok := fields[i].(string)
		if !ok {
			continue
		}

		value := fields[i+1]
		switch v := value.(type) {
		case string:
			emitFields = append(emitFields, emit.ZString(key, v))
		case int:
			emitFields = append(emitFields, emit.ZInt(key, v))
		case int64:
			emitFields = append(emitFields, emit.ZInt64(key, v))
		case time.Duration:
			emitFields = append(emitFields, emit.ZDuration(key, v))
		case bool:
			emitFields = append(emitFields, emit.ZBool(key, v))
		case error:
			emitFields = append(emitFields, emit.ZString(key, v.Error()))
		default:
			// For other types, convert to string
			emitFields = append(emitFields, emit.ZString(key, fmt.Sprintf("%v", v)))
		}
	}

	structuredLogger(msg, emitFields...)
}

func main() {
	// Create an emit logger adapter
	emitLogger := EmitAdapter{}

	// Create MongoDB client with custom emit logger
	client, err := mongodb.NewClient(
		mongodb.WithDatabase("example_db"),
		mongodb.WithAppName("emit-logger-example"),
		mongodb.WithLogger(emitLogger), // Inject our emit logger adapter
	)
	if err != nil {
		log.Fatalf("Failed to create MongoDB client: %v", err)
	}
	defer client.Close()

	emit.Info.Msg("MongoDB client created with emit logger integration")

	// Example operations - these will now use our emit logger for internal logging
	collection := client.Collection("users")

	// Insert a document
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user := map[string]interface{}{
		"name":       "John Doe",
		"email":      "john@example.com",
		"created_at": time.Now(),
	}

	result, err := collection.InsertOne(ctx, user)
	if err != nil {
		emit.Error.StructuredFields("Failed to insert user",
			emit.ZString("error", err.Error()))
		return
	}

	emit.Info.StructuredFields("User inserted successfully",
		emit.ZString("id", result.InsertedID),
		emit.ZString("collection", "users"))

	// Health check - this will also use our logger
	health := client.HealthCheck()
	if health.IsHealthy {
		emit.Info.StructuredFields("MongoDB connection is healthy",
			emit.ZDuration("latency", health.Latency))
	} else {
		emit.Warn.StructuredFields("MongoDB connection unhealthy",
			emit.ZString("error", health.Error))
	}

	emit.Info.Msg("Example completed successfully")
}
