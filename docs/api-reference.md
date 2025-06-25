# API Reference

[Home](../README.md) &nbsp;/&nbsp; [Docs](README.md) &nbsp;/&nbsp; API Reference

&nbsp;

This document provides a comprehensive overview of all available functions in the go-mongodb package.

&nbsp;

## Core Functions

&nbsp;

### Client Creation

| Function | Description |
|----------|-------------|
| `NewClient()` | Creates a client using MONGODB_* environment variables |
| `NewClientWithPrefix(prefix)` | Creates a client with custom environment prefix (e.g., MYAPP_MONGODB_*) |
| `NewClientWithConfig(config)` | Creates a client with explicit configuration |

🔝 [back to top](#api-reference)

&nbsp;

### Database Operations

| Function | Description |
|----------|-------------|
| `client.Database(name)` | Get a database instance |
| `client.Collection(name)` | Get a collection instance (uses default database) |
| `client.IsConnected()` | Check if client is connected |
| `client.HealthCheck()` | Perform health check and return status |
| `client.Close()` | Close the connection gracefully |

🔝 [back to top](#api-reference)

&nbsp;

## Environment Configuration

| Function | Description |
|----------|-------------|
| `LoadConfig()` | Loads configuration from MONGODB_* environment variables |
| `LoadConfigWithPrefix(prefix)` | Loads configuration with custom prefix (e.g., MYAPP_MONGODB_*) |

🔝 [back to top](#api-reference)

&nbsp;

## Collection Operations

&nbsp;

### Document Operations

| Function | Description |
|----------|-------------|
| `collection.InsertOne(ctx, doc)` | Insert a single document with auto-generated ULID |
| `collection.InsertMany(ctx, docs)` | Insert multiple documents with auto-generated ULIDs |
| `collection.FindOne(ctx, filter)` | Find a single document |
| `collection.Find(ctx, filter, opts)` | Find multiple documents with options |
| `collection.UpdateOne(ctx, filter, update)` | Update a single document |
| `collection.UpdateMany(ctx, filter, update)` | Update multiple documents |
| `collection.ReplaceOne(ctx, filter, replacement)` | Replace a single document |
| `collection.DeleteOne(ctx, filter)` | Delete a single document |
| `collection.DeleteMany(ctx, filter)` | Delete multiple documents |

🔝 [back to top](#api-reference)

&nbsp;

### Aggregation Operations

| Function | Description |
|----------|-------------|
| `collection.Aggregate(ctx, pipeline, opts)` | Execute aggregation pipeline |
| `collection.CountDocuments(ctx, filter)` | Count documents matching filter |
| `collection.EstimatedDocumentCount(ctx)` | Get estimated document count |
| `collection.Distinct(ctx, fieldName, filter)` | Get distinct values for a field |

🔝 [back to top](#api-reference)

&nbsp;

### Index Operations

| Function | Description |
|----------|-------------|
| `collection.CreateIndex(ctx, keys, opts)` | Create a single index |
| `collection.CreateIndexes(ctx, models)` | Create multiple indexes |
| `collection.DropIndex(ctx, name)` | Drop an index by name |
| `collection.DropIndexes(ctx)` | Drop all indexes |
| `collection.ListIndexes(ctx)` | List all indexes |

🔝 [back to top](#api-reference)

&nbsp;

## Transaction Operations

| Function | Description |
|----------|-------------|
| `client.StartSession(opts)` | Start a new session |
| `session.StartTransaction(opts)` | Start a transaction |
| `session.CommitTransaction(ctx)` | Commit the current transaction |
| `session.AbortTransaction(ctx)` | Abort the current transaction |
| `session.WithTransaction(ctx, fn)` | Execute function within a transaction |

🔝 [back to top](#api-reference)

&nbsp;

## ULID Operations

| Function | Description |
|----------|-------------|
| `GenerateULID()` | Generate a new ULID string |
| `GenerateULIDFromTime(time)` | Generate a ULID with a specific timestamp |
| `EnhanceDocument(doc)` | Add ULID and metadata to a document |
| `NewULID()` | Alias for GenerateULID() |

🔝 [back to top](#api-reference)

&nbsp;

## Configuration Types

&nbsp;

### Config

```go
type Config struct {
    // Connection settings
    Host                 string
    Port                 int
    Username             string
    Password             string
    Database             string
    AuthDatabase         string
    ReplicaSet           string

    // Pool settings
    MaxPoolSize          uint64
    MinPoolSize          uint64
    MaxIdleTime          time.Duration
    MaxConnIdleTime      time.Duration

    // Timeout settings
    ConnectTimeout       time.Duration
    ServerSelectTimeout  time.Duration
    SocketTimeout        time.Duration

    // Reconnection settings
    ReconnectEnabled     bool
    ReconnectDelay       time.Duration
    MaxReconnectDelay    time.Duration
    ReconnectBackoff     float64
    MaxReconnectAttempts int

    // Health check settings
    HealthCheckEnabled   bool
    HealthCheckInterval  time.Duration

    // Performance settings
    CompressionEnabled   bool
    CompressionAlgorithm string
    ReadPreference       string
    WriteConcern         string
    ReadConcern          string

    // Application settings
    AppName              string
    ConnectionName       string

    // ID Generation settings
    IDMode               IDMode        // "ulid", "objectid", or "custom"

    // Logging settings
    LogLevel             string
    LogFormat            string
}
```

🔝 [back to top](#api-reference)

&nbsp;

### IDMode

```go
type IDMode string

const (
    // IDModeULID generates ULID strings as document IDs (default)
    IDModeULID     IDMode = "ulid"

    // IDModeObjectID generates MongoDB ObjectIDs as document IDs
    IDModeObjectID IDMode = "objectid"

    // IDModeCustom allows users to provide their own _id fields
    IDModeCustom   IDMode = "custom"
)
```

The IDMode controls how document IDs are generated:

- **`ulid`** (default): Generates 26-character ULID strings with temporal ordering
- **`objectid`**: Generates standard MongoDB ObjectIDs
- **`custom`**: No automatic ID generation; user must provide `_id` or MongoDB will auto-generate

🔝 [back to top](#api-reference)

&nbsp;

### Result Types

```go
type InsertOneResult struct {
    InsertedID  string    // ULID string
    GeneratedAt time.Time
}

type InsertManyResult struct {
    InsertedIDs   []string  // ULID strings
    InsertedCount int64
    GeneratedAt   time.Time
}

type UpdateResult struct {
    MatchedCount  int64
    ModifiedCount int64
    UpsertedID    string  // ULID string
    UpsertedCount int64
}

type DeleteResult struct {
    DeletedCount int64
}
```

🔝 [back to top](#api-reference)

&nbsp;

## Usage Patterns

&nbsp;

### Basic CRUD Operations

```go
client, _ := mongodb.NewClient()
collection := client.Collection("users")

// Create
result, err := collection.InsertOne(ctx, map[string]any{
    "name": "John Doe",
    "email": "john@example.com",
})

// Read
var user map[string]any
err = collection.FindOne(ctx, map[string]any{
    "email": "john@example.com",
}).Decode(&user)

// Update
_, err = collection.UpdateOne(ctx,
    map[string]any{"email": "john@example.com"},
    map[string]any{"$set": map[string]any{"status": "active"}},
)

// Delete
_, err = collection.DeleteOne(ctx, map[string]any{
    "email": "john@example.com",
})
```

🔝 [back to top](#api-reference)

&nbsp;

### Transaction Example

```go
client, _ := mongodb.NewClient()

err := client.WithTransaction(ctx, func(ctx context.Context) error {
    collection := client.Collection("accounts")

    // Transfer money between accounts
    _, err := collection.UpdateOne(ctx,
        map[string]any{"account_id": "from_account"},
        map[string]any{"$inc": map[string]any{"balance": -100}},
    )
    if err != nil {
        return err
    }

    _, err = collection.UpdateOne(ctx,
        map[string]any{"account_id": "to_account"},
        map[string]any{"$inc": map[string]any{"balance": 100}},
    )
    return err
})
```

🔝 [back to top](#api-reference)

&nbsp;

### Index Management

```go
collection := client.Collection("users")

// Create index
_, err := collection.CreateIndex(ctx, map[string]any{
    "email": 1,
}, &options.IndexOptions{
    Unique: true,
})

// Create compound index
_, err = collection.CreateIndex(ctx, map[string]any{
    "status": 1,
    "created_at": -1,
})

// Create text index
_, err = collection.CreateIndex(ctx, map[string]any{
    "name": "text",
    "description": "text",
})
```

🔝 [back to top](#api-reference)

&nbsp;

## Advanced Usage Patterns

&nbsp;

### Custom Configuration

```go
config := &mongodb.Config{
    Host:                 "mongodb.example.com",
    Port:                 27017,
    Database:             "production",
    Username:             "app_user",
    MaxPoolSize:          200,
    ConnectTimeout:       15 * time.Second,
    ReconnectEnabled:     true,
    HealthCheckEnabled:   true,
    CompressionEnabled:   true,
    CompressionAlgorithm: "zstd",
    ReadPreference:       "primaryPreferred",
    WriteConcern:         "majority",
    AppName:              "my-app-v2",
}

client, err := mongodb.NewClientWithConfig(config)
```

🔝 [back to top](#api-reference)

&nbsp;

### Aggregation Pipeline

```go
collection := client.Collection("orders")

pipeline := []map[string]any{
    {"$match": map[string]any{"status": "completed"}},
    {"$group": map[string]any{
        "_id": "$customer_id",
        "total_orders": map[string]any{"$sum": 1},
        "total_amount": map[string]any{"$sum": "$amount"},
    }},
    {"$sort": map[string]any{"total_amount": -1}},
    {"$limit": 10},
}

cursor, err := collection.Aggregate(ctx, pipeline)
defer cursor.Close(ctx)

var results []map[string]any
err = cursor.All(ctx, &results)
```

🔝 [back to top](#api-reference)

&nbsp;

## Best Practices

&nbsp;

### Error Handling

```go
result, err := collection.InsertOne(ctx, document)
if err != nil {
    if mongodb.IsDuplicateKeyError(err) {
        // Handle duplicate key error
        return fmt.Errorf("document already exists: %w", err)
    }
    if mongodb.IsTimeoutError(err) {
        // Handle timeout error
        return fmt.Errorf("operation timed out: %w", err)
    }
    return fmt.Errorf("failed to insert document: %w", err)
}
```

🔝 [back to top](#api-reference)

&nbsp;

### Context Usage

```go
// Use context with timeout for operations
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := collection.FindOne(ctx, filter).Decode(&document)
```

🔝 [back to top](#api-reference)

&nbsp;

### Resource Management

```go
client, err := mongodb.NewClient()
if err != nil {
    return err
}
defer func() {
    if err := client.Close(); err != nil {
        log.Printf("Failed to close client: %v", err)
    }
}()
```

🔝 [back to top](#api-reference)

&nbsp;

## Environment Variables

The following environment variables are supported for configuration:

### Connection Settings

- `MONGODB_HOST` - MongoDB host (default: localhost)
- `MONGODB_PORT` - MongoDB port (default: 27017)
- `MONGODB_USERNAME` - Username for authentication
- `MONGODB_PASSWORD` - Password for authentication
- `MONGODB_DATABASE` - Database name (default: app)
- `MONGODB_AUTH_DATABASE` - Authentication database (default: admin)
- `MONGODB_REPLICA_SET` - Replica set name

### ID Generation

- `MONGODB_ID_MODE` - ID generation strategy: `ulid`, `objectid`, or `custom` (default: ulid)

### Connection Pool

- `MONGODB_MAX_POOL_SIZE` - Maximum connection pool size (default: 100)
- `MONGODB_MIN_POOL_SIZE` - Minimum connection pool size (default: 5)
- `MONGODB_MAX_IDLE_TIME` - Maximum idle time (default: 5m)
- `MONGODB_MAX_CONN_IDLE_TIME` - Maximum connection idle time (default: 10m)

### Timeouts

- `MONGODB_CONNECT_TIMEOUT` - Connection timeout (default: 10s)
- `MONGODB_SERVER_SELECT_TIMEOUT` - Server selection timeout (default: 5s)
- `MONGODB_SOCKET_TIMEOUT` - Socket timeout (default: 10s)

### Reconnection

- `MONGODB_RECONNECT_ENABLED` - Enable auto-reconnection (default: true)
- `MONGODB_RECONNECT_DELAY` - Initial reconnection delay (default: 5s)
- `MONGODB_MAX_RECONNECT_DELAY` - Maximum reconnection delay (default: 1m)
- `MONGODB_RECONNECT_BACKOFF` - Backoff multiplier (default: 2.0)
- `MONGODB_MAX_RECONNECT_ATTEMPTS` - Maximum reconnection attempts (default: 10)

### Health Checks

- `MONGODB_HEALTH_CHECK_ENABLED` - Enable health checks (default: true)
- `MONGODB_HEALTH_CHECK_INTERVAL` - Health check interval (default: 30s)

### Performance

- `MONGODB_COMPRESSION_ENABLED` - Enable compression (default: true)
- `MONGODB_COMPRESSION_ALGORITHM` - Compression algorithm: `snappy`, `zlib`, `zstd` (default: snappy)
- `MONGODB_READ_PREFERENCE` - Read preference: `primary`, `primaryPreferred`, `secondary`, `secondaryPreferred`, `nearest` (default: primary)
- `MONGODB_WRITE_CONCERN` - Write concern (default: majority)
- `MONGODB_READ_CONCERN` - Read concern: `local`, `available`, `majority`, `linearizable` (default: local)

### Application

- `MONGODB_APP_NAME` - Application name (default: go-mongodb-app)
- `MONGODB_CONNECTION_NAME` - Connection name for identification

### Logging

- `MONGODB_LOG_LEVEL` - Log level: `debug`, `info`, `warn`, `error` (default: info)
- `MONGODB_LOG_FORMAT` - Log format: `json`, `text` (default: json)

🔝 [back to top](#api-reference)

&nbsp;

## Performance Considerations

- **Connection Pooling**: Configure `MaxPoolSize` and `MinPoolSize` based on your workload
- **Timeouts**: Set appropriate timeouts for your use case
- **Indexes**: Create indexes for frequently queried fields
- **Compression**: Enable compression for network-bound workloads
- **Read Preferences**: Use appropriate read preferences for your consistency requirements
- **ULID IDs**: Benefit from natural ordering and better database performance

🔝 [back to top](#api-reference)

&nbsp;

---

&nbsp;

An open source project brought to you by the [Cloudresty](https://cloudresty.com) team.

[Website](https://cloudresty.com) &nbsp;|&nbsp; [LinkedIn](https://www.linkedin.com/company/cloudresty) &nbsp;|&nbsp; [BlueSky](https://bsky.app/profile/cloudresty.com) &nbsp;|&nbsp; [GitHub](https://github.com/cloudresty)

&nbsp;
