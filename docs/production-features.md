# Production Features

[Home](../README.md) &nbsp;/&nbsp; [Docs](README.md) &nbsp;/&nbsp; Production Features

&nbsp;

This document covers all production-ready features designed for high-availability, fault-tolerant deployments.

&nbsp;

## Auto-Reconnection

Intelligent reconnection with exponential backoff for network resilience.

&nbsp;

### Basic Auto-Reconnection

```go
// Modern functional options pattern
client, err := mongodb.NewClient(
    mongodb.WithHosts("mongodb.example.com:27017"),
    mongodb.WithReconnectEnabled(true),
    mongodb.WithReconnectDelay(5*time.Second),
    mongodb.WithMaxReconnectDelay(1*time.Minute),
    mongodb.WithReconnectBackoff(2.0),
    mongodb.WithMaxReconnectAttempts(10),
)
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// Or use environment variables
client, err := mongodb.NewClient(mongodb.FromEnv())
if err != nil {
    log.Fatal(err)
}
defer client.Close()
```

🔝 [back to top](#production-features)

&nbsp;

### Environment Configuration

```bash
export MONGODB_HOSTS=mongodb.example.com:27017
export MONGODB_RECONNECT_ENABLED=true
export MONGODB_RECONNECT_DELAY=5s
export MONGODB_MAX_RECONNECT_DELAY=1m
export MONGODB_RECONNECT_BACKOFF=2.0
export MONGODB_MAX_RECONNECT_ATTEMPTS=10
```

🔝 [back to top](#production-features)

&nbsp;

### Reconnection Features

- **Exponential Backoff**: Intelligent delay progression to avoid overwhelming the server
- **Maximum Delay Cap**: Prevents excessively long wait times
- **Attempt Limiting**: Configurable maximum reconnection attempts
- **Connection State Tracking**: Monitor reconnection status and count
- **Automatic Recovery**: Seamless operation resumption after reconnection

🔝 [back to top](#production-features)

&nbsp;

## Health Checks

Comprehensive health monitoring for proactive issue detection.

&nbsp;

### Basic Health Checks

```go
import (
    "context"
    "log"

    "github.com/cloudresty/go-mongodb"
)

client, _ := mongodb.NewClient(mongodb.FromEnv())

// The single, reliable way to check health
ctx := context.Background()
err := client.Ping(ctx)
if err != nil {
    log.Printf("MongoDB connection failed: %v", err)
} else {
    log.Println("MongoDB connection is active and healthy")
}

// For detailed metrics, use Stats()
stats := client.Stats()
log.Printf("Cluster version: %s, Active connections: %d",
    stats.ServerVersion, stats.ActiveConnections)
```

🔝 [back to top](#production-features)

&nbsp;

### Automated Health Monitoring

```go
// Modern functional options pattern
client, err := mongodb.NewClient(
    mongodb.FromEnv(),
    mongodb.WithHealthCheckEnabled(true),
    mongodb.WithHealthCheckInterval(30*time.Second),
)
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// Health checks run automatically in the background
```

🔝 [back to top](#production-features)

&nbsp;

### Health Check Environment Variables

```bash
export MONGODB_HOSTS=mongodb.example.com:27017
export MONGODB_HEALTH_CHECK_ENABLED=true
export MONGODB_HEALTH_CHECK_INTERVAL=30s
```

🔝 [back to top](#production-features)

&nbsp;

### Health Check Features

- **Automated Monitoring**: Background health checks at configurable intervals
- **Connection Validation**: Ping operations to verify connectivity
- **Error Detection**: Early detection of connection issues
- **Status Reporting**: Detailed health status with error information
- **Reconnection Triggering**: Automatic reconnection on health failures

🔝 [back to top](#production-features)

&nbsp;

## Transaction Support

ACID transactions for data consistency and integrity.

&nbsp;

### Basic Transactions

```go
import (
    "context"
    "log"

    "github.com/cloudresty/go-mongodb"
    "github.com/cloudresty/go-mongodb/filter"
    "github.com/cloudresty/go-mongodb/update"
)

client, _ := mongodb.NewClient(mongodb.FromEnv())

err := client.WithTransaction(ctx, func(ctx context.Context) error {
    collection := client.Database("myapp").Collection("accounts")

    // Debit account using fluent builders
    _, err := collection.UpdateOne(ctx,
        filter.Eq("account_id", "account_1"),
        update.Inc("balance", -100), // ← Clean, safe, and fluent
    )
    if err != nil {
        return err
    }

    // Credit account using fluent builders
    _, err = collection.UpdateOne(ctx,
        filter.Eq("account_id", "account_2"),
        update.Inc("balance", 100), // ← Consistent and readable
    )
    return err
})

if err != nil {
    log.Printf("Transaction failed: %v", err)
} else {
    log.Println("✅ Transaction completed successfully")
}
```

🔝 [back to top](#production-features)

&nbsp;

### Advanced Transaction Configuration

```go
client, _ := mongodb.NewClient()

session, err := client.StartSession()
if err != nil {
    return err
}
defer session.EndSession(ctx)

err = session.StartTransaction(&mongodb.TransactionOptions{
    ReadConcern:    readconcern.Majority(),
    WriteConcern:   writeconcern.Majority(),
    ReadPreference: readpref.Primary(),
    MaxCommitTime:  time.Second * 30,
})

// Perform operations...

if err := session.CommitTransaction(ctx); err != nil {
    session.AbortTransaction(ctx)
    return err
}
```

🔝 [back to top](#production-features)

&nbsp;

### Transaction Features

- **ACID Compliance**: Full support for MongoDB transactions
- **Multi-Document Operations**: Atomic operations across multiple documents
- **Cross-Collection Support**: Transactions spanning multiple collections
- **Configurable Isolation**: Read/write concern configuration
- **Automatic Retry**: Built-in retry logic for transient failures
- **Timeout Protection**: Configurable transaction timeouts

🔝 [back to top](#production-features)

&nbsp;

## Timeout Configuration

Comprehensive timeout controls for production reliability.

&nbsp;

### Environment Variables

```bash
# Timeout environment variables
export MONGODB_HOSTS=mongodb.example.com:27017
export MONGODB_CONNECT_TIMEOUT=30s
export MONGODB_SERVER_SELECT_TIMEOUT=10s
export MONGODB_SOCKET_TIMEOUT=10s
```

🔝 [back to top](#production-features)

&nbsp;

### Programmatic Configuration

```go
// Modern functional options pattern
client, err := mongodb.NewClient(
    mongodb.FromEnv(),
    mongodb.WithConnectTimeout(30*time.Second),
    mongodb.WithServerSelectTimeout(10*time.Second),
    mongodb.WithSocketTimeout(10*time.Second),
)
if err != nil {
    log.Fatal(err)
}
defer client.Close()
```

🔝 [back to top](#production-features)

&nbsp;

### Timeout Types

| Timeout | Description | Default | Recommended |
|---------|-------------|---------|-------------|
| **Connect** | Initial connection establishment | 10s | 30s for production |
| **Server Selection** | Finding available server | 5s | 10s for replica sets |
| **Socket** | Individual operation timeout | 10s | Based on operation complexity |

🔝 [back to top](#production-features)

&nbsp;

### Detailed Timeout Configuration

```go
// Modern functional options pattern
client, err := mongodb.NewClient(
    mongodb.FromEnv(),

    // Connection timeouts
    mongodb.WithConnectTimeout(30*time.Second),      // Initial connection
    mongodb.WithServerSelectTimeout(10*time.Second), // Server discovery
    mongodb.WithSocketTimeout(60*time.Second),       // Operation timeout

    // Pool timeouts
    mongodb.WithMaxIdleTime(5*time.Minute),          // Connection idle timeout
    mongodb.WithMaxConnIdleTime(10*time.Minute),     // Maximum connection age
)
if err != nil {
    log.Fatal(err)
}
defer client.Close()
```

🔝 [back to top](#production-features)

&nbsp;

### Production Timeout Recommendations

```bash
# High-traffic production environment
export MONGODB_HOSTS=prod.mongodb.example.com:27017
export MONGODB_CONNECT_TIMEOUT=45s
export MONGODB_SERVER_SELECT_TIMEOUT=15s
export MONGODB_SOCKET_TIMEOUT=120s
export MONGODB_MAX_IDLE_TIME=10m
export MONGODB_MAX_CONN_IDLE_TIME=30m

# Low-latency environment
export MONGODB_HOSTS=fast.mongodb.example.com:27017
export MONGODB_CONNECT_TIMEOUT=10s
export MONGODB_SERVER_SELECT_TIMEOUT=5s
export MONGODB_SOCKET_TIMEOUT=30s
export MONGODB_MAX_IDLE_TIME=2m
export MONGODB_MAX_CONN_IDLE_TIME=5m
```

🔝 [back to top](#production-features)

&nbsp;

### Timeout Best Practices

- **Set realistic timeouts** based on your network conditions
- **Consider operation complexity** when setting socket timeouts
- **Monitor timeout errors** to identify infrastructure issues
- **Use different timeouts per environment** (dev vs. staging vs. production)
- **Account for replica set failover time** in server selection timeout

🔝 [back to top](#production-features)

&nbsp;

## Graceful Shutdown

Production-ready graceful shutdown with coordinated resource cleanup.

&nbsp;

### Basic Graceful Shutdown

```go
client, _ := mongodb.NewClient()

// Set up signal handling
shutdownManager := mongodb.NewShutdownManager(&mongodb.ShutdownConfig{
    Timeout: 30 * time.Second,
})

shutdownManager.SetupSignalHandler()
shutdownManager.Register(client)

// Application logic here...

// Wait for shutdown signal
shutdownManager.Wait()
```

🔝 [back to top](#production-features)

&nbsp;

### Advanced Coordinated Shutdown

```go
// Multiple clients and resources
clientA, _ := mongodb.NewClient()
clientB, _ := mongodb.NewClient(mongodb.FromEnvWithPrefix("SERVICE_B_"))

shutdownManager := mongodb.NewShutdownManager(&mongodb.ShutdownConfig{
    Timeout:           30 * time.Second,
    GracePeriod:       5 * time.Second,
    ForceKillTimeout:  10 * time.Second,
})

// Register all resources
shutdownManager.Register(clientA, clientB)
shutdownManager.SetupSignalHandler()

// Start background workers
go backgroundWorker(shutdownManager.Context())
go healthChecker(shutdownManager.Context())

// Wait for shutdown
shutdownManager.Wait()
```

🔝 [back to top](#production-features)

&nbsp;

### Shutdown Features

- **Signal Handling**: Automatic SIGINT/SIGTERM signal processing
- **In-Flight Tracking**: Waits for pending operations to complete
- **Timeout Protection**: Prevents indefinite waiting during shutdown
- **Component Coordination**: Unified shutdown across multiple clients
- **Zero Data Loss**: Ensures operation completion before exit

🔝 [back to top](#production-features)

&nbsp;

## Performance Characteristics

Optimized for high-throughput, low-latency operations.

&nbsp;

### Connection Pooling

```go
// Modern functional options pattern
client, err := mongodb.NewClient(
    mongodb.FromEnv(),
    mongodb.WithMaxPoolSize(200),                    // Maximum connections
    mongodb.WithMinPoolSize(10),                     // Minimum connections
    mongodb.WithMaxIdleTime(5*time.Minute),          // Idle connection timeout
    mongodb.WithMaxConnIdleTime(30*time.Minute),     // Maximum connection age
)
if err != nil {
    log.Fatal(err)
}
defer client.Close()
```

🔝 [back to top](#production-features)

&nbsp;

### Compression Configuration

```go
// Modern functional options pattern
client, err := mongodb.NewClient(
    mongodb.FromEnv(),
    mongodb.WithCompressionEnabled(true),
    mongodb.WithCompressionAlgorithm("zstd"), // snappy, zlib, zstd
)
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// Environment configuration
// export MONGODB_HOSTS=mongodb.example.com:27017
// export MONGODB_COMPRESSION_ENABLED=true
// export MONGODB_COMPRESSION_ALGORITHM=zstd
```

🔝 [back to top](#production-features)

&nbsp;

### Read/Write Preferences

```go
// Modern functional options pattern
client, err := mongodb.NewClient(
    mongodb.FromEnv(),
    mongodb.WithReadPreference("primaryPreferred"), // primary, primaryPreferred, secondary, secondaryPreferred, nearest
    mongodb.WithWriteConcern("majority"),           // majority, 1, 2, 3, etc.
    mongodb.WithReadConcern("local"),              // local, available, majority, linearizable
)
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// Environment configuration
// export MONGODB_HOSTS=mongodb.example.com:27017
// export MONGODB_READ_PREFERENCE=primaryPreferred
// export MONGODB_WRITE_CONCERN=majority
// export MONGODB_READ_CONCERN=local
```

🔝 [back to top](#production-features)

&nbsp;

### Performance Benchmarks

| Operation | Throughput | Latency (P99) | Notes |
|-----------|------------|---------------|-------|
| **Insert** | 50K ops/sec | <5ms | With ULID IDs |
| **Find** | 100K ops/sec | <2ms | Simple queries with indexes |
| **Update** | 40K ops/sec | <8ms | Single document updates |
| **Aggregate** | 10K ops/sec | <50ms | Complex pipelines |

Note: Benchmarks performed on MongoDB 7.0, 16 CPU cores, 32GB RAM

🔝 [back to top](#production-features)

&nbsp;

## Production Checklist

Essential items for production deployment:

**Completed Features:**

- **Implement proper logging** - emit library integrated with structured, high-performance logging
- **Configure appropriate timeouts** - connection, server selection, and socket timeouts
- **Implement graceful shutdown** - shutdown manager, in-flight tracking, signal handling
- **Environment-first configuration** - zero-config setup with MONGODB_* environment variables
- **Auto-reconnection** - intelligent retry with configurable backoff
- **Health checks** - automated monitoring and status reporting
- **ULID IDs** - high-performance, database-optimized identifiers
- **Transaction support** - ACID compliance with retry logic
- **Enhanced find operations** - QueryOptions integration with sort, limit, skip, projection
- **Pipeline builder** - fluent aggregation pipeline construction
- **Advanced query capabilities** - comprehensive sorting and pagination support

**Additional Recommended Items:**

- [ ] Set up monitoring and metrics
- [ ] Test failover scenarios
- [ ] Monitor memory usage
- [ ] Set up alerting for connection failures
- [ ] Configure load balancing for MongoDB cluster
- [ ] Implement backup and disaster recovery
- [ ] Performance testing under load
- [ ] Security audit and hardening

🔝 [back to top](#production-features)

&nbsp;

---

&nbsp;

An open source project brought to you by the [Cloudresty](https://cloudresty.com) team.

[Website](https://cloudresty.com) &nbsp;|&nbsp; [LinkedIn](https://www.linkedin.com/company/cloudresty) &nbsp;|&nbsp; [BlueSky](https://bsky.app/profile/cloudresty.com) &nbsp;|&nbsp; [GitHub](https://github.com/cloudresty) &nbsp;|&nbsp; [Docker Hub](https://hub.docker.com/u/cloudresty)

&nbsp;
