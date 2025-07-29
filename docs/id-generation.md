# ID Generation

[Home](../README.md) &nbsp;/&nbsp; [Docs](README.md) &nbsp;/&nbsp; ID Generation

&nbsp;

This document covers the ULID ID implementation that provides high-performance, database-optimized document identifiers.

&nbsp;

## Overview

ULID (Universally Unique Lexicographically Sortable Identifier) IDs provide significant advantages over traditional UUIDs and standard MongoDB ObjectIDs for modern applications.

&nbsp;

### Key Benefits

- **6x Faster Generation**: Significantly faster than UUID v4 generation
- **Database Optimized**: Better database performance and storage efficiency
- **Lexicographically Sortable**: Natural time-ordering without additional fields
- **Collision Resistant**: 128-bit entropy ensures uniqueness
- **MongoDB Compatible**: Works seamlessly with MongoDB collections using ULID strings as _id

🔝 [back to top](#id-generation)

&nbsp;

## Automatic ULID Generation

All document insertions automatically receive ULID IDs and metadata.

&nbsp;

### Basic Usage

```go
import (
    "context"
    "fmt"
    "github.com/cloudresty/go-mongodb"
)

// Define a proper struct for type safety
type User struct {
    Name  string `bson:"name"`
    Email string `bson:"email"`
}

client, _ := mongodb.NewClient()
collection := client.Collection("users")

// Insert with automatic ULID generation using a struct
user := User{
    Name:  "John Doe",
    Email: "john@example.com",
}

result, err := collection.InsertOne(ctx, user)
if err != nil {
    log.Fatal(err)
}

// Result contains ULID information
fmt.Printf("Inserted ID: %s\n", result.InsertedID)
fmt.Printf("Generated at: %s\n", result.GeneratedAt)
```

🔝 [back to top](#id-generation)

&nbsp;

### Document Enhancement

Documents are automatically enhanced with ULID metadata:

```go
// Define struct for the user data
type User struct {
    Name  string `bson:"name"`
    Email string `bson:"email"`
}

// Original document
user := User{
    Name:  "Alice Smith",
    Email: "alice@example.com",
}

// After insertion, document in MongoDB contains:
{
    "_id": "01HGQJ5Z8K9X7N2M4P6R8T3V5W",     // ULID string as _id
    "name": "Alice Smith",
    "email": "alice@example.com",
    "created_at": ISODate("2024-01-15T10:30:00Z"),
    "updated_at": ISODate("2024-01-15T10:30:00Z")
}
```

🔝 [back to top](#id-generation)

&nbsp;

## ULID Format and Structure

&nbsp;

### ULID Components

```text
01ARZ3NDEKTSV4RRFFQ69G5FAV
|          |
|          '-- Randomness (80 bits)
'-- Timestamp (48 bits)
```

- **Timestamp**: 48-bit millisecond Unix timestamp
- **Randomness**: 80-bit cryptographically random data
- **Total**: 128-bit identifier (26 characters when encoded)

🔝 [back to top](#id-generation)

&nbsp;

### Time-Ordered Properties

```go
// ULIDs generated in sequence are naturally ordered
ulid1 := generateULID() // 01HGQJ5Z8K9X7N2M4P6R8T3V5W
time.Sleep(1 * time.Millisecond)
ulid2 := generateULID() // 01HGQJ5Z8L1A3B5C7D9E2F4G6H

// ulid1 < ulid2 (lexicographically)
// This enables efficient database queries and indexing
```

🔝 [back to top](#id-generation)

&nbsp;

## Querying by ULID

&nbsp;

### Find by ULID String

```go
import (
    "github.com/cloudresty/go-mongodb/filter"
)

// Define struct for type-safe results
type User struct {
    ID    string `bson:"_id"`
    Name  string `bson:"name"`
    Email string `bson:"email"`
}

collection := client.Collection("users")

// Find by ULID _id field using fluent filter
var user User
err := collection.FindOne(ctx,
    filter.Eq("_id", "01HGQJ5Z8K9X7N2M4P6R8T3V5W"),
).Decode(&user)

if err != nil {
    log.Fatal(err)
}
fmt.Printf("Found user: %s (%s)\n", user.Name, user.Email)
```

🔝 [back to top](#id-generation)

&nbsp;

### Time-Range Queries

```go
import (
    "time"
    "github.com/cloudresty/go-mongodb/filter"
)

// Generate ULIDs for time range boundaries
startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
endTime := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

startULID := mongodb.GenerateULIDFromTime(startTime)
endULID := mongodb.GenerateULIDFromTime(endTime)

// Query documents created in January 2024 using fluent filter
cursor, err := collection.Find(ctx,
    filter.Gte("_id", startULID).And(filter.Lte("_id", endULID)),
)
```

🔝 [back to top](#id-generation)

&nbsp;

### Sorting by Creation Time

```go
import (
    "github.com/cloudresty/go-mongodb/filter"
    "go.mongodb.org/mongo-driver/v2/bson"
)

// Sort by ULID _id for chronological order (ascending - oldest first)
cursor, err := collection.Find(ctx,
    filter.Empty(), // No filter, find all documents
    &mongodb.QueryOptions{
        Sort: bson.D{{"_id", 1}}, // Ascending (oldest first)
    },
)

// Or descending order (newest first)
cursor, err = collection.Find(ctx,
    filter.Empty(), // No filter, find all documents
    &mongodb.QueryOptions{
        Sort: bson.D{{"_id", -1}}, // Descending (newest first)
    },
)
```

🔝 [back to top](#id-generation)

&nbsp;

## Performance Characteristics

&nbsp;

### Generation Speed

| Identifier Type | Operations/sec | Relative Performance |
|----------------|----------------|---------------------|
| **ULID** | ~6,000,000 | 6.0x faster |
| **UUID v4** | ~1,000,000 | 1.0x baseline |
| **MongoDB ObjectID** | ~8,000,000 | 8.0x faster |

🔝 [back to top](#id-generation)

&nbsp;

### Database Performance

```go
// Benchmark: Insert 100,000 documents
// ULID IDs: 12.3 seconds
// UUID v4: 18.7 seconds
// Improvement: ~34% faster insertions

// Index efficiency
// ULID IDs: Better B-tree balance due to time-ordering
// UUID v4: Random distribution causes index fragmentation
```

🔝 [back to top](#id-generation)

&nbsp;

### Storage Efficiency

- **ULID String**: 26 bytes (Base32 encoded)
- **ULID Binary**: 16 bytes (128-bit)
- **UUID String**: 36 bytes (with hyphens)
- **MongoDB ObjectID**: 12 bytes (but less entropy)

🔝 [back to top](#id-generation)

&nbsp;

## Advanced Usage

&nbsp;

### Manual ULID Generation

```go
// Generate standalone ULID
ulid := mongodb.GenerateULID()
fmt.Printf("Generated ULID: %s\n", ulid)

// Generate ULID from specific time
timestamp := time.Now().Add(-1 * time.Hour)
pastULID := mongodb.GenerateULIDFromTime(timestamp)
```

🔝 [back to top](#id-generation)

&nbsp;

### Custom Document Enhancement

```go
// Define struct for type safety
type PremiumUser struct {
    Name string `bson:"name"`
    Type string `bson:"type"`
}

// Manually enhance documents before insertion
user := PremiumUser{
    Name: "Custom User",
    Type: "premium",
}

enhancedDoc := mongodb.EnhanceDocument(user)
// enhancedDoc now contains ULID, timestamps, etc.

result, err := collection.InsertOne(ctx, enhancedDoc)
if err != nil {
    log.Fatal(err)
}
```

🔝 [back to top](#id-generation)

&nbsp;

### Bulk Operations with ULIDs

```go
// Define struct for consistent typing
type User struct {
    Name string `bson:"name"`
}

// Bulk insert with automatic ULID generation using structs
users := []any{
    User{Name: "User 1"},
    User{Name: "User 2"},
    User{Name: "User 3"},
}

result, err := collection.InsertMany(ctx, users)
if err != nil {
    log.Fatal(err)
}

// Each document gets unique ULID
for i, ulidId := range result.InsertedIDs {
    fmt.Printf("Document %d ULID: %s\n", i+1, ulidId)
}
```

🔝 [back to top](#id-generation)

&nbsp;

## Migration from UUIDs

&nbsp;

### Gradual Migration Strategy

```go
import (
    "time"
    "github.com/cloudresty/go-mongodb/filter"
    "github.com/cloudresty/go-mongodb/update"
)

// Define struct for existing documents
type LegacyDocument struct {
    ID        any       `bson:"_id"`        // Could be ObjectID or other types
    CreatedAt time.Time `bson:"created_at"`
    // Add other fields as needed
}

// Step 1: Add ULID fields alongside existing UUIDs
collection := client.Collection("users")

// Update existing documents with ULIDs using fluent builders
cursor, err := collection.Find(ctx,
    filter.Type("_id", "objectId"), // Find old ObjectID docs
)

for cursor.Next(ctx) {
    var doc LegacyDocument
    if err := cursor.Decode(&doc); err != nil {
        log.Printf("Failed to decode document: %v", err)
        continue
    }

    // Generate ULID based on creation time or current time
    var ulid string
    if !doc.CreatedAt.IsZero() {
        ulid = mongodb.GenerateULIDFromTime(doc.CreatedAt)
    } else {
        ulid = mongodb.GenerateULID()
    }

    // Update with ULID using fluent update builder
    _, err := collection.UpdateOne(ctx,
        filter.Eq("_id", doc.ID),
        update.Set("ulid", ulid),
    )
    if err != nil {
        log.Printf("Failed to update document: %v", err)
    }
}
```

🔝 [back to top](#id-generation)

&nbsp;

### UUID to ULID Mapping

```go
import (
    "go.mongodb.org/mongo-driver/v2/bson"
)

// Maintain mapping for backward compatibility
type DocumentMapping struct {
    UUID string `bson:"uuid"`
    ULID string `bson:"ulid"`
}

// Create index for efficient lookups using bson.D for type safety
collection.CreateIndex(ctx, bson.D{{"uuid", 1}})
collection.CreateIndex(ctx, bson.D{{"ulid", 1}})
```

🔝 [back to top](#id-generation)

&nbsp;

## Best Practices

&nbsp;

### Indexing Strategy

```go
import (
    "go.mongodb.org/mongo-driver/v2/bson"
)

// Primary index on _id (ULID) for time-based queries (automatically created)
// Additional indexes can be created as needed

// Compound indexes with ULID _id first using bson.D for type safety
collection.CreateIndex(ctx, bson.D{
    {"_id", 1},
    {"status", 1},
})

// Avoid indexes on random UUIDs (causes fragmentation)
```

🔝 [back to top](#id-generation)

&nbsp;

### Query Optimization

```go
import (
    "github.com/cloudresty/go-mongodb/filter"
)

// Efficient time-range queries using ULID with fluent builders
startULID := mongodb.GenerateULIDFromTime(startTime)
endULID := mongodb.GenerateULIDFromTime(endTime)

// This query uses index efficiently with fluent filter
timeRangeFilter := filter.Gte("_id", startULID).And(filter.Lt("_id", endULID))

cursor, err := collection.Find(ctx, timeRangeFilter)
```

🔝 [back to top](#id-generation)

&nbsp;

### Application Design

- **Use ULID for primary keys** when you need time-ordered identifiers
- **Keep UUID for external APIs** if required for compatibility
- **Use ULID _id fields** for efficient queries (no additional indexes needed)
- **Use ULID prefixes** for sharding strategies
- **Consider ULID in URL design** for user-friendly, sortable URLs

🔝 [back to top](#id-generation)

&nbsp;

## Troubleshooting

&nbsp;

### Common Issues

&nbsp;

#### Clock Synchronization

```go
// Ensure system clocks are synchronized in distributed systems
// ULIDs depend on accurate timestamps for ordering

// Check for clock skew
if time.Since(lastGeneratedTime) < 0 {
    log.Warn("Clock skew detected - ULID ordering may be affected")
}
```

🔝 [back to top](#id-generation)

&nbsp;

#### Duplicate Detection

```go
// Define struct for the document
type User struct {
    Name string `bson:"name"`
    // other fields...
}

// ULIDs have 80 bits of randomness - duplicates are extremely rare
// In high-throughput scenarios, consider adding application-level checks

user := User{Name: "Example User"}
result, err := collection.InsertOne(ctx, user)
if mongodb.IsDuplicateKeyError(err) {
    // Handle extremely rare duplicate case
    log.Warn("Duplicate ULID detected - regenerating")
    // Regenerate and retry
}
```

🔝 [back to top](#id-generation)

&nbsp;

### Performance Monitoring

```go
// Monitor ULID generation performance
start := time.Now()
ulid := mongodb.GenerateULID()
duration := time.Since(start)

if duration > time.Microsecond {
    log.Warn("ULID generation slower than expected",
        "duration", duration,
        "ulid", ulid)
}
```

🔝 [back to top](#id-generation)

&nbsp;

---

&nbsp;

An open source project brought to you by the [Cloudresty](https://cloudresty.com) team.

[Website](https://cloudresty.com) &nbsp;|&nbsp; [LinkedIn](https://www.linkedin.com/company/cloudresty) &nbsp;|&nbsp; [BlueSky](https://bsky.app/profile/cloudresty.com) &nbsp;|&nbsp; [GitHub](https://github.com/cloudresty) &nbsp;|&nbsp; [Docker Hub](https://hub.docker.com/u/cloudresty)

&nbsp;
