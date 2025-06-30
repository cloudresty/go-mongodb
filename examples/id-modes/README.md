# ID Modes Example

[Home](../../README.md) &nbsp;/&nbsp; ID Modes Examples

&nbsp;

This example demonstrates the different ID generation strategies available in the go-mongodb package.

&nbsp;

## Available ID Modes

&nbsp;

### 1. ULID (Default)

- **Value**: `ulid`
- **Description**: Generates 26-character ULID strings with temporal ordering
- **Best for**: Applications needing time-based sorting and better database performance

🔝 [back to top](#id-modes-example)

&nbsp;

### 2. ObjectID

- **Value**: `objectid`
- **Description**: Generates standard MongoDB ObjectIDs
- **Best for**: Compatibility with existing MongoDB applications

🔝 [back to top](#id-modes-example)

&nbsp;

### 3. Custom

- **Value**: `custom`
- **Description**: No automatic ID generation; user provides their own `_id`
- **Best for**: Applications that need full control over document IDs

🔝 [back to top](#id-modes-example)

&nbsp;

## Configuration

&nbsp;

### Environment Variable

```bash
export MONGODB_ID_MODE=ulid      # Default - ULID strings
export MONGODB_ID_MODE=objectid  # MongoDB ObjectIDs
export MONGODB_ID_MODE=custom    # User-provided IDs
```

🔝 [back to top](#id-modes-example)

&nbsp;

### Programmatic Configuration

```go
config := &mongodb.Config{
    Host:     "localhost",
    Port:     27017,
    Database: "myapp",
    IDMode:   mongodb.IDModeULID,      // Default
    // IDMode:   mongodb.IDModeObjectID,  // ObjectID mode
    // IDMode:   mongodb.IDModeCustom,    // Custom mode
}

client, err := mongodb.ConnectWithConfig(config)
```

🔝 [back to top](#id-modes-example)

&nbsp;

## Running the Example

```bash
go run main.go
```

This will demonstrate all three ID modes and show how each one behaves.

🔝 [back to top](#id-modes-example)

&nbsp;

---

&nbsp;

An open source project brought to you by the [Cloudresty](https://cloudresty.com) team.

[Website](https://cloudresty.com) &nbsp;|&nbsp; [LinkedIn](https://www.linkedin.com/company/cloudresty) &nbsp;|&nbsp; [BlueSky](https://bsky.app/profile/cloudresty.com) &nbsp;|&nbsp; [GitHub](https://github.com/cloudresty)

&nbsp;
