// go-mongodb v2 Example: ULID Demo
//
// This example demonstrates ULID-based document IDs with v2.
//
// v2 Changes:
// - mongoid.NewULID() now panics on entropy failure (fail-fast prevents silent corruption)
// - Use mongoid.NewULIDWithError() for explicit error handling
// - ID field must be string type when using IDModeULID (type validation enforced)
package main

import (
	"context"
	"log"
	"os"

	"github.com/cloudresty/go-mongodb/v2"
	"github.com/cloudresty/go-mongodb/v2/filter"
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
	log.Println("Starting ULID demonstration example")

	client, err := mongodb.NewClient()
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		os.Exit(1)
	}
	defer func() { _ = client.Close() }()

	collection := client.Database("testdb").Collection("ulid_demo")
	ctx := context.Background()

	// Clean up existing data
	_, _ = collection.DeleteMany(ctx, filter.New())

	log.Println("Demonstrating automatic ULID generation")

	// Insert documents and show ULID enhancement
	users := []User{
		{Name: "Alice Johnson", Email: "alice@example.com", Index: 1},
		{Name: "Bob Smith", Email: "bob@example.com", Index: 2},
		{Name: "Carol Davis", Email: "carol@example.com", Index: 3},
	}

	for _, user := range users {
		log.Printf("Inserting document: %s (%s)", user.Name, user.Email)

		result, err := collection.InsertOne(ctx, user)
		if err != nil {
			log.Printf("Failed to insert document: %v", err)
			continue
		}

		// Retrieve the document to see the generated ULID
		var insertedUser User
		err = collection.FindOne(ctx, filter.Eq("_id", result.InsertedID)).Decode(&insertedUser)
		if err != nil {
			log.Printf("Failed to retrieve inserted document: %v", err)
			continue
		}

		log.Printf("Document enhanced with ULID: %s (ULID: %s, Length: %d)",
			insertedUser.Name, insertedUser.ULID, len(insertedUser.ULID))
	}

	// Demonstrate ULID-based queries
	log.Println("Demonstrating ULID-based queries")

	// Count documents with ULIDs
	count, err := collection.CountDocuments(ctx, filter.New())
	if err != nil {
		log.Printf("Failed to count documents: %v", err)
	} else {
		log.Printf("Total documents with ULIDs: %d", count)
	}

	// Find all documents and show their ULIDs
	cursor, err := collection.Find(ctx, filter.New())
	if err != nil {
		log.Printf("Failed to find documents: %v", err)
		return
	}
	defer func() { _ = cursor.Close(ctx) }()

	var users_result []User
	if err := cursor.All(ctx, &users_result); err != nil {
		log.Printf("Failed to decode documents: %v", err)
		return
	}

	log.Println("All documents with ULIDs:")
	for _, user := range users_result {
		log.Printf("- %s (%s): ULID=%s", user.Name, user.Email, user.ULID)
	}

	// Demonstrate ULID sorting (ULIDs are naturally sortable by creation time)
	log.Println("Demonstrating ULID time-based ordering:")
	for i, user := range users_result {
		log.Printf("%d. %s - ULID: %s", i+1, user.Name, user.ULID)
	}

	// Example: Find by ULID pattern (first few characters)
	if len(users_result) > 0 {
		firstULID := users_result[0].ULID
		if len(firstULID) >= 4 {
			prefix := firstULID[:4]
			log.Printf("Searching for documents with ULID prefix '%s':", prefix)

			// MongoDB regex search using the prefix
			regexFilter := filter.Regex("ulid", "^"+prefix)
			prefixCursor, err := collection.Find(ctx, regexFilter)
			if err != nil {
				log.Printf("Failed to search by ULID prefix: %v", err)
			} else {
				defer func() { _ = prefixCursor.Close(ctx) }()
				var prefixResults []User
				if err := prefixCursor.All(ctx, &prefixResults); err != nil {
					log.Printf("Failed to decode prefix results: %v", err)
				} else {
					log.Printf("Found %d documents with ULID prefix '%s'", len(prefixResults), prefix)
				}
			}
		}
	}

	log.Println("ULID demonstration completed successfully!")
}
