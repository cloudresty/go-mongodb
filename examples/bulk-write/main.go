// Package main demonstrates enterprise-grade BulkWrite usage with the go-mongodb v2 library.
//
// This example shows:
//   - Mixed write operations (insert, update, upsert, delete) in a single round-trip
//   - Automatic ULID injection into InsertOneModel documents
//   - Ordered vs. unordered execution trade-offs
//   - Inspecting the BulkWriteResult, including generated InsertedIDs
//
// Run:
//
//	export MONGODB_HOSTS=localhost:27017
//	export MONGODB_DATABASE=bulk_write_example
//	go run examples/bulk-write/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	mongodb "github.com/cloudresty/go-mongodb/v2"
	"github.com/cloudresty/go-mongodb/v2/filter"
	"github.com/cloudresty/go-mongodb/v2/update"
)

// User is the domain document stored in MongoDB.
type User struct {
	ID        string    `bson:"_id,omitempty" json:"id,omitempty"`
	Name      string    `bson:"name"          json:"name"`
	Email     string    `bson:"email"         json:"email"`
	Active    bool      `bson:"active"        json:"active"`
	CreatedAt time.Time `bson:"created_at"    json:"created_at"`
}

func main() {
	// ── Client setup ─────────────────────────────────────────────────────────
	client, err := mongodb.NewClient(mongodb.FromEnv())
	if err != nil {
		log.Fatalf("Failed to create MongoDB client: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("Failed to close MongoDB client: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}
	fmt.Println("✓ Connected to MongoDB")

	col := client.Database(getEnv("MONGODB_DATABASE", "bulk_write_example")).Collection("users")

	// ── Seed: insert a couple of users that the bulk batch will operate on ───
	seed := []any{
		User{Name: "Alice", Email: "alice@example.com", Active: true, CreatedAt: time.Now()},
		User{Name: "Bob", Email: "bob@example.com", Active: true, CreatedAt: time.Now()},
	}
	_, err = col.InsertMany(ctx, seed)
	if err != nil {
		log.Fatalf("Seed failed: %v", err)
	}
	fmt.Println("✓ Seeded 2 existing users (Alice, Bob)")

	// ── Build a mixed batch of write models ──────────────────────────────────
	//
	// 1. Insert new users — ULIDs are injected automatically by the library.
	// 2. Update Alice's status using the type-safe update builder.
	// 3. Upsert an unknown user (Eve) — creates the document if absent.
	// 4. Delete Bob.
	var models []mongo.WriteModel

	// (1) Inserts: no _id required — the library generates a ULID for each.
	for i := 1; i <= 3; i++ {
		models = append(models, mongo.NewInsertOneModel().SetDocument(User{
			Name:      fmt.Sprintf("New User %d", i),
			Email:     fmt.Sprintf("new.user%d@example.com", i),
			Active:    true,
			CreatedAt: time.Now(),
		}))
	}

	// (2) Update Alice — deactivate her account.
	models = append(models, mongo.NewUpdateOneModel().
		SetFilter(filter.Eq("email", "alice@example.com")).
		SetUpdate(update.Set("active", false).Set("updated_at", time.Now())))

	// (3) Upsert Eve — insert if missing, skip modification if already present.
	models = append(models, mongo.NewUpdateOneModel().
		SetFilter(filter.Eq("email", "eve@example.com")).
		SetUpdate(update.Set("name", "Eve").Set("active", true).Set("created_at", time.Now())).
		SetUpsert(true))

	// (4) Delete Bob.
	models = append(models, mongo.NewDeleteOneModel().
		SetFilter(filter.Eq("email", "bob@example.com")))

	// ── Execute BulkWrite in a single round-trip ─────────────────────────────
	// SetOrdered(true) — stop on first error (default).
	// SetOrdered(false) — continue past errors for higher throughput.
	result, err := col.BulkWrite(ctx, models,
		options.BulkWrite().SetOrdered(true))
	if err != nil {
		log.Fatalf("BulkWrite failed: %v", err)
	}

	// ── Inspect the result ───────────────────────────────────────────────────
	fmt.Println("\n── BulkWrite Result ──────────────────────────────")
	fmt.Printf("  Inserted:  %d\n", result.InsertedCount)
	fmt.Printf("  Matched:   %d\n", result.MatchedCount)
	fmt.Printf("  Modified:  %d\n", result.ModifiedCount)
	fmt.Printf("  Deleted:   %d\n", result.DeletedCount)
	fmt.Printf("  Upserted:  %d\n", result.UpsertedCount)

	if len(result.InsertedIDs) > 0 {
		fmt.Println("\n  Auto-generated ULIDs (model index → ULID):")
		for idx, id := range result.InsertedIDs {
			fmt.Printf("    model[%d] → %v\n", idx, id)
		}
	}

	if len(result.UpsertedIDs) > 0 {
		fmt.Println("\n  Upserted document IDs (model index → ID):")
		for idx, id := range result.UpsertedIDs {
			fmt.Printf("    model[%d] → %v\n", idx, id)
		}
	}

	fmt.Println("\n✓ BulkWrite example completed successfully")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
