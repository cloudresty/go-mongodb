package main

import (
	"context"
	"os"

	"github.com/cloudresty/emit"
	"github.com/cloudresty/go-mongodb"
)

func main() {
	emit.Info.Msg("Starting basic MongoDB client example")

	// Create client from environment variables
	client, err := mongodb.NewClient()
	if err != nil {
		emit.Error.StructuredFields("Failed to create client",
			emit.ZString("error", err.Error()),
			emit.ZString("hint", "Set MONGODB_* environment variables or defaults will be used"))
		os.Exit(1)
	}
	defer client.Close()

	emit.Info.Msg("MongoDB client connected successfully")

	// Test connection
	ctx := context.Background()
	err = client.Ping(ctx)
	if err != nil {
		emit.Error.StructuredFields("Failed to ping MongoDB",
			emit.ZString("error", err.Error()))
		os.Exit(1)
	}

	emit.Info.Msg("MongoDB ping successful")

	// Get database and collection
	db := client.Database("testdb")
	_ = db.Collection("testcollection") // Just to show usage

	emit.Info.StructuredFields("Connected to database and collection",
		emit.ZString("database", "testdb"),
		emit.ZString("collection", "testcollection"))

	emit.Info.Msg("Basic client example completed successfully!")
}
