# API Reference

[Home](../README.md) &nbsp;/&nbsp; [Docs](README.md) &nbsp;/&nbsp; API Reference

&nbsp;

This document provides the definitive API reference for the `go-mongodb` package. It is designed to be your primary resource for mastering the entire library, from client creation and environment configuration to advanced queries, CRUD operations, and production features.

We've crafted this API to be powerful, consistent, and idiomatically Go. You'll discover how our resource-oriented services, fluent query builders, and type-safe operations can help you write cleaner, more reliable, and more productive MongoDB applications.

&nbsp;

## Core Functions

&nbsp;

### Client Creation

| Function | Description |
|----------|-------------|
| `NewClient(options...)` | Creates a client with functional options (use `FromEnv()` to load from environment variables) |

🔝 [back to top](#api-reference)

&nbsp;

#### Client Options

| Option | Description |
|--------|-------------|
| `WithHosts(hosts ...string)` | Sets custom hosts (overrides environment) |
| `WithCredentials(username, password string)` | Sets username and password for authentication (overrides environment) |
| `WithDatabase(name string)` | Sets default database (overrides environment) |
| `WithAppName(name string)` | Sets application name for logging and identification |
| `WithConnectionName(name string)` | Sets local client identifier for application logging |
| `WithMaxPoolSize(size int)` | Sets maximum connection pool size |
| `WithMinPoolSize(size int)` | Sets minimum connection pool size |
| `WithTimeout(duration time.Duration)` | Sets default operation timeout |
| `WithReplicaSet(name string)` | Sets replica set name |
| `WithDirectConnection(enabled bool)` | Enables direct connection mode (bypasses replica set discovery) |
| `WithTLS(enabled bool)` | Enables or disables TLS |
| `WithLogger(logger Logger)` | Sets a custom logger implementation (defaults to NopLogger - silent) |

🔝 [back to top](#api-reference)

&nbsp;

### Logger Interface

The `Logger` interface allows you to integrate your preferred logging solution with the MongoDB client for internal operation logging.

```go
type Logger interface {
    Info(msg string, fields ...any)
    Warn(msg string, fields ...any)
    Error(msg string, fields ...any)
    Debug(msg string, fields ...any)
}
```

**Default behavior**: If no logger is provided via `WithLogger()`, the client uses `NopLogger` (silent - no output).

**Usage patterns**:

- Fields are provided as alternating key-value pairs: `logger.Info("message", "key1", value1, "key2", value2)`
- Supported field types: `string`, `int`, `int64`, `time.Duration`, `bool`, `error`
- See [Custom Logging Example](../examples/custom-logger-emit/) for `emit` library integration

🔝 [back to top](#api-reference)

&nbsp;

### Connection Operations

| Function | Description |
|----------|-------------|
| `client.Ping(ctx context.Context) error` | Test connection and update internal state |
| `client.Stats() *ClientStats` | Get connection statistics (reconnect count, operations, etc.) |
| `client.Name() string` | Get the connection name for this client instance |
| `client.Close() error` | Close the client and stop background routines |

🔝 [back to top](#api-reference)

&nbsp;

## Environment Configuration

| Function | Description |
|----------|-------------|
| `FromEnv() Option` | Load configuration from `MONGODB_*` environment variables (functional option) |
| `FromEnvWithPrefix(prefix string) Option` | Load configuration with custom prefix (e.g., `MYAPP_MONGODB_*`) (functional option) |

🔝 [back to top](#api-reference)

&nbsp;

## Resource-Oriented API

&nbsp;

### Database Operations

| Function | Description |
|----------|-------------|
| `client.Database(name string) *Database` | Get a database handle for the specified name |
| `database.Name() string` | Get the database name |
| `database.Collection(name string) *Collection` | Get a collection handle for the specified name |
| `database.Drop(ctx context.Context) error` | Drop the database |
| `database.ListCollections(ctx context.Context) ([]string, error)` | List all collections in the database |

🔝 [back to top](#api-reference)

&nbsp;

### Collection Operations

| Function | Description |
|----------|-------------|
| `collection.Name() string` | Get the collection name |
| `collection.Database() *Database` | Get the parent database |

🔝 [back to top](#api-reference)

&nbsp;

## Document Operations

&nbsp;

### Basic CRUD Operations

| Function | Description |
|----------|-------------|
| `collection.InsertOne(ctx, document) (*InsertOneResult, error)` | Insert a single document |
| `collection.InsertMany(ctx, documents) (*InsertManyResult, error)` | Insert multiple documents |
| `collection.FindOne(ctx, filter) *FindOneResult` | Find a single document |
| `collection.Find(ctx, filter, opts...) (*Cursor, error)` | Find multiple documents |
| `collection.UpdateOne(ctx, filter, update) (*UpdateResult, error)` | Update a single document |
| `collection.UpdateMany(ctx, filter, update) (*UpdateResult, error)` | Update multiple documents |
| `collection.ReplaceOne(ctx, filter, replacement) (*UpdateResult, error)` | Replace a single document |
| `collection.DeleteOne(ctx, filter) (*DeleteResult, error)` | Delete a single document |
| `collection.DeleteMany(ctx, filter) (*DeleteResult, error)` | Delete multiple documents |
| `collection.CountDocuments(ctx, filter) (int64, error)` | Count documents matching filter |

🔝 [back to top](#api-reference)

&nbsp;

### Advanced Operations

| Function | Description |
|----------|-------------|
| `collection.Aggregate(ctx, pipeline) (*Cursor, error)` | Run aggregation pipeline |
| `collection.Distinct(ctx, field, filter) ([]any, error)` | Get distinct values for a field |
| `collection.Watch(ctx, pipeline, opts...) (*ChangeStream, error)` | Watch for changes |

🔝 [back to top](#api-reference)

&nbsp;

## Fluent Query Builders

&nbsp;

### Filter Builder (package `filter`)

| Function | Description |
|----------|-------------|
| `filter.New()` | Create a new filter builder |
| `filter.Eq(field, value)` | Create an equality filter |
| `filter.Ne(field, value)` | Create a not-equal filter |
| `filter.Gt(field, value)` | Create a greater-than filter |
| `filter.Gte(field, value)` | Create a greater-than-or-equal filter |
| `filter.Lt(field, value)` | Create a less-than filter |
| `filter.Lte(field, value)` | Create a less-than-or-equal filter |
| `filter.In(field, values...)` | Create an in filter |
| `filter.Nin(field, values...)` | Create a not-in filter |

🔝 [back to top](#api-reference)

&nbsp;

#### Logical Operations (Fluent Methods)

| Method | Description |
|--------|-------------|
| `builder.And(filters...)` | Combine filters with logical AND (fluent method) |
| `builder.Or(filters...)` | Combine filters with logical OR (fluent method) |
| `builder.Not()` | Negate the current filter |

🔝 [back to top](#api-reference)

&nbsp;

#### Array Operations

| Function | Description |
|----------|-------------|
| `filter.ElemMatch(field, filter)` | Create an elemMatch filter |
| `filter.Size(field, size)` | Create a size filter |
| `filter.All(field, values...)` | Create an all filter |

🔝 [back to top](#api-reference)

&nbsp;

#### String Operations

| Function | Description |
|----------|-------------|
| `filter.Regex(field, pattern, options...)` | Create a regex filter |
| `filter.Text(query)` | Create a text search filter |

🔝 [back to top](#api-reference)

&nbsp;

#### Existence Operations

| Function | Description |
|----------|-------------|
| `filter.Exists(field, exists)` | Create an exists filter |
| `filter.Type(field, bsonType)` | Create a type filter |

🔝 [back to top](#api-reference)

&nbsp;

### Update Builder (package `update`)

| Function | Description |
|----------|-------------|
| `update.New()` | Create a new update builder |
| `update.Set(field, value)` | Create a set operation |
| `update.Unset(fields...)` | Create an unset operation |
| `update.Inc(field, value)` | Create an increment operation |
| `update.Mul(field, value)` | Create a multiply operation |
| `update.Rename(from, to)` | Create a rename operation |
| `update.SetOnInsert(field, value)` | Create a setOnInsert operation |

🔝 [back to top](#api-reference)

&nbsp;

#### Array Update Operations

| Function | Description |
|----------|-------------|
| `update.Push(field, value)` | Create a push operation |
| `update.PushEach(field, values...)` | Create a push operation with multiple values |
| `update.Pull(field, filter)` | Create a pull operation |
| `update.AddToSet(field, value)` | Create an addToSet operation |
| `update.PopFirst(field)` | Create a pop first operation |
| `update.PopLast(field)` | Create a pop last operation |

🔝 [back to top](#api-reference)

&nbsp;

### Pipeline Builder (package `pipeline`)

| Function | Description |
|----------|-------------|
| `pipeline.New()` | Create a new pipeline builder |
| `builder.Match(filter)` | Add a match stage |
| `builder.Project(fields)` | Add a project stage |
| `builder.Sort(sorts)` | Add a sort stage |
| `builder.Limit(limit)` | Add a limit stage |
| `builder.Skip(skip)` | Add a skip stage |
| `builder.Group(id, fields)` | Add a group stage |
| `builder.Lookup(from, localField, foreignField, as)` | Add a lookup stage |
| `builder.Unwind(path)` | Add an unwind stage |
| `builder.AddFields(fields)` | Add an addFields stage |

🔝 [back to top](#api-reference)

&nbsp;

## Transaction Operations

| Function | Description |
|----------|-------------|
| `client.WithTransaction(ctx, fn)` | Execute a function within a transaction |

🔝 [back to top](#api-reference)

&nbsp;

## ID Generation (package `mongoid`)

&nbsp;

### ULID Operations

| Function | Description |
|----------|-------------|
| `mongoid.NewULID()` | Generate a new ULID string |
| `mongoid.ParseULID(str)` | Parse a ULID string |
| `mongoid.FindByULID(ctx, coll, id)` | Find document by ULID |
| `mongoid.UpdateByULID(ctx, coll, id, update)` | Update document by ULID |
| `mongoid.DeleteByULID(ctx, coll, id)` | Delete document by ULID |

🔝 [back to top](#api-reference)

&nbsp;

### ObjectID Operations

| Function | Description |
|----------|-------------|
| `mongoid.NewObjectID()` | Generate a new ObjectID |
| `mongoid.FindByObjectID(ctx, coll, id)` | Find document by ObjectID |
| `mongoid.UpdateByObjectID(ctx, coll, id, update)` | Update document by ObjectID |
| `mongoid.DeleteByObjectID(ctx, coll, id)` | Delete document by ObjectID |

🔝 [back to top](#api-reference)

&nbsp;

## Error Handling

&nbsp;

### Error Types

| Type | Description |
|------|-------------|
| `DuplicateKeyError` | Duplicate key violation |
| `ValidationError` | Document validation error |
| `ConnectionError` | Connection-related error |
| `WriteError` | Write operation error |

🔝 [back to top](#api-reference)

&nbsp;

### Error Checking Functions

| Function | Description |
|----------|-------------|
| `mongodb.IsDuplicateKeyError(err)` | Check if error is a duplicate key error |
| `mongodb.IsValidationError(err)` | Check if error is a validation error |
| `mongodb.IsConnectionError(err)` | Check if error is a connection error |
| `mongodb.IsNotFoundError(err)` | Check if error is a not found error |

🔝 [back to top](#api-reference)

&nbsp;

## Result Types

&nbsp;

### Insert Results

| Type | Description |
|------|-------------|
| `InsertOneResult` | Result of single document insert |
| `InsertManyResult` | Result of multiple document insert |

🔝 [back to top](#api-reference)

&nbsp;

### Update Results

| Type | Description |
|------|-------------|
| `UpdateResult` | Result of update operations |
| `ReplaceOneResult` | Result of replace operations |

🔝 [back to top](#api-reference)

&nbsp;

### Delete Results

| Type | Description |
|------|-------------|
| `DeleteResult` | Result of delete operations |

🔝 [back to top](#api-reference)

&nbsp;

### Find Results

| Type | Description |
|------|-------------|
| `FindOneResult` | Result of single document find |
| `Cursor` | Cursor for iterating over multiple documents |
| `ChangeStream` | Stream for watching collection changes |

🔝 [back to top](#api-reference)

&nbsp;

---

&nbsp;

An open source project brought to you by the [Cloudresty](https://cloudresty.com) team.

[Website](https://cloudresty.com) &nbsp;|&nbsp; [LinkedIn](https://www.linkedin.com/company/cloudresty) &nbsp;|&nbsp; [BlueSky](https://bsky.app/profile/cloudresty.com) &nbsp;|&nbsp; [GitHub](https://github.com/cloudresty) &nbsp;|&nbsp; [Docker Hub](https://hub.docker.com/u/cloudresty)

&nbsp;
