# Examples

[Home](../README.md) &nbsp;/&nbsp; [Docs](README.md) &nbsp;/&nbsp; Examples

&nbsp;

This document provides comprehensive examples demonstrating different features of the go-mongodb package.

&nbsp;

## Available Examples

The package includes several examples in the `examples/` directory:

- **`basic-client/`** - Basic MongoDB client setup using environment variables
- **`env-config/`** - Environment variable configuration examples
- **`transactions/`** - Multi-document transactions
- **`ulid-demo/`** - ULID-based ObjectIDs with MongoDB
- **`reconnection-test/`** - Auto-reconnection behavior demonstration

🔝 [back to top](#examples)

&nbsp;

## Quick Start Examples

&nbsp;

### Basic Client Setup

```go
package main

import (
    "context"
    "os"

    "github.com/cloudresty/emit"
    "github.com/cloudresty/go-mongodb"
)

func main() {
    // Create client from environment variables
    client, err := mongodb.NewClient()
    if err != nil {
        emit.Error.StructuredFields("Failed to create client",
            emit.ZString("error", err.Error()))
        os.Exit(1)
    }
    defer client.Close()

    // Get database and collection
    db := client.Database("myapp")
    coll := db.Collection("users")

    emit.Info.Msg("MongoDB client connected successfully")
}
```

🔝 [back to top](#examples)

&nbsp;

### Basic CRUD Operations

```go
package main

import (
    "context"
    "os"

    "go.mongodb.org/mongo-driver/v2/bson"
    "github.com/cloudresty/emit"
    "github.com/cloudresty/go-mongodb"
)

type User struct {
    Name     string    `bson:"name" json:"name"`
    Email    string    `bson:"email" json:"email"`
    Active   bool      `bson:"active" json:"active"`
}

func main() {
    client, err := mongodb.NewClient()
    if err != nil {
        emit.Error.StructuredFields("Failed to create client",
            emit.ZString("error", err.Error()))
        os.Exit(1)
    }
    defer client.Close()

    coll := client.Collection("users")

    // Create - ULID is automatically generated
    user := User{
        Name:   "John Doe",
        Email:  "john@example.com",
        Active: true,
    }

    result, err := coll.InsertOne(context.Background(), user)
    if err != nil {
        emit.Error.StructuredFields("Failed to insert user",
            emit.ZString("error", err.Error()))
        os.Exit(1)
    }

    emit.Info.StructuredFields("Inserted user with ULID",
        emit.ZString("inserted_id", result.InsertedID.Hex()),
        emit.ZString("ulid", result.ULID))

    // Read - find by the generated ID
    var foundUser User
    err = coll.FindOne(context.Background(), bson.M{"_id": result.InsertedID}).Decode(&foundUser)
    if err != nil {
        emit.Error.StructuredFields("Failed to find user",
            emit.ZString("error", err.Error()))
        os.Exit(1)
    }

    // Update
    _, err = coll.UpdateOne(
        context.Background(),
        bson.M{"_id": result.InsertedID},
        bson.M{"$set": bson.M{"active": false}},
    )
    if err != nil {
        emit.Error.StructuredFields("Failed to update user",
            emit.ZString("error", err.Error()))
        os.Exit(1)
    }

    // Delete
    _, err = coll.DeleteOne(context.Background(), bson.M{"_id": result.InsertedID})
    if err != nil {
        emit.Error.StructuredFields("Failed to delete user",
            emit.ZString("error", err.Error()))
        os.Exit(1)
    }

    emit.Info.Msg("CRUD operations completed successfully")
}
```

🔝 [back to top](#examples)

&nbsp;

## Running Examples

To run the examples:

```bash
# Set up environment
export MONGODB_HOST=localhost
export MONGODB_PORT=27017
export MONGODB_DATABASE=myapp
export MONGODB_CONNECTION_NAME=example-client

# Run basic client example
go run examples/client/main.go

# Run CRUD example
go run examples/crud/main.go

# Run transaction example
go run examples/transactions/main.go
```

🔝 [back to top](#examples)

&nbsp;

## Custom Connection Names

```go
package main

import (
    "context"
    "os"

    "github.com/cloudresty/emit"
    "github.com/cloudresty/go-mongodb"
)

func main() {
    config, err := mongodb.LoadConfig()
    if err != nil {
        emit.Error.StructuredFields("Failed to load config",
            emit.ZString("error", err.Error()))
        os.Exit(1)
    }
    config.ConnectionName = "user-service-v1.2.3"
    config.Host = "localhost"
    config.Port = 27017
    config.Database = "userdb"

    client, err := mongodb.NewClientWithConfig(config)
    if err != nil {
        emit.Error.StructuredFields("Failed to create client",
            emit.ZString("error", err.Error()))
        os.Exit(1)
    }
    defer client.Close()

    emit.Info.StructuredFields("Connected with custom name",
        emit.ZString("connection_name", config.ConnectionName))
}
```

🔝 [back to top](#examples)

&nbsp;

## Production Configuration

&nbsp;

### With Environment Variables

```bash
# Production MongoDB setup
export MONGODB_HOST=prod-cluster.mongodb.net
export MONGODB_PORT=27017
export MONGODB_USERNAME=prod-user
export MONGODB_PASSWORD=secure-password
export MONGODB_DATABASE=production
export MONGODB_CONNECTION_NAME=production-service-v2.1.0
export MONGODB_MAX_POOL_SIZE=20
export MONGODB_MIN_POOL_SIZE=5
export MONGODB_MAX_IDLE_TIME=5m
export MONGODB_SERVER_SELECT_TIMEOUT=30s
export MONGODB_SOCKET_TIMEOUT=5m
```

🔝 [back to top](#examples)

&nbsp;

### With Custom Configuration

```go
package main

import (
    "context"
    "os"
    "time"

    "github.com/cloudresty/emit"
    "github.com/cloudresty/go-mongodb"
    "go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
    config, err := mongodb.LoadConfig()
    if err != nil {
        emit.Error.StructuredFields("Failed to load config",
            emit.ZString("error", err.Error()))
        os.Exit(1)
    }
    config.Host = "cluster.mongodb.net"
    config.Port = 27017
    config.Database = "production"
    config.ConnectionName = "production-service"

    // Production-ready settings
    config.MaxPoolSize = 20
    config.MinPoolSize = 5
    config.MaxIdleTime = 5 * time.Minute
    config.ServerSelectionTimeout = 30 * time.Second
    config.SocketTimeout = 5 * time.Minute
    config.HeartbeatFrequency = 10 * time.Second

    client, err := mongodb.NewClientWithConfig(config)
    if err != nil {
        emit.Error.StructuredFields("Failed to create production client",
            emit.ZString("error", err.Error()))
        os.Exit(1)
    }
    defer client.Close()

    emit.Info.Msg("Production client configured successfully")
}
```

🔝 [back to top](#examples)

&nbsp;

## Transaction Patterns

&nbsp;

### Multi-Document Transaction

```go
package main

import (
    "context"
    "os"

    "go.mongodb.org/mongo-driver/v2/bson"
    "go.mongodb.org/mongo-driver/v2/mongo"
    "go.mongodb.org/mongo-driver/v2/mongo/options"
    "go.mongodb.org/mongo-driver/v2/mongo/readconcern"
    "go.mongodb.org/mongo-driver/v2/mongo/writeconcern"
    "github.com/cloudresty/emit"
    "github.com/cloudresty/go-mongodb"
    "github.com/cloudresty/ulid"
)

func main() {
    client, err := mongodb.NewClient()
    if err != nil {
        emit.Error.StructuredFields("Failed to create client",
            emit.ZString("error", err.Error()))
        os.Exit(1)
    }
    defer client.Close()

    db := client.Database("ecommerce")
    orders := db.Collection("orders")
    inventory := db.Collection("inventory")

    // Start transaction
    session, err := client.StartSession()
    if err != nil {
        emit.Error.StructuredFields("Failed to start session",
            emit.ZString("error", err.Error()))
        os.Exit(1)
    }
    defer session.EndSession(context.Background())

    // Define transaction operation
    txnOpts := options.Transaction().SetReadConcern(readconcern.Majority()).
        SetWriteConcern(writeconcern.Majority())

    result, err := session.WithTransaction(context.Background(), func(sessCtx context.Context) (any, error) {
        // Create order
        order := bson.M{
            "_id":       bson.NewObjectID(),
            "product":   "laptop",
            "quantity":  1,
            "amount":    999.99,
            "status":    "pending",
        }

        _, err := orders.InsertOne(sessCtx, order)
        if err != nil {
            return nil, err
        }

        // Update inventory
        filter := bson.M{"product": "laptop"}
        update := bson.M{"$inc": bson.M{"quantity": -1}}

        updateResult, err := inventory.UpdateOne(sessCtx, filter, update)
        if err != nil {
            return nil, err
        }

        if updateResult.ModifiedCount == 0 {
            return nil, mongo.ErrNoDocuments
        }

        return order, nil
    }, txnOpts)

    if err != nil {
        emit.Error.StructuredFields("Transaction failed",
            emit.ZString("error", err.Error()))
        os.Exit(1)
    }

    emit.Info.StructuredFields("Transaction completed",
        emit.ZAny("result", result))
}
```

🔝 [back to top](#examples)

&nbsp;

### Bulk Operations

```go
package main

import (
    "context"
    "fmt"
    "os"

    "go.mongodb.org/mongo-driver/v2/bson"
    "go.mongodb.org/mongo-driver/v2/mongo"
    "github.com/cloudresty/emit"
    "github.com/cloudresty/go-mongodb"
    "github.com/cloudresty/ulid"
)

func main() {
    client, err := mongodb.NewClient()
    if err != nil {
        emit.Error.StructuredFields("Failed to create client",
            emit.ZString("error", err.Error()))
        os.Exit(1)
    }
    defer client.Close()

    coll := client.Database("myapp").Collection("users")

    // Bulk insert
    var documents []any
    for i := 0; i < 1000; i++ {
        documents = append(documents, bson.M{
            "_id":    bson.NewObjectID(),
            "name":   fmt.Sprintf("User %d", i),
            "email":  fmt.Sprintf("user%d@example.com", i),
            "active": true,
        })
    }

    insertResult, err := coll.InsertMany(context.Background(), documents)
    if err != nil {
        emit.Error.StructuredFields("Bulk insert failed",
            emit.ZString("error", err.Error()))
        os.Exit(1)
    }

    // Bulk write operations
    var operations []mongo.WriteModel

    // Add update operations
    for i := 0; i < 100; i++ {
        filter := bson.M{"name": fmt.Sprintf("User %d", i)}
        update := bson.M{"$set": bson.M{"active": false}}
        operations = append(operations, mongo.NewUpdateOneModel().
            SetFilter(filter).SetUpdate(update))
    }

    // Add delete operations
    for i := 900; i < 1000; i++ {
        filter := bson.M{"name": fmt.Sprintf("User %d", i)}
        operations = append(operations, mongo.NewDeleteOneModel().
            SetFilter(filter))
    }

    bulkResult, err := coll.BulkWrite(context.Background(), operations)
    if err != nil {
        emit.Error.StructuredFields("Bulk write failed",
            emit.ZString("error", err.Error()))
        os.Exit(1)
    }

    emit.Info.StructuredFields("Bulk operations completed",
        emit.ZInt("inserted", len(insertResult.InsertedIDs)),
        emit.ZInt64("updated", bulkResult.ModifiedCount),
        emit.ZInt64("deleted", bulkResult.DeletedCount))
}
```

🔝 [back to top](#examples)

&nbsp;

## Error Handling

&nbsp;

### Connection Errors

```go
package main

import (
    "context"
    "errors"
    "os"
    "time"

    "go.mongodb.org/mongo-driver/v2/mongo"
    "github.com/cloudresty/emit"
    "github.com/cloudresty/go-mongodb"
)

func main() {
    config, err := mongodb.LoadConfig()
    if err != nil {
        emit.Error.StructuredFields("Failed to load config",
            emit.ZString("error", err.Error()))
        os.Exit(1)
    }
    config.Host = "invalid-host"
    config.Port = 27017
    config.ServerSelectTimeout = 5 * time.Second

    client, err := mongodb.NewClientWithConfig(config)
    if err != nil {
        emit.Error.StructuredFields("Client creation failed",
            emit.ZString("error", err.Error()))

        // Handle specific connection errors
        if mongo.IsTimeout(err) {
            emit.Error.Msg("Connection timeout - check MongoDB server")
        } else if mongo.IsNetworkError(err) {
            emit.Error.Msg("Network error - check connectivity")
        }

        os.Exit(1)
    }
    defer client.Close()
}
```

🔝 [back to top](#examples)

&nbsp;

### Operation Errors

```go
package main

import (
    "context"
    "errors"
    "os"

    "go.mongodb.org/mongo-driver/v2/bson"
    "go.mongodb.org/mongo-driver/v2/mongo"
    "github.com/cloudresty/emit"
    "github.com/cloudresty/go-mongodb"
)

func main() {
    client, err := mongodb.NewClient()
    if err != nil {
        emit.Error.StructuredFields("Failed to create client",
            emit.ZString("error", err.Error()))
        os.Exit(1)
    }
    defer client.Close()

    coll := client.Database("myapp").Collection("users")

    // Handle not found errors
    var user bson.M
    err = coll.FindOne(context.Background(), bson.M{"_id": "nonexistent"}).Decode(&user)
    if err != nil {
        if errors.Is(err, mongo.ErrNoDocuments) {
            emit.Warn.Msg("User not found")
        } else {
            emit.Error.StructuredFields("Find operation failed",
                emit.ZString("error", err.Error()))
        }
    }

    // Handle duplicate key errors
    document := bson.M{"_id": "duplicate", "name": "Test"}
    _, err = coll.InsertOne(context.Background(), document)
    _, err = coll.InsertOne(context.Background(), document) // This will fail

    if err != nil {
        if mongo.IsDuplicateKeyError(err) {
            emit.Warn.Msg("Document already exists")
        } else {
            emit.Error.StructuredFields("Insert failed",
                emit.ZString("error", err.Error()))
        }
    }
}
```

🔝 [back to top](#examples)

&nbsp;

## Development and Testing

### Running Tests

```bash
# Unit tests only
make test

# Integration tests (requires MongoDB)
make test-integration

# Start MongoDB in Docker
make docker-mongodb

# Run specific example
go run examples/client/main.go
go run examples/crud/main.go
go run examples/transactions/main.go
```

🔝 [back to top](#examples)

&nbsp;

### Building Examples

```bash
# Build all examples
make build

# Format and lint
make fmt
make lint

# Full CI pipeline
make ci
```

🔝 [back to top](#examples)

&nbsp;

### Testing Best Practices

- **Use Docker for integration tests** - Consistent test environment
- **Mock connections for unit tests** - Fast, isolated testing
- **Test error scenarios and edge cases** - Ensure robust error handling
- **Test with realistic data volumes** - Performance validation
- **Test reconnection scenarios** - Network resilience validation

🔝 [back to top](#examples)

&nbsp;

For more detailed examples, see the [`examples/`](../examples/) directory in the repository.

&nbsp;

---

&nbsp;

An open source project brought to you by the [Cloudresty](https://cloudresty.com) team.

[Website](https://cloudresty.com) &nbsp;|&nbsp; [LinkedIn](https://www.linkedin.com/company/cloudresty) &nbsp;|&nbsp; [BlueSky](https://bsky.app/profile/cloudresty.com) &nbsp;|&nbsp; [GitHub](https://github.com/cloudresty)

&nbsp;
