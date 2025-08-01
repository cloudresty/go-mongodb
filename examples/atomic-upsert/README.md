# Atomic Upsert Example

This example demonstrates the atomic upsert functionality using MongoDB's `$setOnInsert` operator to prevent race conditions and ensure data integrity.

## What it demonstrates

- **Atomic upsert operations** that only insert when documents don't exist
- **Multiple approaches** for the same atomic upsert behavior
- **Race condition prevention** - documents are never modified once created
- **Convenience methods** for cleaner, more readable code

## Features showcased

### 1. Individual Field SetOnInsert (Original)

```go
updateBuilder := update.New().
    SetOnInsert("_id", event.ID).
    SetOnInsert("media_id", event.MediaID).
    SetOnInsert("title", event.Title)
```

### 2. Bulk SetOnInsert with Map (NEW)

```go
fieldMap := map[string]any{
    "_id": event.ID,
    "media_id": event.MediaID,
    "title": event.Title,
}
updateBuilder := update.New().SetOnInsertMap(fieldMap)
```

### 3. Bulk SetOnInsert with Struct (NEW)

```go
updateBuilder := update.New().SetOnInsertStruct(event)
```

### 4. Convenience UpsertByField (NEW)

```go
result, err := collection.UpsertByField(ctx, "url", event.URL, event)
```

### 5. Convenience UpsertByFieldMap (NEW)

```go
result, err := collection.UpsertByFieldMap(ctx, "url", event.URL, fieldMap)
```

## Key Benefits

- **Prevents duplicate entries** based on unique field values
- **Atomic operations** - no race conditions
- **Existing documents are never modified** with `$setOnInsert`
- **Multiple convenient APIs** for different use cases

## Running the example

```bash
# Set up your MongoDB connection
export MONGODB_HOSTS=localhost:27017
export MONGODB_DATABASE=test

# Run the example
go run examples/atomic-upsert/main.go
```

## Expected output

The example will show each method creating a document, then demonstrate that subsequent attempts with the same unique field value will not modify the existing document, proving the atomic behavior.

## Use cases

- **Event deduplication** - ensure events are only recorded once
- **User registration** - prevent duplicate accounts
- **Cache-like operations** - set values only if they don't exist
- **Idempotent operations** - safe to retry without side effects
