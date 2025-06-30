package main

import (
	"context"
	"os"
	"time"

	"github.com/cloudresty/emit"
	"github.com/cloudresty/go-mongodb"
	"github.com/cloudresty/go-mongodb/filter"
)

// TestDoc represents a test document for reconnection testing
type TestDoc struct {
	ID        string    `bson:"_id,omitempty" json:"id"`
	Iteration int       `bson:"iteration" json:"iteration"`
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
	Message   string    `bson:"message" json:"message"`
}

func main() {
	emit.Info.Msg("Starting MongoDB auto-reconnection test example")

	client, err := mongodb.NewClient()
	if err != nil {
		emit.Error.StructuredFields("Failed to create client",
			emit.ZString("error", err.Error()))
		os.Exit(1)
	}
	defer client.Close()

	collection := client.Database("testdb").Collection("reconnection_test")
	ctx := context.Background()

	emit.Info.Msg("Testing MongoDB connection and auto-reconnection")

	// Test loop - continuously try operations
	for i := range 10 {
		emit.Info.StructuredFields("Testing operation",
			emit.ZInt("iteration", i+1))

		// Try to ping MongoDB to check connectivity
		err := client.Ping(ctx)
		if err != nil {
			emit.Error.StructuredFields("Ping failed - MongoDB may be unavailable",
				emit.ZString("error", err.Error()))
		} else {
			emit.Info.Msg("Ping successful")
		}

		// Try a basic operation
		testDoc := TestDoc{
			Iteration: i + 1,
			Timestamp: time.Now(),
			Message:   "reconnection test",
		}

		result, err := collection.InsertOne(ctx, testDoc)
		if err != nil {
			emit.Error.StructuredFields("Insert operation failed",
				emit.ZInt("iteration", i+1),
				emit.ZString("error", err.Error()))
		} else {
			emit.Info.StructuredFields("Insert operation successful",
				emit.ZInt("iteration", i+1),
				emit.ZString("inserted_id", result.InsertedID))
		}

		// Try to count documents
		count, err := collection.CountDocuments(ctx, filter.New())
		if err != nil {
			emit.Error.StructuredFields("Count operation failed",
				emit.ZString("error", err.Error()))
		} else {
			emit.Info.StructuredFields("Count operation successful",
				emit.ZInt64("documents", count))
		}

		// Wait before next iteration
		if i < 9 { // Don't wait after the last iteration
			emit.Info.Msg("Waiting 3 seconds before next test...")
			emit.Info.Msg("(You can stop/restart MongoDB during this time to test reconnection)")
			time.Sleep(3 * time.Second)
		}
	}

	// Final cleanup
	count, err := collection.CountDocuments(ctx, filter.New())
	if err != nil {
		emit.Error.StructuredFields("Failed to count final documents",
			emit.ZString("error", err.Error()))
	} else {
		emit.Info.StructuredFields("Final document count",
			emit.ZInt64("total_documents", count))

		// Clean up test documents
		_, err = collection.DeleteMany(ctx, filter.Eq("message", "reconnection test"))
		if err != nil {
			emit.Error.StructuredFields("Failed to cleanup test documents",
				emit.ZString("error", err.Error()))
		} else {
			emit.Info.Msg("Test documents cleaned up")
		}
	}

	emit.Info.Msg("MongoDB auto-reconnection test completed!")
	emit.Info.Msg("Results show how the client handles connection issues and automatic reconnection")
}
