# Migration Guide: v1 to v2

[Home](../README.md) &nbsp;/&nbsp; [Docs](README.md) &nbsp;/&nbsp; Migration Guide v2

&nbsp;

This guide covers the breaking changes and migration steps from go-mongodb v1 to v2.

&nbsp;

## Why v2?

Version 2.0.0 represents a major architectural overhaul focused on:

- **Safety**: Strict type validation prevents data corruption
- **Performance**: Zero-allocation ID injection for optimal speed
- **API Design**: Idiomatic Go patterns with methods on Collection
- **Reliability**: Removed manual reconnection logic in favor of driver's SDAM

&nbsp;

## Breaking Changes

### 1. Module Path Change

**v1:**
```go
import "github.com/cloudresty/go-mongodb"
import "github.com/cloudresty/go-mongodb/filter"
```

**v2:**
```go
import "github.com/cloudresty/go-mongodb/v2"
import "github.com/cloudresty/go-mongodb/v2/filter"
```

### 2. NewULID() Now Panics on Entropy Failure

**v1:** Returns empty string on failure (silent data corruption risk)
```go
id := mongoid.NewULID() // Could return "" on entropy failure
```

**v2:** Panics on failure (fail-fast, prevents duplicate key errors)
```go
id := mongoid.NewULID() // Panics if entropy source fails

// For explicit error handling:
id, err := mongoid.NewULIDWithError()
```

### 3. Strict ID Type Validation in ULID Mode

**v1:** Allowed any ID field type (could cause decode failures)
```go
type User struct {
    ID int64 `bson:"_id"` // Would insert string ULID, crash on read
}
```

**v2:** Rejects incompatible ID types with clear error
```go
type User struct {
    ID int64 `bson:"_id"` // Returns ErrULIDIncompatibleType
}

// Use string or interface{} for ULID mode:
type User struct {
    ID string `bson:"_id"` // ✓ Compatible with ULID
}
```

### 4. ID-Based Operations Moved to Collection Methods

**v1:** Separate mongoid package with broken interface
```go
// This didn't work due to interface mismatch
result := mongoid.FindByULID(ctx, collection, id)
```

**v2:** Methods directly on Collection (idiomatic Go)
```go
// Clean, type-safe API
result := collection.FindByID(ctx, id)
updateResult, err := collection.UpdateByID(ctx, id, updateBuilder)
deleteResult, err := collection.DeleteByID(ctx, id)
```

### 5. Removed Reconnection Configuration

**v1:** Manual reconnection settings (caused "zombie collection" bugs)
```go
config := &mongodb.Config{
    ReconnectEnabled:     true,
    ReconnectDelay:       5 * time.Second,
    MaxReconnectAttempts: 10,
}
```

**v2:** Driver handles reconnection automatically via SDAM
```go
config := &mongodb.Config{
    // Reconnection settings removed - driver handles this
}
```

### 6. Removed Deprecated Methods

**v1:**
```go
client.GetReconnectCount()    // Always returned 0
client.GetLastReconnectTime() // Always returned zero time
```

**v2:** These methods are removed entirely.

&nbsp;

## Migration Steps

### Step 1: Update go.mod

```bash
# Remove v1
go mod edit -droprequire github.com/cloudresty/go-mongodb

# Add v2
go get github.com/cloudresty/go-mongodb/v2
```

### Step 2: Update Imports

Find and replace all imports:

```bash
# Using sed (macOS/Linux)
find . -name "*.go" -type f | xargs sed -i '' 's|github.com/cloudresty/go-mongodb|github.com/cloudresty/go-mongodb/v2|g'
```

### Step 3: Update ID Field Types

Ensure all structs used with ULID mode have string ID fields:

```go
// Before (may cause issues)
type User struct {
    ID primitive.ObjectID `bson:"_id"`
}

// After (compatible with ULID mode)
type User struct {
    ID string `bson:"_id"`
}

// Or use ObjectID mode if you need ObjectID:
client, _ := mongodb.NewClient(mongodb.WithIDMode(mongodb.IDModeObjectID))
```

### Step 4: Update ID-Based Operations

Replace mongoid helper functions with Collection methods:

```go
// Before
result := mongoid.FindByULID(ctx, col, id)

// After
result := col.FindByID(ctx, id)
```

### Step 5: Remove Reconnection Configuration

Remove any reconnection-related configuration:

```go
// Before
config := &mongodb.Config{
    ReconnectEnabled: true,
    // ... other reconnection settings
}

// After
config := &mongodb.Config{
    // Reconnection is automatic via driver's SDAM
}
```

### Step 6: Test Thoroughly

```bash
go build ./...
go test -race ./...
```

&nbsp;

## New Features in v2

### FindByID, UpdateByID, DeleteByID

```go
// Find by any ID type
result := col.FindByID(ctx, "01ARZ3NDEKTSV4RRFFQ69G5FAV")

// Update by ID
updateResult, err := col.UpdateByID(ctx, id, update.Set("name", "New Name"))

// Delete by ID
deleteResult, err := col.DeleteByID(ctx, id)
```

### Zero-Allocation ID Injection

For struct pointers with string ID fields, v2 sets the ULID directly on the struct without marshal/unmarshal overhead.

### Command Monitoring

```go
client, _ := mongodb.NewClient(
    mongodb.WithMonitor(&event.CommandMonitor{
        Started: func(ctx context.Context, e *event.CommandStartedEvent) {
            log.Printf("Command: %s", e.CommandName)
        },
    }),
)
```

&nbsp;

---

### Cloudresty

[Website](https://cloudresty.com) &nbsp;|&nbsp; [GitHub](https://github.com/cloudresty)

<sub>&copy; Cloudresty</sub>

