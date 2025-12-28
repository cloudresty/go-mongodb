# Getting Started with go-mongodb

[Home](../README.md) &nbsp;/&nbsp; [Docs](README.md) &nbsp;/&nbsp; Getting Started

&nbsp;

This guide will help you get up and running with the go-mongodb package quickly using the modern, fluent API with type-safe builders.

&nbsp;

## Prerequisites

- Go 1.24 or later
- MongoDB 8+ running locally or remotely
- Basic understanding of Go and MongoDB concepts

&nbsp;

🔝 [back to top](#getting-started-with-go-mongodb)

&nbsp;

## Quick Installation

```bash
go mod init your-project
go get github.com/cloudresty/go-mongodb
```

&nbsp;

🔝 [back to top](#getting-started-with-go-mongodb)

&nbsp;

## Basic Setup

&nbsp;

### 1. Environment Variables (Recommended)

Create a `.env` file in your project root:

```bash
MONGODB_HOSTS=localhost:27017
MONGODB_DATABASE=myapp
MONGODB_APP_NAME=my-app
```

&nbsp;

🔝 [back to top](#getting-started-with-go-mongodb)

&nbsp;

### 2. Basic Client Creation

```go
package main

import (
    "context"
    "log"

    "github.com/cloudresty/go-mongodb"
)

func main() {
    // Create client from environment variables
    client, err := mongodb.NewClient(mongodb.FromEnv())
    if err != nil {
        log.Fatal("Failed to create client:", err)
    }
    defer client.Close()

    // Test the connection
    ctx := context.Background()
    if err := client.Ping(ctx); err != nil {
        log.Fatal("Failed to ping MongoDB:", err)
    }

    log.Println("✓ Connected to MongoDB successfully!")
}
```

&nbsp;

🔝 [back to top](#getting-started-with-go-mongodb)

&nbsp;

## Your First Document

&nbsp;

### Index a Document

```go
import "github.com/cloudresty/go-mongodb/mongoid"

// Define your document structure
type User struct {
    ID       string    `bson:"_id"`
    Name     string    `bson:"name"`
    Email    string    `bson:"email"`
    Age      int       `bson:"age"`
    Active   bool      `bson:"active"`
    CreatedAt time.Time `bson:"created_at"`
}

// Create and insert a document
user := User{
    ID:        mongoid.NewULID(),
    Name:      "John Doe",
    Email:     "john@example.com",
    Age:       30,
    Active:    true,
    CreatedAt: time.Now(),
}

// Use the resource-oriented API
users := client.Database("myapp").Collection("users")
result, err := users.InsertOne(ctx, user)
if err != nil {
    log.Fatal("Failed to insert document:", err)
}

log.Printf("Document inserted with ID: %s", result.InsertedID)
```

&nbsp;

🔝 [back to top](#getting-started-with-go-mongodb)

&nbsp;

### Search Documents

```go
import (
    "github.com/cloudresty/go-mongodb/filter"
    "github.com/cloudresty/go-mongodb/update"
)

// Create a type-safe filter using the fluent filter builder
searchFilter := filter.Eq("active", true).
    And(filter.Gt("age", 25))

// Find documents with the filter
cursor, err := users.Find(ctx, searchFilter)
if err != nil {
    log.Fatal("Search failed:", err)
}
defer cursor.Close(ctx)

// Process results - strongly typed!
var foundUsers []User
err = cursor.All(ctx, &foundUsers)
if err != nil {
    log.Fatal("Failed to decode users:", err)
}

log.Printf("Found %d active users over 25", len(foundUsers))
for _, user := range foundUsers {
    log.Printf("User: %s (%s) - Age: %d",
        user.Name, user.Email, user.Age)
}
```

&nbsp;

🔝 [back to top](#getting-started-with-go-mongodb)

&nbsp;

## Configuration Options

&nbsp;

### Using Environment Variables

The package supports extensive configuration through environment variables:

```bash
# Connection (hosts with ports required)
MONGODB_HOSTS=localhost:27017,localhost:27018
MONGODB_USERNAME=your-username
MONGODB_PASSWORD=your-password

# Database and App Settings
MONGODB_DATABASE=production
MONGODB_APP_NAME=my-production-app

# Performance and Connection Pool
MONGODB_MAX_POOL_SIZE=100
MONGODB_MIN_POOL_SIZE=5
MONGODB_TIMEOUT=30s

# ID Generation
MONGODB_ID_MODE=ulid  # or 'objectid' or 'custom'

# Replica Set and Advanced
MONGODB_REPLICA_SET=rs0
MONGODB_READ_PREFERENCE=secondaryPreferred
MONGODB_WRITE_CONCERN=majority
```

&nbsp;

🔝 [back to top](#getting-started-with-go-mongodb)

&nbsp;

### Using Configuration Struct

```go
import "time"

config := &mongodb.Config{
    Host:         "mongodb.example.com",
    Port:         27017,
    Username:     "myuser",
    Password:     "mypassword",
    Database:     "production",
    AppName:      "my-production-app",
    IDMode:       mongodb.IDModeULID,
    MaxPoolSize:  100,
    MinPoolSize:  10,
    Timeout:      30 * time.Second,
}

client, err := mongodb.NewClient(mongodb.WithConfig(config))
```

&nbsp;

🔝 [back to top](#getting-started-with-go-mongodb)

&nbsp;

## Modern Query Experience

&nbsp;

### Fluent Filter Builder

The package includes a powerful, type-safe filter builder in a dedicated `filter` sub-package:

```go
import "github.com/cloudresty/go-mongodb/filter"

// Build complex filters with a fluent API
complexFilter := filter.Eq("status", "active").
    And(
        filter.Gt("age", 21).Or(filter.In("role", "admin", "moderator")),
        filter.Regex("name", "^John", "i"),
    )

// You can also build single filters directly
termFilter := filter.Eq("category", "electronics")
rangeFilter := filter.Gte("price", 10).
    And(filter.Lte("price", 100))

arrayFilter := filter.In("tags", "golang", "mongodb", "database")
```

&nbsp;

🔝 [back to top](#getting-started-with-go-mongodb)

&nbsp;

### Fluent Update Builder

Get strongly-typed updates without manual BSON construction:

```go
import "github.com/cloudresty/go-mongodb/update"

// Build complex updates with a fluent API
complexUpdate := update.New().
    Set("last_login", time.Now()).
    Inc("login_count", 1).
    Push("activity_log", map[string]any{
        "action":    "login",
        "timestamp": time.Now(),
    }).
    Unset("temp_session_data")

// Execute the update
result, err := users.UpdateOne(ctx,
    filter.Eq("email", "john@example.com"),
    complexUpdate)

if err != nil {
    log.Fatal("Update failed:", err)
}

log.Printf("Modified %d documents", result.ModifiedCount)
```

&nbsp;

🔝 [back to top](#getting-started-with-go-mongodb)

&nbsp;

## ID Generation Strategies

&nbsp;

### ULID (Recommended)

```bash
export MONGODB_ID_MODE=ulid
```

```go
import "github.com/cloudresty/go-mongodb/mongoid"

// ULID provides time-ordered, globally unique IDs
user := User{
    ID:   mongoid.NewULID(), // e.g., "01ARZ3NDEKTSV4RRFFQ69G5FAV"
    Name: "John Doe",
}
result, _ := users.InsertOne(ctx, user)
// result.InsertedID will be the ULID string
```

&nbsp;

🔝 [back to top](#getting-started-with-go-mongodb)

&nbsp;

### MongoDB ObjectID

```bash
export MONGODB_ID_MODE=objectid
```

```go
import "github.com/cloudresty/go-mongodb/mongoid"

// ObjectID is MongoDB's default - optimal for sharding
user := User{
    ID:   mongoid.NewObjectID().Hex(), // e.g., "507f1f77bcf86cd799439011"
    Name: "John Doe",
}
result, _ := users.InsertOne(ctx, user)
```

&nbsp;

🔝 [back to top](#getting-started-with-go-mongodb)

&nbsp;

### Custom IDs

```bash
export MONGODB_ID_MODE=custom
```

```go
// Provide your own ID generation logic
user := User{
    ID:   "custom-user-123", // Your custom ID scheme
    Name: "John Doe",
}
result, _ := users.InsertOne(ctx, user)
```

&nbsp;

🔝 [back to top](#getting-started-with-go-mongodb)

&nbsp;

## Production Features

&nbsp;

### Health Checks

```go
import "time"

// Test basic connectivity
if err := client.Ping(ctx); err == nil {
    log.Println("✓ MongoDB is reachable")
}

// Get detailed client statistics
stats := client.Stats()
log.Printf("Active connections: %d", stats.ActiveConnections)
log.Printf("Operations executed: %d", stats.OperationsExecuted)
log.Printf("Reconnect attempts: %d", stats.ReconnectAttempts)
```

&nbsp;

🔝 [back to top](#getting-started-with-go-mongodb)

&nbsp;

### Transactions

```go
// Execute multiple operations atomically
err := client.WithTransaction(ctx, func(ctx context.Context) error {
    users := client.Database("myapp").Collection("users")
    orders := client.Database("myapp").Collection("orders")

    // Insert user
    userResult, err := users.InsertOne(ctx, newUser)
    if err != nil {
        return err
    }

    // Insert order for that user
    order.UserID = userResult.InsertedID
    _, err = orders.InsertOne(ctx, order)
    return err
})

if err != nil {
    log.Printf("Transaction failed: %v", err)
} else {
    log.Println("Transaction completed successfully")
}
```

&nbsp;

🔝 [back to top](#getting-started-with-go-mongodb)

&nbsp;

### Graceful Shutdown

```go
import (
    "os"
    "os/signal"
    "syscall"
)

// Handle graceful shutdown
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

go func() {
    <-sigChan
    log.Println("Shutting down gracefully...")
    if err := client.Close(); err != nil {
        log.Printf("Error closing client: %v", err)
    }
    os.Exit(0)
}()
```

&nbsp;

🔝 [back to top](#getting-started-with-go-mongodb)

&nbsp;

## Common Patterns

&nbsp;

### Error Handling

```go
result, err := users.InsertOne(ctx, user)
if err != nil {
    switch {
    case mongodb.IsDuplicateKeyError(err):
        log.Println("User already exists")
    case mongodb.IsConnectionError(err):
        log.Println("Connection error - check MongoDB availability")
    case mongodb.IsValidationError(err):
        log.Println("Validation error - check document structure")
    default:
        log.Printf("Unexpected error: %v", err)
    }
    return
}
```

&nbsp;

🔝 [back to top](#getting-started-with-go-mongodb)

&nbsp;

### Bulk Operations

```go
// Efficient batch operations
documents := []User{
    {ID: mongoid.NewULID(), Name: "Alice", Email: "alice@example.com"},
    {ID: mongoid.NewULID(), Name: "Bob", Email: "bob@example.com"},
    {ID: mongoid.NewULID(), Name: "Charlie", Email: "charlie@example.com"},
}

// Use InsertMany for efficient bulk inserts
result, err := users.InsertMany(ctx, documents)
if err != nil {
    log.Printf("Bulk insert failed: %v", err)
    return
}

log.Printf("Inserted %d documents", len(result.InsertedIDs))
```

&nbsp;

🔝 [back to top](#getting-started-with-go-mongodb)

&nbsp;

### Advanced Queries with Pipeline

```go
import "github.com/cloudresty/go-mongodb/pipeline"

// Complex aggregation with pipeline builder
p := pipeline.New().
    Match(filter.Eq("status", "active")).
    Group(
        bson.M{"category": "$category"},
        map[string]any{
            "count": bson.M{"$sum": 1},
            "avg_age": bson.M{"$avg": "$age"},
        },
    ).
    Sort(map[string]int{"count": -1}).
    Limit(10)

cursor, err := users.Aggregate(ctx, p)
if err != nil {
    log.Fatal("Aggregation failed:", err)
}
defer cursor.Close(ctx)

var results []CategoryStats
err = cursor.All(ctx, &results)
```

&nbsp;

🔝 [back to top](#getting-started-with-go-mongodb)

&nbsp;

## Resource-Oriented API

&nbsp;

### Document Operations

```go
// Get the resource-oriented services
db := client.Database("myapp")
users := db.Collection("users")

// Create a document
result, err := users.InsertOne(ctx, user)

// Get a document
var foundUser User
err = users.FindOne(ctx, filter.Eq("_id", result.InsertedID)).Decode(&foundUser)

// Update a document
updateResult, err := users.UpdateOne(ctx,
    filter.Eq("_id", result.InsertedID),
    update.Set("last_seen", time.Now()))

// Delete a document
deleteResult, err := users.DeleteOne(ctx, filter.Eq("_id", result.InsertedID))
```

&nbsp;

🔝 [back to top](#getting-started-with-go-mongodb)

&nbsp;

### Database Operations

```go
// Get database handle
db := client.Database("myapp")

// Database information
log.Printf("Database name: %s", db.Name())

// List collections
collections, err := db.ListCollections(ctx)
if err == nil {
    log.Printf("Collections: %v", collections)
}

// Drop database (careful!)
// err = db.Drop(ctx)
```

&nbsp;

🔝 [back to top](#getting-started-with-go-mongodb)

&nbsp;

## Next Steps

1. Read the [API Reference](api-reference.md) - Complete API documentation
2. Explore Examples - Check the `examples/` directory for comprehensive demos
3. Review [Production Features](production-features.md) - Production deployment guidance
4. Configure [Environment Variables](environment-variables.md) - All supported variables
5. Learn [Environment Configuration](environment-configuration.md) - Setup patterns and examples

&nbsp;

🔝 [back to top](#getting-started-with-go-mongodb)

&nbsp;

## Common Issues

&nbsp;

### Connection Refused

```bash
# Check if MongoDB is running
mongosh --eval "db.runCommand('ping')"

# Or start MongoDB locally
mongod --dbpath /usr/local/var/mongodb

# Or use Docker
docker run -d -p 27017:27017 mongo:latest
```

&nbsp;

🔝 [back to top](#getting-started-with-go-mongodb)

&nbsp;

### Authentication Errors

```bash
# Set credentials
export MONGODB_USERNAME=myuser
export MONGODB_PASSWORD=mypassword

# Or use connection string
export MONGODB_HOSTS=mongodb+srv://user:pass@cluster.mongodb.net
```

&nbsp;

🔝 [back to top](#getting-started-with-go-mongodb)

&nbsp;

### Collection Not Found

```go
// MongoDB creates collections automatically on first write
user := User{Name: "John", Email: "john@example.com"}
_, err := users.InsertOne(ctx, user)
// Collection "users" is created automatically
```

&nbsp;

🔝 [back to top](#getting-started-with-go-mongodb)

&nbsp;

## Getting Help

- **Documentation**: Check the [docs/](README.md) directory
- **Examples**: Run examples with `go run examples/<example-name>/main.go`
- **Issues**: Report issues on [GitHub](https://github.com/cloudresty/go-mongodb/issues)
- **Community**: Join discussions on [GitHub Discussions](https://github.com/cloudresty/go-mongodb/discussions)

&nbsp;

🔝 [back to top](#getting-started-with-go-mongodb)

&nbsp;

&nbsp;

---

### Cloudresty

[Website](https://cloudresty.com) &nbsp;|&nbsp; [LinkedIn](https://www.linkedin.com/company/cloudresty) &nbsp;|&nbsp; [BlueSky](https://bsky.app/profile/cloudresty.com) &nbsp;|&nbsp; [GitHub](https://github.com/cloudresty) &nbsp;|&nbsp; [Docker Hub](https://hub.docker.com/u/cloudresty)

<sub>&copy; Cloudresty</sub>

&nbsp;
