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
| :--- | :--- |
| `NewClient(options...)` | Creates a client with functional options (use `FromEnv()` to load from environment variables) |

&nbsp;

🔝 [back to top](#api-reference)

&nbsp;

#### Client Options

| Option | Description |
| :--- | :--- |
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

&nbsp;

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

&nbsp;

**Default behavior**: If no logger is provided via `WithLogger()`, the client uses `NopLogger` (silent - no output).

&nbsp;

**Usage patterns**:

- Fields are provided as alternating key-value pairs: `logger.Info("message", "key1", value1, "key2", value2)`
- Supported field types: `string`, `int`, `int64`, `time.Duration`, `bool`, `error`
- See [Custom Logging Example](../examples/custom-logger-emit/) for `emit` library integration

&nbsp;

🔝 [back to top](#api-reference)

&nbsp;

### Connection Operations

| Function | Description |
| :--- | :--- |
| `client.Ping(ctx context.Context) error` | Test connection and update internal state |
| `client.Stats() *ClientStats` | Get connection statistics (reconnect count, operations, etc.) |
| `client.Name() string` | Get the connection name for this client instance |
| `client.Close() error` | Close the client and stop background routines |

&nbsp;

🔝 [back to top](#api-reference)

&nbsp;

## Environment Configuration

| Function | Description |
| :--- | :--- |
| `FromEnv() Option` | Load configuration from `MONGODB_*` environment variables (functional option) |
| `FromEnvWithPrefix(prefix string) Option` | Load configuration with custom prefix (e.g., `MYAPP_MONGODB_*`) (functional option) |

&nbsp;

🔝 [back to top](#api-reference)

&nbsp;

## Resource-Oriented API

&nbsp;

### Database Operations

| Function | Description |
| :--- | :--- |
| `client.Database(name string) *Database` | Get a database handle for the specified name |
| `database.Name() string` | Get the database name |
| `database.Collection(name string) *Collection` | Get a collection handle for the specified name |
| `database.Drop(ctx context.Context) error` | Drop the database |
| `database.ListCollections(ctx context.Context) ([]string, error)` | List all collections in the database |

&nbsp;

🔝 [back to top](#api-reference)

&nbsp;

### Collection Operations

| Function | Description |
| :--- | :--- |
| `collection.Name() string` | Get the collection name |
| `collection.Database() *Database` | Get the parent database |

&nbsp;

🔝 [back to top](#api-reference)

&nbsp;

## Document Operations

&nbsp;

### Basic CRUD Operations

| Function | Description |
| :--- | :--- |
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

&nbsp;

🔝 [back to top](#api-reference)

&nbsp;

### Enhanced Find Operations

| Function | Description |
| :--- | :--- |
| `collection.FindWithOptions(ctx, filter, queryOpts) (*FindResult, error)` | Find documents with QueryOptions (sort, limit, skip, projection) |
| `collection.FindOneWithOptions(ctx, filter, queryOpts) *FindOneResult` | Find single document with QueryOptions |
| `collection.FindSorted(ctx, filter, sort, opts...) (*FindResult, error)` | Find documents with sort order |
| `collection.FindOneSorted(ctx, filter, sort) *FindOneResult` | Find single document with sort order |
| `collection.FindWithLimit(ctx, filter, limit) (*FindResult, error)` | Find documents with limit |
| `collection.FindWithSkip(ctx, filter, skip) (*FindResult, error)` | Find documents with skip offset |
| `collection.FindWithProjection(ctx, filter, projection) (*FindResult, error)` | Find documents with field projection |

&nbsp;

🔝 [back to top](#api-reference)

&nbsp;

### Advanced Operations

| Function | Description |
| :--- | :--- |
| `collection.Aggregate(ctx, pipeline) (*Cursor, error)` | Run aggregation pipeline |
| `collection.AggregateWithPipeline(ctx, pipelineBuilder, opts...) (*AggregateResult, error)` | Run aggregation using pipeline builder |
| `collection.Distinct(ctx, field, filter) ([]any, error)` | Get distinct values for a field |
| `collection.Watch(ctx, pipeline, opts...) (*ChangeStream, error)` | Watch for changes |

&nbsp;

🔝 [back to top](#api-reference)

&nbsp;

### Atomic Upsert Operations

| Function | Description |
| :--- | :--- |
| `collection.UpsertByField(ctx, field, value, document) (*UpdateResult, error)` | Atomic upsert using $setOnInsert for struct |
| `collection.UpsertByFieldMap(ctx, field, value, fields) (*UpdateResult, error)` | Atomic upsert using $setOnInsert for map |
| `collection.UpsertByFieldWithOptions(ctx, field, value, document, opts) (*UpdateResult, error)` | Atomic upsert with configuration options |

**Note**: All upsert methods use `$setOnInsert` by default, ensuring existing documents are never modified and preventing race conditions.

&nbsp;

🔝 [back to top](#api-reference)

&nbsp;

### Atomic Find-And-Modify Operations

| Function | Description |
| :--- | :--- |
| `collection.FindOneAndUpdate(ctx, filter, update, opts...) *FindOneResult` | Atomically find, update, and return a document |
| `collection.FindOneAndReplace(ctx, filter, replacement, opts...) *FindOneResult` | Atomically find, replace, and return a document |
| `collection.FindOneAndDelete(ctx, filter, opts...) *FindOneResult` | Atomically find and delete a document, returning the deleted document |

&nbsp;

### FindOneAndUpdate Options

| Type / Function | Description |
| :--- | :--- |
| `ReturnDocument` | Enum type: `ReturnBefore` (default) or `ReturnAfter` |
| `FindOneAndUpdateOpts()` | Create new options with defaults |
| `SetReturnDocument(rd ReturnDocument)` | Set whether to return document before or after update |
| `SetUpsert(upsert bool)` | Set whether to insert if no document matches |
| `SetSort(sort bson.D)` | Set sort order to determine which document to update |
| `SetProjection(projection bson.D)` | Set which fields to return |

&nbsp;

### FindOneAndReplace Options

| Type / Function | Description |
| :--- | :--- |
| `FindOneAndReplaceOpts()` | Create new options with defaults |
| `SetReturnDocument(rd ReturnDocument)` | Set whether to return document before or after replacement |
| `SetUpsert(upsert bool)` | Set whether to insert if no document matches |
| `SetSort(sort bson.D)` | Set sort order to determine which document to replace |
| `SetProjection(projection bson.D)` | Set which fields to return |

&nbsp;

### FindOneAndDelete Options

| Type / Function | Description |
| :--- | :--- |
| `FindOneAndDeleteOpts()` | Create new options with defaults |
| `SetSort(sort bson.D)` | Set sort order to determine which document to delete |
| `SetProjection(projection bson.D)` | Set which fields to return |

&nbsp;

🔝 [back to top](#api-reference)

&nbsp;

## Fluent Query Builders

&nbsp;

### Filter Builder (package `filter`)

| Function | Description |
| :--- | :--- |
| `filter.New()` | Create a new filter builder |
| `filter.Eq(field, value)` | Create an equality filter |
| `filter.Ne(field, value)` | Create a not-equal filter |
| `filter.Gt(field, value)` | Create a greater-than filter |
| `filter.Gte(field, value)` | Create a greater-than-or-equal filter |
| `filter.Lt(field, value)` | Create a less-than filter |
| `filter.Lte(field, value)` | Create a less-than-or-equal filter |
| `filter.In(field, values...)` | Create an in filter |
| `filter.Nin(field, values...)` | Create a not-in filter |

&nbsp;

🔝 [back to top](#api-reference)

&nbsp;

#### Logical Operations (Fluent Methods)

| Method | Description |
| :--- | :--- |
| `builder.And(filters...)` | Combine filters with logical AND (fluent method) |
| `builder.Or(filters...)` | Combine filters with logical OR (fluent method) |
| `builder.Not()` | Negate the current filter |

&nbsp;

🔝 [back to top](#api-reference)

&nbsp;

#### Array Operations

| Function | Description |
| :--- | :--- |
| `filter.ElemMatch(field, filter)` | Create an elemMatch filter |
| `filter.Size(field, size)` | Create a size filter |
| `filter.All(field, values...)` | Create an all filter |

&nbsp;

🔝 [back to top](#api-reference)

&nbsp;

#### String Operations

| Function | Description |
| :--- | :--- |
| `filter.Regex(field, pattern, options...)` | Create a regex filter |
| `filter.Text(query)` | Create a text search filter |

&nbsp;

🔝 [back to top](#api-reference)

&nbsp;

#### Existence Operations

| Function | Description |
| :--- | :--- |
| `filter.Exists(field, exists)` | Create an exists filter |
| `filter.Type(field, bsonType)` | Create a type filter |

&nbsp;

🔝 [back to top](#api-reference)

&nbsp;

### Update Builder (package `update`)

| Function | Description |
| :--- | :--- |
| `update.New()` | Create a new update builder |
| `update.Set(field, value)` | Create a set operation |
| `update.SetMap(fields)` | Create a set operation for multiple fields from map |
| `update.SetStruct(document)` | Create a set operation for all fields from struct |
| `update.Unset(fields...)` | Create an unset operation |
| `update.Inc(field, value)` | Create an increment operation |
| `update.Mul(field, value)` | Create a multiply operation |
| `update.Rename(from, to)` | Create a rename operation |
| `update.SetOnInsert(field, value)` | Create a setOnInsert operation for single field |
| `update.SetOnInsertMap(fields)` | Create a setOnInsert operation for multiple fields from map |
| `update.SetOnInsertStruct(document)` | Create a setOnInsert operation for all fields from struct |

&nbsp;

🔝 [back to top](#api-reference)

&nbsp;

#### Array Update Operations

| Function | Description |
| :--- | :--- |
| `update.Push(field, value)` | Create a push operation |
| `update.PushEach(field, values...)` | Create a push operation with multiple values |
| `update.Pull(field, filter)` | Create a pull operation |
| `update.AddToSet(field, value)` | Create an addToSet operation |
| `update.PopFirst(field)` | Create a pop first operation |
| `update.PopLast(field)` | Create a pop last operation |

&nbsp;

🔝 [back to top](#api-reference)

&nbsp;

### Pipeline Builder (package `pipeline`)

| Function | Description |
| :--- | :--- |
| `pipeline.New()` | Create a new pipeline builder |
| `builder.Match(filter)` | Add a $match stage with filter builder |
| `builder.MatchRaw(filter)` | Add a $match stage with raw bson.M |
| `builder.Project(fields)` | Add a $project stage |
| `builder.Sort(sorts)` | Add a $sort stage with bson.D |
| `builder.SortMap(sorts)` | Add a $sort stage with map |
| `builder.Limit(limit)` | Add a $limit stage |
| `builder.Skip(skip)` | Add a $skip stage |
| `builder.Group(id, fields)` | Add a $group stage |
| `builder.Lookup(from, localField, foreignField, as)` | Add a $lookup stage |
| `builder.Unwind(path)` | Add an $unwind stage |
| `builder.UnwindWithOptions(path, preserveNull, arrayIndex)` | Add $unwind with options |
| `builder.AddFields(fields)` | Add an $addFields stage |
| `builder.ReplaceRoot(newRoot)` | Add a $replaceRoot stage |
| `builder.Facet(facets)` | Add a $facet stage |
| `builder.Count(field)` | Add a $count stage |
| `builder.Sample(size)` | Add a $sample stage |
| `builder.Raw(stage)` | Add a custom stage |
| `builder.Build()` | Build pipeline as []bson.M |
| `builder.ToBSONArray()` | Build pipeline as bson.A |

&nbsp;

🔝 [back to top](#api-reference)

&nbsp;

#### Standalone Pipeline Functions

| Function | Description |
| :--- | :--- |
| `pipeline.Match(filter)` | Create pipeline starting with $match |
| `pipeline.MatchRaw(filter)` | Create pipeline starting with raw $match |
| `pipeline.Project(fields)` | Create pipeline starting with $project |
| `pipeline.Sort(sorts)` | Create pipeline starting with $sort |
| `pipeline.SortMap(sorts)` | Create pipeline starting with $sort from map |
| `pipeline.Limit(limit)` | Create pipeline starting with $limit |
| `pipeline.Skip(skip)` | Create pipeline starting with $skip |
| `pipeline.Group(id, fields)` | Create pipeline starting with $group |

&nbsp;

🔝 [back to top](#api-reference)

&nbsp;

#### Aggregation with Pipeline Builder

| Function | Description |
| :--- | :--- |
| `collection.AggregateWithPipeline(ctx, pipelineBuilder, opts...)` | Execute aggregation with pipeline builder |

&nbsp;

🔝 [back to top](#api-reference)

&nbsp;

## Transaction Operations

| Function | Description |
| :--- | :--- |
| `client.WithTransaction(ctx, fn)` | Execute a function within a transaction |

&nbsp;

🔝 [back to top](#api-reference)

&nbsp;

## ID Generation (package `mongoid`)

&nbsp;

### ULID Operations

| Function | Description |
| :--- | :--- |
| `mongoid.NewULID()` | Generate a new ULID string |
| `mongoid.ParseULID(str)` | Parse a ULID string |
| `mongoid.FindByULID(ctx, coll, id)` | Find document by ULID |
| `mongoid.UpdateByULID(ctx, coll, id, update)` | Update document by ULID |
| `mongoid.DeleteByULID(ctx, coll, id)` | Delete document by ULID |

&nbsp;

🔝 [back to top](#api-reference)

&nbsp;

### ObjectID Operations

| Function | Description |
| :--- | :--- |
| `mongoid.NewObjectID()` | Generate a new ObjectID |
| `mongoid.FindByObjectID(ctx, coll, id)` | Find document by ObjectID |
| `mongoid.UpdateByObjectID(ctx, coll, id, update)` | Update document by ObjectID |
| `mongoid.DeleteByObjectID(ctx, coll, id)` | Delete document by ObjectID |

&nbsp;

🔝 [back to top](#api-reference)

&nbsp;

## Index Management

The library provides a self-contained index management system with helper functions that eliminate the need to import mongo-driver directly for index operations.

&nbsp;

### Index Operations

| Function | Description |
| :--- | :--- |
| `collection.CreateIndex(ctx, model)` | Create a single index using IndexModel |
| `collection.CreateIndexes(ctx, models)` | Create multiple indexes using []IndexModel |
| `collection.DropIndex(ctx, name)` | Drop an index by name |
| `collection.ListIndexes(ctx)` | List all indexes in the collection |
| `collection.Indexes()` | Get the IndexView for advanced index operations |

&nbsp;

🔝 [back to top](#api-reference)

&nbsp;

### Index Model Type

```go
type IndexModel struct {
    Keys    bson.D                       // Index keys with sort direction
    Options *options.IndexOptionsBuilder // Optional index options
}
```

&nbsp;

🔝 [back to top](#api-reference)

&nbsp;

### Basic Index Helpers

| Function | Description |
| :--- | :--- |
| `IndexAsc(fields...)` | Create an ascending index on one or more fields |
| `IndexDesc(fields...)` | Create a descending index on one or more fields |
| `IndexUnique(fields...)` | Create a unique ascending index |
| `IndexText(fields...)` | Create a text search index |

&nbsp;

🔝 [back to top](#api-reference)

&nbsp;

### Compound Index Helpers

| Function | Description |
| :--- | :--- |
| `IndexCompound(field, dir, ...)` | Create a compound index with mixed directions (1=asc, -1=desc) |

&nbsp;

🔝 [back to top](#api-reference)

&nbsp;

### Special Index Types

| Function | Description |
| :--- | :--- |
| `IndexTTL(field, duration)` | Create a TTL index for automatic document expiration |
| `IndexSparse(fields...)` | Create a sparse index (only includes documents with the field) |
| `IndexHashed(field)` | Create a hashed index for sharding or hash-based lookups |
| `Index2DSphere(field)` | Create a geospatial 2dsphere index for GeoJSON data |
| `IndexPartial(filter, fields...)` | Create a partial index with a filter expression |

&nbsp;

🔝 [back to top](#api-reference)

&nbsp;

### Index Modifiers

| Function | Description |
| :--- | :--- |
| `IndexWithName(name, model)` | Add a custom name to any IndexModel |
| `IndexUniqueWithOptions(fields, sparse, name)` | Create a unique index with additional options |

&nbsp;

🔝 [back to top](#api-reference)

&nbsp;

## Error Handling

&nbsp;

### Error Types

| Type | Description |
| :--- | :--- |
| `DuplicateKeyError` | Duplicate key violation |
| `ValidationError` | Document validation error |
| `ConnectionError` | Connection-related error |
| `WriteError` | Write operation error |

&nbsp;

🔝 [back to top](#api-reference)

&nbsp;

### Error Checking Functions

| Function | Description |
| :--- | :--- |
| `mongodb.IsDuplicateKeyError(err)` | Check if error is a duplicate key error |
| `mongodb.IsValidationError(err)` | Check if error is a validation error |
| `mongodb.IsConnectionError(err)` | Check if error is a connection error |
| `mongodb.IsNotFoundError(err)` | Check if error is a not found error |

&nbsp;

🔝 [back to top](#api-reference)

&nbsp;

## Result Types

&nbsp;

### Insert Results

| Type | Description |
| :--- | :--- |
| `InsertOneResult` | Result of single document insert |
| `InsertManyResult` | Result of multiple document insert |

&nbsp;

🔝 [back to top](#api-reference)

&nbsp;

### Update Results

| Type | Description |
| :--- | :--- |
| `UpdateResult` | Result of update operations |
| `ReplaceOneResult` | Result of replace operations |

&nbsp;

🔝 [back to top](#api-reference)

&nbsp;

### Delete Results

| Type | Description |
| :--- | :--- |
| `DeleteResult` | Result of delete operations |

&nbsp;

🔝 [back to top](#api-reference)

&nbsp;

### Find Results

| Type | Description |
| :--- | :--- |
| `FindOneResult` | Result of single document find |
| `FindResult` | Result of enhanced find operations with convenient cursor methods |
| `AggregateResult` | Result of aggregation operations with cursor functionality |
| `Cursor` | Cursor for iterating over multiple documents |
| `ChangeStream` | Stream for watching collection changes |

&nbsp;

🔝 [back to top](#api-reference)

&nbsp;

&nbsp;

---

### Cloudresty

[Website](https://cloudresty.com) &nbsp;|&nbsp; [LinkedIn](https://www.linkedin.com/company/cloudresty) &nbsp;|&nbsp; [BlueSky](https://bsky.app/profile/cloudresty.com) &nbsp;|&nbsp; [GitHub](https://github.com/cloudresty) &nbsp;|&nbsp; [Docker Hub](https://hub.docker.com/u/cloudresty)

<sub>&copy; Cloudresty</sub>

&nbsp;
