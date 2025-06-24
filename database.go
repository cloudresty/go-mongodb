package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudresty/emit"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Collection returns a MongoDB collection instance
func (c *Client) Collection(name string) *Collection {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return &Collection{
		collection: c.database.Collection(name),
		client:     c,
		name:       name,
	}
}

// Database returns the current database instance
func (c *Client) Database(name ...string) *mongo.Database {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if len(name) > 0 {
		return c.client.Database(name[0])
	}
	return c.database
}

// Ping tests the connection to MongoDB
func (c *Client) Ping(ctx context.Context) error {
	c.mutex.RLock()
	client := c.client
	c.mutex.RUnlock()

	if client == nil {
		return fmt.Errorf("client is not connected")
	}

	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
	}

	return client.Ping(ctx, nil)
}

// StartSession starts a new session for transactions
func (c *Client) StartSession(opts ...options.Lister[options.SessionOptions]) (*mongo.Session, error) {
	c.mutex.RLock()
	client := c.client
	c.mutex.RUnlock()

	if client == nil {
		return nil, fmt.Errorf("client is not connected")
	}

	return client.StartSession(opts...)
}

// WithTransaction executes a function within a transaction
func (c *Client) WithTransaction(ctx context.Context, fn func(context.Context) (any, error), opts ...options.Lister[options.TransactionOptions]) (any, error) {
	session, err := c.StartSession()
	if err != nil {
		return nil, fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	callback := func(sessCtx context.Context) (any, error) {
		return fn(sessCtx)
	}

	result, err := session.WithTransaction(ctx, callback, opts...)
	if err != nil {
		emit.Error.StructuredFields("Transaction failed",
			emit.ZString("error", err.Error()))
		return nil, fmt.Errorf("transaction failed: %w", err)
	}

	emit.Debug.Msg("Transaction completed successfully")
	return result, nil
}

// ListDatabases lists all databases
func (c *Client) ListDatabases(ctx context.Context, filter any, opts ...options.Lister[options.ListDatabasesOptions]) (mongo.ListDatabasesResult, error) {
	c.mutex.RLock()
	client := c.client
	c.mutex.RUnlock()

	if client == nil {
		return mongo.ListDatabasesResult{}, fmt.Errorf("client is not connected")
	}

	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
	}

	return client.ListDatabases(ctx, filter, opts...)
}

// ListCollections lists all collections in the current database
func (c *Client) ListCollections(ctx context.Context, filter any, opts ...options.Lister[options.ListCollectionsOptions]) (*mongo.Cursor, error) {
	c.mutex.RLock()
	database := c.database
	c.mutex.RUnlock()

	if database == nil {
		return nil, fmt.Errorf("database is not available")
	}

	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
	}

	return database.ListCollections(ctx, filter, opts...)
}

// DropDatabase drops the current database
func (c *Client) DropDatabase(ctx context.Context) error {
	c.mutex.RLock()
	database := c.database
	c.mutex.RUnlock()

	if database == nil {
		return fmt.Errorf("database is not available")
	}

	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	err := database.Drop(ctx)
	if err != nil {
		emit.Error.StructuredFields("Failed to drop database",
			emit.ZString("database", c.config.Database),
			emit.ZString("error", err.Error()))
		return fmt.Errorf("failed to drop database: %w", err)
	}

	emit.Info.StructuredFields("Database dropped successfully",
		emit.ZString("database", c.config.Database))

	return nil
}

// GetStats returns database statistics
func (c *Client) GetStats(ctx context.Context) (bson.M, error) {
	c.mutex.RLock()
	database := c.database
	c.mutex.RUnlock()

	if database == nil {
		return nil, fmt.Errorf("database is not available")
	}

	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
	}

	var result bson.M
	err := database.RunCommand(ctx, bson.D{bson.E{Key: "dbStats", Value: 1}}).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to get database stats: %w", err)
	}

	return result, nil
}

// CreateIndex creates an index on the specified collection
func (c *Client) CreateIndex(ctx context.Context, collectionName string, index IndexModel) (string, error) {
	collection := c.Collection(collectionName)
	return collection.CreateIndex(ctx, index)
}

// CreateIndexes creates multiple indexes on the specified collection
func (c *Client) CreateIndexes(ctx context.Context, collectionName string, indexes []IndexModel) ([]string, error) {
	collection := c.Collection(collectionName)
	return collection.CreateIndexes(ctx, indexes)
}

// DropIndex drops an index from the specified collection
func (c *Client) DropIndex(ctx context.Context, collectionName string, indexName string) error {
	collection := c.Collection(collectionName)
	return collection.DropIndex(ctx, indexName)
}

// ListIndexes lists all indexes for the specified collection
func (c *Client) ListIndexes(ctx context.Context, collectionName string) (*mongo.Cursor, error) {
	collection := c.Collection(collectionName)
	return collection.ListIndexes(ctx)
}
