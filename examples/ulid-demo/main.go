package main

import (
	"context"
	"os"

	"github.com/cloudresty/emit"
	"github.com/cloudresty/go-mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func main() {
	emit.Info.Msg("Starting ULID demonstration example")

	client, err := mongodb.NewClient()
	if err != nil {
		emit.Error.StructuredFields("Failed to create client",
			emit.ZString("error", err.Error()))
		os.Exit(1)
	}
	defer client.Close()

	collection := client.Database("testdb").Collection("ulid_demo")
	ctx := context.Background()

	// Clean up existing data
	_, _ = collection.DeleteMany(ctx, bson.M{})

	emit.Info.Msg("Demonstrating automatic ULID generation")

	// Insert documents and show ULID enhancement
	for i := range 3 {
		testDoc := bson.M{
			"name":  "User " + string(rune(65+i)), // User A, User B, User C
			"email": "user" + string(rune(97+i)) + "@example.com",
			"index": i,
		}

		emit.Info.StructuredFields("Inserting document",
			emit.ZString("name", testDoc["name"].(string)),
			emit.ZString("email", testDoc["email"].(string)))

		result, err := collection.InsertOne(ctx, testDoc)
		if err != nil {
			emit.Error.StructuredFields("Failed to insert document",
				emit.ZString("error", err.Error()))
			continue
		}

		// Retrieve the inserted document to see ULID enhancement
		var insertedDoc bson.M
		err = collection.FindOne(ctx, bson.M{"_id": result.InsertedID}).Decode(&insertedDoc)
		if err != nil {
			emit.Error.StructuredFields("Failed to retrieve inserted document",
				emit.ZString("error", err.Error()))
			continue
		}

		// Show the enhanced document with ULID fields
		if ulidStr, ok := insertedDoc["ulid"].(string); ok {
			emit.Info.StructuredFields("Document enhanced with ULID",
				emit.ZString("name", insertedDoc["name"].(string)),
				emit.ZString("ulid", ulidStr),
				emit.ZInt("ulid_length", len(ulidStr)))
		}
	}

	// Demonstrate querying by ULID
	emit.Info.Msg("Demonstrating ULID-based queries")

	// Count all documents
	count, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		emit.Error.StructuredFields("Failed to count documents",
			emit.ZString("error", err.Error()))
	} else {
		emit.Info.StructuredFields("Total documents with ULIDs",
			emit.ZInt64("count", count))
	}

	// Find all documents and show their ULIDs
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		emit.Error.StructuredFields("Failed to find documents",
			emit.ZString("error", err.Error()))
	} else {
		defer cursor.Close(ctx)

		var docs []bson.M
		if err = cursor.All(ctx, &docs); err != nil {
			emit.Error.StructuredFields("Failed to decode documents",
				emit.ZString("error", err.Error()))
		} else {
			emit.Info.StructuredFields("Documents found",
				emit.ZInt("count", len(docs)))

			for i, doc := range docs {
				if ulid, ok := doc["ulid"].(string); ok {
					emit.Info.StructuredFields("Document",
						emit.ZInt("position", i+1),
						emit.ZString("name", doc["name"].(string)),
						emit.ZString("ulid", ulid))
				}
			}
		}
	}

	// Cleanup
	_, err = collection.DeleteMany(ctx, bson.M{})
	if err != nil {
		emit.Error.StructuredFields("Failed to cleanup test documents",
			emit.ZString("error", err.Error()))
	} else {
		emit.Info.Msg("Test documents cleaned up")
	}

	emit.Info.Msg("ULID demonstration completed successfully!")
}
