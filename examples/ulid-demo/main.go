package main

import (
	"context"
	"os"

	"github.com/cloudresty/emit"
	"github.com/cloudresty/go-mongodb"
	"github.com/cloudresty/go-mongodb/filter"
)

// User represents a user document with ULID support
type User struct {
	ID    string `bson:"_id,omitempty" json:"id"`
	ULID  string `bson:"ulid,omitempty" json:"ulid"`
	Name  string `bson:"name" json:"name"`
	Email string `bson:"email" json:"email"`
	Index int    `bson:"index" json:"index"`
}

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
	_, _ = collection.DeleteMany(ctx, filter.New())

	emit.Info.Msg("Demonstrating automatic ULID generation")

	// Insert documents and show ULID enhancement
	for i := range 3 {
		user := User{
			Name:  string(rune('A'+i)) + " User", // A User, B User, C User
			Email: "user" + string(rune('a'+i)) + "@example.com",
			Index: i,
		}

		emit.Info.StructuredFields("Inserting document",
			emit.ZString("name", user.Name),
			emit.ZString("email", user.Email))

		result, err := collection.InsertOne(ctx, user)
		if err != nil {
			emit.Error.StructuredFields("Failed to insert document",
				emit.ZString("error", err.Error()))
			continue
		}

		// Retrieve the inserted document to see ULID enhancement
		var insertedUser User
		err = collection.FindOne(ctx, filter.Eq("_id", result.InsertedID)).Decode(&insertedUser)
		if err != nil {
			emit.Error.StructuredFields("Failed to retrieve inserted document",
				emit.ZString("error", err.Error()))
			continue
		}

		// Show the enhanced document with ULID fields
		emit.Info.StructuredFields("Document enhanced with ULID",
			emit.ZString("name", insertedUser.Name),
			emit.ZString("ulid", insertedUser.ULID),
			emit.ZInt("ulid_length", len(insertedUser.ULID)))
	}

	// Demonstrate querying by ULID
	emit.Info.Msg("Demonstrating ULID-based queries")

	// Count all documents
	count, err := collection.CountDocuments(ctx, filter.New())
	if err != nil {
		emit.Error.StructuredFields("Failed to count documents",
			emit.ZString("error", err.Error()))
	} else {
		emit.Info.StructuredFields("Total documents with ULIDs",
			emit.ZInt64("count", count))
	}

	// Find all documents and show their ULIDs
	cursor, err := collection.Find(ctx, filter.New())
	if err != nil {
		emit.Error.StructuredFields("Failed to find documents",
			emit.ZString("error", err.Error()))
	} else {
		defer cursor.Close(ctx)

		var users []User
		if err = cursor.All(ctx, &users); err != nil {
			emit.Error.StructuredFields("Failed to decode documents",
				emit.ZString("error", err.Error()))
		} else {
			emit.Info.StructuredFields("Documents found",
				emit.ZInt("count", len(users)))

			for i, user := range users {
				emit.Info.StructuredFields("Document",
					emit.ZInt("position", i+1),
					emit.ZString("name", user.Name),
					emit.ZString("ulid", user.ULID))
			}
		}
	}

	// Cleanup
	_, err = collection.DeleteMany(ctx, filter.New())
	if err != nil {
		emit.Error.StructuredFields("Failed to cleanup test documents",
			emit.ZString("error", err.Error()))
	} else {
		emit.Info.Msg("Test documents cleaned up")
	}

	emit.Info.Msg("ULID demonstration completed successfully!")
}
