package main

import (
	"context"
	"os"

	"github.com/cloudresty/emit"
	"github.com/cloudresty/go-mongodb"
)

func main() {
	emit.Info.Msg("Starting environment configuration examples")

	// Example 1: Using default MONGODB_ prefix
	emit.Info.Msg("Creating client from environment variables (MONGODB_ prefix)")

	client, err := mongodb.NewClient()
	if err != nil {
		emit.Error.StructuredFields("Failed to create client from environment",
			emit.ZString("error", err.Error()))
		os.Exit(1)
	}
	defer client.Close()

	emit.Info.Msg("Client created successfully from environment variables")

	// Example 2: Using custom prefix (e.g., MYAPP_MONGODB_)
	emit.Info.Msg("Creating client from environment variables with custom prefix")

	clientWithPrefix, err := mongodb.NewClientWithPrefix("MYAPP_")
	if err != nil {
		emit.Error.StructuredFields("Failed to create client from environment with prefix",
			emit.ZString("error", err.Error()))
		// This might fail if MYAPP_ prefixed vars aren't set, which is expected
		emit.Warn.Msg("Custom prefix example failed (expected if MYAPP_* vars not set)")
	} else {
		defer clientWithPrefix.Close()
		emit.Info.Msg("Client with custom prefix created successfully")
	}

	// Example 3: Loading env config and modifying before use
	emit.Info.Msg("Loading environment config and customizing")

	config, err := mongodb.LoadConfig()
	if err != nil {
		emit.Error.StructuredFields("Failed to load config from environment",
			emit.ZString("error", err.Error()))
		os.Exit(1)
	}

	// Customize the config
	config.ConnectionName = "env-config-example"
	config.Database = "custom_database"

	emit.Info.StructuredFields("Loaded and customized config",
		emit.ZString("connection_name", config.ConnectionName),
		emit.ZString("database", config.Database),
		emit.ZString("host", config.Host),
		emit.ZInt("port", config.Port))

	// Create client with customized config
	customClient, err := mongodb.NewClientWithConfig(config)
	if err != nil {
		emit.Error.StructuredFields("Failed to create client with custom config",
			emit.ZString("error", err.Error()))
		os.Exit(1)
	}
	defer customClient.Close()

	emit.Info.Msg("Client with customized config created successfully")

	// Test the connections
	ctx := context.Background()
	if err := client.Ping(ctx); err != nil {
		emit.Error.StructuredFields("Default client ping failed",
			emit.ZString("error", err.Error()))
	} else {
		emit.Info.Msg("Default client ping successful")
	}

	if err := customClient.Ping(ctx); err != nil {
		emit.Error.StructuredFields("Custom client ping failed",
			emit.ZString("error", err.Error()))
	} else {
		emit.Info.Msg("Custom client ping successful")
	}

	emit.Info.Msg("Environment configuration examples completed!")
}
