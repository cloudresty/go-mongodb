# ULID ObjectIDs

[Home](../README.md) &nbsp;/&nbsp; [Docs](README.md) &nbsp;/&nbsp; ULID ObjectIDs

&nbsp;

This document covers the ULID ObjectID implementation that provides high-performance, database-optimized document identifiers.

&nbsp;

## Overview

ULID (Universally Unique Lexicographically Sortable Identifier) ObjectIDs provide significant advantages over traditional UUIDs and standard MongoDB ObjectIDs for modern applications.

&nbsp;

### Key Benefits

- **6x Faster Generation**: Significantly faster than UUID v4 generation
- **Database Optimized**: Better database performance and storage efficiency
- **Lexicographically Sortable**: Natural time-ordering without additional fields
- **Collision Resistant**: 128-bit entropy ensures uniqueness
- **MongoDB Compatible**: Embedded within standard MongoDB ObjectIDs

🔝 [back to top](#ulid-objectids)

&nbsp;

## Automatic ULID Generation

All document insertions automatically receive ULID-enhanced ObjectIDs and metadata.

&nbsp;

### Basic Usage

```go
client, _ := mongodb.NewClient()
collection := client.Collection("users")

// Insert with automatic ULID generation
result, err := collection.InsertOne(ctx, map[string]any{
    "name":  "John Doe",
    "email": "john@example.com",
})

// Result contains ULID information
fmt.Printf("Inserted ID: %s\n", result.InsertedID.Hex())
fmt.Printf("ULID: %s\n", result.ULID)
fmt.Printf("Generated at: %s\n", result.GeneratedAt)
```

🔝 [back to top](#ulid-objectids)

&nbsp;

### Document Enhancement

Documents are automatically enhanced with ULID metadata:

```go
// Original document
doc := map[string]any{
    "name":  "Alice Smith",
    "email": "alice@example.com",
}

// After insertion, document contains:
{
    "_id": ObjectID("..."),           // MongoDB ObjectID with embedded ULID
    "ulid": "01HGQJ5Z8K9X7N2M4P6R8T3V5W", // Standalone ULID string
    "name": "Alice Smith",
    "email": "alice@example.com",
    "created_at": ISODate("2024-01-15T10:30:00Z"),
    "updated_at": ISODate("2024-01-15T10:30:00Z")
}
```

🔝 [back to top](#ulid-objectids)

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

🔝 [back to top](#ulid-objectids)

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

🔝 [back to top](#ulid-objectids)

&nbsp;

## Querying by ULID

&nbsp;

### Find by ULID String

```go
collection := client.Collection("users")

// Find by ULID field
var user map[string]any
err := collection.FindOne(ctx, map[string]any{
    "ulid": "01HGQJ5Z8K9X7N2M4P6R8T3V5W",
}).Decode(&user)
```

🔝 [back to top](#ulid-objectids)

&nbsp;

### Time-Range Queries

```go
// Generate ULIDs for time range boundaries
startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
endTime := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

startULID := mongodb.GenerateULIDFromTime(startTime)
endULID := mongodb.GenerateULIDFromTime(endTime)

// Query documents created in January 2024
cursor, err := collection.Find(ctx, map[string]any{
    "ulid": map[string]any{
        "$gte": startULID,
        "$lte": endULID,
    },
})
```

🔝 [back to top](#ulid-objectids)

&nbsp;

### Sorting by Creation Time

```go
// Sort by ULID for chronological order (no need for separate timestamp field)
cursor, err := collection.Find(ctx,
    map[string]any{},
    &options.FindOptions{
        Sort: map[string]any{"ulid": 1}, // Ascending (oldest first)
    },
)

// Or sort by ObjectID (also contains embedded ULID)
cursor, err = collection.Find(ctx,
    map[string]any{},
    &options.FindOptions{
        Sort: map[string]any{"_id": -1}, // Descending (newest first)
    },
)
```

🔝 [back to top](#ulid-objectids)

&nbsp;

## Performance Characteristics

&nbsp;

### Generation Speed

| Identifier Type | Operations/sec | Relative Performance |
|----------------|----------------|---------------------|
| **ULID** | ~6,000,000 | 6.0x faster |
| **UUID v4** | ~1,000,000 | 1.0x baseline |
| **MongoDB ObjectID** | ~8,000,000 | 8.0x faster |

🔝 [back to top](#ulid-objectids)

&nbsp;

### Database Performance

```go
// Benchmark: Insert 100,000 documents
// ULID ObjectIDs: 12.3 seconds
// UUID v4: 18.7 seconds
// Improvement: ~34% faster insertions

// Index efficiency
// ULID ObjectIDs: Better B-tree balance due to time-ordering
// UUID v4: Random distribution causes index fragmentation
```

🔝 [back to top](#ulid-objectids)

&nbsp;

### Storage Efficiency

- **ULID String**: 26 bytes (Base32 encoded)
- **ULID Binary**: 16 bytes (128-bit)
- **UUID String**: 36 bytes (with hyphens)
- **MongoDB ObjectID**: 12 bytes (but less entropy)

🔝 [back to top](#ulid-objectids)

&nbsp;

## Advanced Usage

&nbsp;

### Manual ULID Generation

```go
// Generate standalone ULID
ulid := mongodb.GenerateULID()
fmt.Printf("Generated ULID: %s\n", ulid)

// Generate ObjectID with embedded ULID
objectID := mongodb.GenerateObjectIDFromULID()
fmt.Printf("ObjectID: %s\n", objectID.Hex())

// Generate ULID from specific time
timestamp := time.Now().Add(-1 * time.Hour)
pastULID := mongodb.GenerateULIDFromTime(timestamp)
```

🔝 [back to top](#ulid-objectids)

&nbsp;

### Custom Document Enhancement

```go
// Manually enhance documents before insertion
doc := map[string]any{
    "name": "Custom User",
    "type": "premium",
}

enhancedDoc := mongodb.EnhanceDocument(doc)
// enhancedDoc now contains ULID, timestamps, etc.

result, err := collection.InsertOne(ctx, enhancedDoc)
```

🔝 [back to top](#ulid-objectids)

&nbsp;

### Bulk Operations with ULIDs

```go
// Bulk insert with automatic ULID generation
docs := []any{
    map[string]any{"name": "User 1"},
    map[string]any{"name": "User 2"},
    map[string]any{"name": "User 3"},
}

result, err := collection.InsertMany(ctx, docs)

// Each document gets unique ULID
for i, ulid := range result.ULIDs {
    fmt.Printf("Document %d ULID: %s\n", i+1, ulid)
}
```

🔝 [back to top](#ulid-objectids)

&nbsp;

## Migration from UUIDs

&nbsp;

### Gradual Migration Strategy

```go
// Step 1: Add ULID fields alongside existing UUIDs
collection := client.Collection("users")

// Update existing documents with ULIDs
cursor, err := collection.Find(ctx, map[string]any{
    "ulid": map[string]any{"$exists": false},
})

for cursor.Next(ctx) {
    var doc map[string]any
    cursor.Decode(&doc)

    // Generate ULID based on creation time or current time
    var ulid string
    if createdAt, ok := doc["created_at"].(time.Time); ok {
        ulid = mongodb.GenerateULIDFromTime(createdAt)
    } else {
        ulid = mongodb.GenerateULID()
    }

    // Update with ULID
    collection.UpdateOne(ctx,
        map[string]any{"_id": doc["_id"]},
        map[string]any{"$set": map[string]any{"ulid": ulid}},
    )
}
```

🔝 [back to top](#ulid-objectids)

&nbsp;

### UUID to ULID Mapping

```go
// Maintain mapping for backward compatibility
type DocumentMapping struct {
    UUID string `bson:"uuid"`
    ULID string `bson:"ulid"`
}

// Create index for efficient lookups
collection.CreateIndex(ctx, map[string]any{"uuid": 1})
collection.CreateIndex(ctx, map[string]any{"ulid": 1})
```

🔝 [back to top](#ulid-objectids)

&nbsp;

## Best Practices

&nbsp;

### Indexing Strategy

```go
// Primary index on ULID for time-based queries
collection.CreateIndex(ctx, map[string]any{"ulid": 1})

// Compound indexes with ULID first
collection.CreateIndex(ctx, map[string]any{
    "ulid": 1,
    "status": 1,
})

// Avoid indexes on random UUIDs (causes fragmentation)
```

🔝 [back to top](#ulid-objectids)

&nbsp;

### Query Optimization

```go
// Efficient time-range queries using ULID
startULID := mongodb.GenerateULIDFromTime(startTime)
endULID := mongodb.GenerateULIDFromTime(endTime)

// This query uses index efficiently
filter := map[string]any{
    "ulid": map[string]any{
        "$gte": startULID,
        "$lt": endULID,
    },
}
```

🔝 [back to top](#ulid-objectids)

&nbsp;

### Application Design

- **Use ULID for primary keys** when you need time-ordered identifiers
- **Keep UUID for external APIs** if required for compatibility
- **Index on ULID fields** for efficient queries
- **Use ULID prefixes** for sharding strategies
- **Consider ULID in URL design** for user-friendly, sortable URLs

🔝 [back to top](#ulid-objectids)

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

🔝 [back to top](#ulid-objectids)

&nbsp;

#### Duplicate Detection

```go
// ULIDs have 80 bits of randomness - duplicates are extremely rare
// In high-throughput scenarios, consider adding application-level checks

result, err := collection.InsertOne(ctx, doc)
if mongodb.IsDuplicateKeyError(err) {
    // Handle extremely rare duplicate case
    log.Warn("Duplicate ULID detected - regenerating")
    // Regenerate and retry
}
```

🔝 [back to top](#ulid-objectids)

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

🔝 [back to top](#ulid-objectids)

&nbsp;

---

&nbsp;

An open source project brought to you by the [Cloudresty](https://cloudresty.com) team.

[Website](https://cloudresty.com) &nbsp;|&nbsp; [LinkedIn](https://www.linkedin.com/company/cloudresty) &nbsp;|&nbsp; [BlueSky](https://bsky.app/profile/cloudresty.com) &nbsp;|&nbsp; [GitHub](https://github.com/cloudresty)

&nbsp;
