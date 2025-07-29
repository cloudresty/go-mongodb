package main

import (
	"context"
	"log"
	"os"
	"time"

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
	log.Println("Starting MongoDB auto-reconnection test example")

	client, err := mongodb.NewClient()
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		os.Exit(1)
	}
	defer client.Close()

	collection := client.Database("testdb").Collection("reconnection_test")

	log.Println("Testing MongoDB connection and auto-reconnection")

	for i := 0; i < 10; i++ {
		log.Printf("Testing operation - iteration %d", i+1)

		// Test connection with ping
		if err := client.Ping(context.Background()); err != nil {
			log.Printf("Ping failed - MongoDB may be unavailable: %v", err)
		} else {
			log.Println("Ping successful")
		}

		// Test write operation
		testDoc := TestDoc{
			Iteration: i + 1,
			Timestamp: time.Now(),
			Message:   "Auto-reconnection test document",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		result, err := collection.InsertOne(ctx, testDoc)
		cancel()

		if err != nil {
			log.Printf("Insert operation failed - iteration %d: %v", i+1, err)
		} else {
			log.Printf("Insert operation successful - iteration %d, ID: %v", i+1, result.InsertedID)
		}

		// Test read operation
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		count, err := collection.CountDocuments(ctx, filter.New())
		cancel()

		if err != nil {
			log.Printf("Count operation failed: %v", err)
		} else {
			log.Printf("Count operation successful: %d documents", count)
		}

		// Wait before next iteration
		log.Printf("Waiting 3 seconds before next iteration...")
		time.Sleep(3 * time.Second)
	}

	log.Println("Reconnection test completed")

	// Final verification - show all test documents
	log.Println("Retrieving all test documents:")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, filter.New())
	if err != nil {
		log.Printf("Failed to retrieve documents: %v", err)
		return
	}
	defer cursor.Close(ctx)

	var docs []TestDoc
	if err := cursor.All(ctx, &docs); err != nil {
		log.Printf("Failed to decode documents: %v", err)
		return
	}

	log.Printf("Retrieved %d test documents:", len(docs))
	for _, doc := range docs {
		log.Printf("- Iteration %d: %s (created at %s)",
			doc.Iteration, doc.Message, doc.Timestamp.Format(time.RFC3339))
	}

	// Clean up test data
	deleteResult, err := collection.DeleteMany(ctx, filter.New())
	if err != nil {
		log.Printf("Failed to clean up test data: %v", err)
	} else {
		log.Printf("Cleaned up %d test documents", deleteResult.DeletedCount)
	}

	log.Println("MongoDB auto-reconnection test example completed")
}
