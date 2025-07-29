# Examples

[Home](../README.md) &nbsp;/&nbsp; [Docs](README.md) &nbsp;/&nbsp; Examples

&nbsp;

This document provides comprehensive examples demonstrating different features of the go-mongodb package.

&nbsp;

## Available Examples

The package includes several examples in the `examples/` directory:

- **`basic-client/`** - Basic MongoDB client setup using environment variables
- **`custom-logger-emit/`** - Custom logging integration with the emit library
- **`env-config/`** - Environment variable configuration examples
- **`transactions/`** - Multi-document transactions
- **`ulid-demo/`** - ULID-based IDs with MongoDB
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
    "log"
    "os"

    "github.com/cloudresty/go-mongodb"
)

func main() {
    // Create client from environment variables - uses silent logging by default
    client, err := mongodb.NewClient(mongodb.FromEnv())
    if err != nil {
        log.Printf("Failed to create client: %v", err)
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

### Custom Logging Integration

The MongoDB client supports pluggable logging through the `Logger` interface. By default, the client is silent (uses `NopLogger`). Here's how to integrate with the `emit` logging library:

```go
package main

import (
    "fmt"
    "time"

    "github.com/cloudresty/emit"
    "github.com/cloudresty/go-mongodb"
)

// EmitAdapter adapts the emit logger to satisfy the mongodb.Logger interface
type EmitAdapter struct{}

func (e EmitAdapter) Info(msg string, fields ...any) {
    e.logWithFields(emit.Info.StructuredFields, emit.Info.Msg, msg, fields...)
}

func (e EmitAdapter) Warn(msg string, fields ...any) {
    e.logWithFields(emit.Warn.StructuredFields, emit.Warn.Msg, msg, fields...)
}

func (e EmitAdapter) Error(msg string, fields ...any) {
    e.logWithFields(emit.Error.StructuredFields, emit.Error.Msg, msg, fields...)
}

func (e EmitAdapter) Debug(msg string, fields ...any) {
    e.logWithFields(emit.Debug.StructuredFields, emit.Debug.Msg, msg, fields...)
}

func (e EmitAdapter) logWithFields(structuredLogger func(string, ...emit.ZField), msgLogger func(string), msg string, fields ...any) {
    if len(fields) == 0 {
        msgLogger(msg)
        return
    }

    emitFields := make([]emit.ZField, 0, len(fields)/2)
    for i := 0; i < len(fields)-1; i += 2 {
        key, ok := fields[i].(string)
        if !ok {
            continue
        }

        value := fields[i+1]
        switch v := value.(type) {
        case string:
            emitFields = append(emitFields, emit.ZString(key, v))
        case int:
            emitFields = append(emitFields, emit.ZInt(key, v))
        case time.Duration:
            emitFields = append(emitFields, emit.ZDuration(key, v))
        case error:
            emitFields = append(emitFields, emit.ZString(key, v.Error()))
        default:
            emitFields = append(emitFields, emit.ZString(key, fmt.Sprintf("%v", v)))
        }
    }

    structuredLogger(msg, emitFields...)
}

func main() {
    // Create MongoDB client with emit logger integration
    emitLogger := EmitAdapter{}

    client, err := mongodb.NewClient(
        mongodb.WithDatabase("example_db"),
        mongodb.WithLogger(emitLogger), // Inject our emit adapter
    )
    if err != nil {
        emit.Error.StructuredFields("Failed to create client", emit.ZString("error", err.Error()))
        return
    }
    defer client.Close()

    // All internal MongoDB operations will now use emit for logging
    emit.Info.Msg("MongoDB client created with emit logger integration")
}
```

**Note**: This approach allows you to integrate any logging library by implementing the `Logger` interface. The client operations will use your logger for internal logging while remaining completely decoupled from any specific logging framework.

🔝 [back to top](#examples)

&nbsp;

### Basic CRUD Operations

```go
package main

import (
    "context"
    "os"

    "github.com/cloudresty/emit"
    "github.com/cloudresty/go-mongodb"
    "github.com/cloudresty/go-mongodb/filter"
    "github.com/cloudresty/go-mongodb/update"
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
        emit.ZString("inserted_id", result.InsertedID.(string)))

    // Read - find by the generated ID using type-safe filter
    var foundUser User
    err = coll.FindOne(context.Background(),
        filter.Eq("_id", result.InsertedID)).Decode(&foundUser)
    if err != nil {
        emit.Error.StructuredFields("Failed to find user",
            emit.ZString("error", err.Error()))
        os.Exit(1)
    }

    // Update using type-safe update builder
    _, err = coll.UpdateOne(
        context.Background(),
        filter.Eq("_id", result.InsertedID),
        update.Set("active", false),
    )
    if err != nil {
        emit.Error.StructuredFields("Failed to update user",
            emit.ZString("error", err.Error()))
        os.Exit(1)
    }

    // Delete using type-safe filter
    _, err = coll.DeleteOne(context.Background(),
        filter.Eq("_id", result.InsertedID))
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

## Fluent Query Builders

&nbsp;

### Filter Builder Examples

The package includes a powerful, type-safe filter builder that supports fluent method chaining:

&nbsp;

#### Filter Method-Based Approach (Recommended)

```go
import "github.com/cloudresty/go-mongodb/filter"

// Build complex filters using fluent methods
complexFilter := filter.Eq("status", "active").
    And(filter.Gt("age", 25), filter.Lt("age", 65)).
    Or(filter.Eq("premium", true))

// Equivalent to: status = "active" AND (age > 25 AND age < 65) OR premium = true

// Use with collection methods
cursor, err := collection.Find(ctx, complexFilter)
```

🔝 [back to top](#examples)

&nbsp;

#### Filter Alternative Pattern

```go
// Start with filter.New() for more complex compositions
complexFilter := filter.New().
    And(
        filter.Eq("category", "electronics"),
        filter.Gte("price", 100),
    ).
    Or(filter.Eq("featured", true))

// Use in queries
count, err := collection.CountDocuments(ctx, complexFilter)
```

🔝 [back to top](#examples)

&nbsp;

### Update Builder Examples

Create complex update operations using the fluent update builder:

&nbsp;

#### Update Method-Based Approach (Recommended)

```go
import "github.com/cloudresty/go-mongodb/update"

// Build complex updates using fluent methods
updateOp := update.Set("name", "John Doe").
    Set("last_login", time.Now()).
    Inc("login_count", 1).
    Push("tags", "active")

// Use with collection update methods
result, err := collection.UpdateOne(ctx, filter.Eq("_id", userID), updateOp)
if err != nil {
    log.Fatal("Update failed:", err)
}

log.Printf("Modified %d documents", result.ModifiedCount)
```

🔝 [back to top](#examples)

&nbsp;

#### Update Alternative Pattern

```go
// Start with update.New() for more complex compositions
updateOp := update.New().
    Set("profile.email", "john@example.com").
    Set("profile.verified", true).
    AddToSet("roles", "user").
    Unset("temp_data")

// Execute the update
result, err := collection.UpdateMany(ctx, filter.Eq("status", "pending"), updateOp)
```

🔝 [back to top](#examples)

&nbsp;

## Running Examples

To run the examples:

```bash
# Set up environment
export MONGODB_HOSTS=localhost:27017
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
    // Create client with functional options - load from environment and override specific values
    client, err := mongodb.NewClient(
        mongodb.FromEnv(),                            // Load from environment
        mongodb.WithAppName("user-service-v1.2.3"),  // Override app name
        mongodb.WithHosts("localhost"),               // Override host
        mongodb.WithDatabase("userdb"),               // Override database
    )
    if err != nil {
        emit.Error.StructuredFields("Failed to create client",
            emit.ZString("error", err.Error()))
        os.Exit(1)
    }
    defer client.Close()

    emit.Info.StructuredFields("Connected with custom name",
        emit.ZString("connection_name", client.Name()))
}
```

🔝 [back to top](#examples)

&nbsp;

## Production Configuration

&nbsp;

### With Environment Variables

```bash
# Production MongoDB setup
export MONGODB_HOSTS=prod-cluster.mongodb.net:27017
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
    // Create production-ready client with functional options
    client, err := mongodb.NewClient(
        mongodb.FromEnv(),                                // Load from environment
        mongodb.WithHosts("cluster.mongodb.net"),         // Override host
        mongodb.WithDatabase("production"),               // Override database
        mongodb.WithAppName("production-service"),        // Override app name
        mongodb.WithMaxPoolSize(20),                      // Production pool settings
        mongodb.WithMinPoolSize(5),
        mongodb.WithMaxIdleTime(5*time.Minute),
        mongodb.WithServerSelectionTimeout(30*time.Second),
    )
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

    "go.mongodb.org/mongo-driver/v2/mongo"
    "go.mongodb.org/mongo-driver/v2/mongo/options"
    "go.mongodb.org/mongo-driver/v2/mongo/readconcern"
    "go.mongodb.org/mongo-driver/v2/mongo/writeconcern"
    "github.com/cloudresty/emit"
    "github.com/cloudresty/go-mongodb"
    "github.com/cloudresty/go-mongodb/filter"
    "github.com/cloudresty/go-mongodb/update"
)

type Order struct {
    Product  string  `bson:"product" json:"product"`
    Quantity int     `bson:"quantity" json:"quantity"`
    Amount   float64 `bson:"amount" json:"amount"`
    Status   string  `bson:"status" json:"status"`
}

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
        // Create order (ULID will be auto-generated)
        order := Order{
            Product:  "laptop",
            Quantity: 1,
            Amount:   999.99,
            Status:   "pending",
        }

        _, err := orders.InsertOne(sessCtx, order)
        if err != nil {
            return nil, err
        }

        // Update inventory using type-safe builders
        updateResult, err := inventory.UpdateOne(sessCtx,
            filter.Eq("product", "laptop"),
            update.Inc("quantity", -1))
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

    "go.mongodb.org/mongo-driver/v2/mongo"
    "github.com/cloudresty/emit"
    "github.com/cloudresty/go-mongodb"
    "github.com/cloudresty/go-mongodb/filter"
    "github.com/cloudresty/go-mongodb/update"
)

type BulkUser struct {
    Name   string `bson:"name" json:"name"`
    Email  string `bson:"email" json:"email"`
    Active bool   `bson:"active" json:"active"`
}

func main() {
    client, err := mongodb.NewClient()
    if err != nil {
        emit.Error.StructuredFields("Failed to create client",
            emit.ZString("error", err.Error()))
        os.Exit(1)
    }
    defer client.Close()

    coll := client.Database("myapp").Collection("users")

    // Bulk insert (ULIDs will be auto-generated)
    var documents []any
    for i := 0; i < 1000; i++ {
        documents = append(documents, BulkUser{
            Name:   fmt.Sprintf("User %d", i),
            Email:  fmt.Sprintf("user%d@example.com", i),
            Active: true,
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

    // Add update operations using type-safe builders
    for i := 0; i < 100; i++ {
        filterBuilder := filter.Eq("name", fmt.Sprintf("User %d", i))
        updateBuilder := update.Set("active", false)
        operations = append(operations, mongo.NewUpdateOneModel().
            SetFilter(filterBuilder).SetUpdate(updateBuilder))
    }

    // Add delete operations using type-safe builders
    for i := 900; i < 1000; i++ {
        filterBuilder := filter.Eq("name", fmt.Sprintf("User %d", i))
        operations = append(operations, mongo.NewDeleteOneModel().
            SetFilter(filterBuilder))
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
    // Create client with invalid host to demonstrate error handling
    client, err := mongodb.NewClient(
        mongodb.FromEnv(),                              // Load from environment
        mongodb.WithHosts("invalid-host"),              // Override with invalid host
        mongodb.WithServerSelectionTimeout(5*time.Second), // Short timeout for demo
    )
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

    "go.mongodb.org/mongo-driver/v2/mongo"
    "github.com/cloudresty/emit"
    "github.com/cloudresty/go-mongodb"
    "github.com/cloudresty/go-mongodb/filter"
)

type User struct {
    Name   string `bson:"name" json:"name"`
    Email  string `bson:"email" json:"email"`
    Active bool   `bson:"active" json:"active"`
}

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
    var user User
    err = coll.FindOne(context.Background(),
        filter.Eq("_id", "nonexistent")).Decode(&user)
    if err != nil {
        if errors.Is(err, mongo.ErrNoDocuments) {
            emit.Warn.Msg("User not found")
        } else {
            emit.Error.StructuredFields("Find operation failed",
                emit.ZString("error", err.Error()))
        }
    }

    // Handle duplicate key errors
    document := User{Name: "Test User", Email: "test@example.com", Active: true}
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

&nbsp;

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

## ID Generation Strategies

The package supports multiple ID generation strategies for document `_id` fields.

&nbsp;

### ULID IDs (Default)

ULID provides temporal ordering and better database performance:

```go
package main

import (
    "context"
    "fmt"

    "github.com/cloudresty/go-mongodb"
)

type User struct {
    Name  string `bson:"name" json:"name"`
    Email string `bson:"email" json:"email"`
    Age   int    `bson:"age" json:"age"`
}

func main() {
    // Default behavior uses ULID
    client, err := mongodb.NewClient()
    if err != nil {
        panic(err)
    }
    defer client.Close()

    collection := client.Collection("users")

    // Insert document with automatic ULID generation
    user := User{
        Name:  "Alice Smith",
        Email: "alice@example.com",
        Age:   30,
    }

    result, err := collection.InsertOne(context.Background(), user)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Inserted document with ULID: %s\n", result.InsertedID)
    // Output: Inserted document with ULID: 01ARZ3NDEKTSV4RRFFQ69G5FAV
}
```

🔝 [back to top](#examples)

&nbsp;

### ObjectID Mode

For compatibility with existing MongoDB applications:

```go
package main

import (
    "context"
    "fmt"

    "github.com/cloudresty/go-mongodb"
)

type Product struct {
    Name     string  `bson:"name" json:"name"`
    Price    float64 `bson:"price" json:"price"`
    Category string  `bson:"category" json:"category"`
    InStock  bool    `bson:"in_stock" json:"in_stock"`
}

func main() {
    // Configure client to use ObjectID via environment or functional options
    client, err := mongodb.NewClient(
        mongodb.FromEnv(),
        mongodb.WithIDMode(mongodb.IDModeObjectID),
    )
    if err != nil {
        panic(err)
    }
    defer client.Close()

    collection := client.Collection("products")

    // Insert document with automatic ObjectID generation
    product := Product{
        Name:     "Widget Pro",
        Price:    29.99,
        Category: "electronics",
        InStock:  true,
    }

    result, err := collection.InsertOne(context.Background(), product)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Inserted document with ObjectID: %s\n", result.InsertedID)
    // Output: Inserted document with ObjectID: 507f1f77bcf86cd799439011
}
```

🔝 [back to top](#examples)

&nbsp;

### User-Provided IDs

Let users control their own ID generation:

```go
package main

import (
    "context"
    "fmt"

    "github.com/cloudresty/go-mongodb"
    "go.mongodb.org/mongo-driver/v2/bson/primitive"
)

type Order struct {
    ID         string  `bson:"_id" json:"id"`
    CustomerID string  `bson:"customer_id" json:"customer_id"`
    Total      float64 `bson:"total" json:"total"`
    Status     string  `bson:"status" json:"status"`
}

func main() {
    // Configure client to not generate IDs automatically
    client, err := mongodb.NewClient(
        mongodb.FromEnv(),
        mongodb.WithIDMode(mongodb.IDModeCustom),
    )
    if err != nil {
        panic(err)
    }
    defer client.Close()

    collection := client.Collection("orders")

    // User provides their own ID
    order := Order{
        ID:         "order-2023-12-001",
        CustomerID: "customer-456",
        Total:      149.99,
        Status:     "pending",
    }

    result, err := collection.InsertOne(context.Background(), order)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Inserted document with custom ID: %s\n", result.InsertedID)
    // Output: Inserted document with custom ID: order-2023-12-001
}
```

🔝 [back to top](#examples)

&nbsp;

### Environment-Based Configuration

Configure ID mode via environment variables:

```bash
# Use ObjectID for production
export MONGODB_ID_MODE=objectid
export MONGODB_HOSTS=prod.mongodb.com:27017
export MONGODB_DATABASE=production

# Use ULID for development (default)
export MONGODB_ID_MODE=ulid
export MONGODB_HOSTS=localhost:27017
export MONGODB_DATABASE=development
```

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/cloudresty/go-mongodb"
)

type AnalyticsEvent struct {
    Event     string `bson:"event" json:"event"`
    UserID    string `bson:"user_id" json:"user_id"`
    Timestamp string `bson:"timestamp" json:"timestamp"`
    IP        string `bson:"ip" json:"ip"`
}

func main() {
    // Client automatically uses MONGODB_ID_MODE environment variable
    client, err := mongodb.NewClient(mongodb.FromEnv())
    if err != nil {
        panic(err)
    }
    defer client.Close()

    collection := client.Collection("analytics")

    // ID generation strategy depends on environment
    event := AnalyticsEvent{
        Event:     "user_login",
        UserID:    "user-789",
        Timestamp: "2023-12-01T10:30:00Z",
        IP:        "192.168.1.100",
    }

    result, err := collection.InsertOne(context.Background(), event)
    if err != nil {
        panic(err)
    }

    idMode := os.Getenv("MONGODB_ID_MODE")
    if idMode == "" {
        idMode = "ulid" // default
    }

    fmt.Printf("Inserted document (%s mode): %s\n", idMode, result.InsertedID)
}
```

🔝 [back to top](#examples)

&nbsp;

### Mixed ID Strategies

You can override the client's ID mode by providing your own `_id`:

```go
package main

import (
    "context"
    "fmt"

    "github.com/cloudresty/go-mongodb"
    "go.mongodb.org/mongo-driver/v2/bson/primitive"
)

type MixedDocument struct {
    ID   any    `bson:"_id,omitempty" json:"id,omitempty"`
    Type string `bson:"type" json:"type"`
    Data string `bson:"data" json:"data"`
}

func main() {
    // Client configured for ULID (default)
    client, err := mongodb.NewClient()
    if err != nil {
        panic(err)
    }
    defer client.Close()

    collection := client.Collection("mixed_collection")

    // Document 1: Automatic ULID generation
    doc1 := MixedDocument{
        Type: "auto_ulid",
        Data: "Generated ULID",
    }
    result1, err := collection.InsertOne(context.Background(), doc1)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Auto ULID: %s\n", result1.InsertedID)

    // Document 2: User-provided string ID
    doc2 := MixedDocument{
        ID:   "custom-string-id-123",
        Type: "custom_string",
        Data: "Custom string ID",
    }
    result2, err := collection.InsertOne(context.Background(), doc2)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Custom string: %s\n", result2.InsertedID)

    // Document 3: User-provided ObjectID
    customObjectID := primitive.NewObjectID()
    doc3 := MixedDocument{
        ID:   customObjectID,
        Type: "custom_objectid",
        Data: "Custom ObjectID",
    }
    result3, err := collection.InsertOne(context.Background(), doc3)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Custom ObjectID: %s\n", result3.InsertedID)
}
```

🔝 [back to top](#examples)

&nbsp;

## Testing Best Practices

- **Use Docker for integration tests** - Consistent test environment
- **Mock connections for unit tests** - Fast, isolated testing
- **Test error scenarios and edge cases** - Ensure robust error handling
- **Test with realistic data volumes** - Performance validation
- **Test reconnection scenarios** - Network resilience validation

🔝 [back to top](#examples)

&nbsp;

For more detailed examples, see the [`examples/`](../examples/) directory in the repository.

🔝 [back to top](#examples)

&nbsp;

---

&nbsp;

An open source project brought to you by the [Cloudresty](https://cloudresty.com) team.

[Website](https://cloudresty.com) &nbsp;|&nbsp; [LinkedIn](https://www.linkedin.com/company/cloudresty) &nbsp;|&nbsp; [BlueSky](https://bsky.app/profile/cloudresty.com) &nbsp;|&nbsp; [GitHub](https://github.com/cloudresty) &nbsp;|&nbsp; [Docker Hub](https://hub.docker.com/u/cloudresty)

&nbsp;
