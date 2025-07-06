# Environment Configuration

[Home](../README.md) &nbsp;/&nbsp; [Docs](README.md) &nbsp;/&nbsp; Environment Configuration

&nbsp;

The package supports loading configuration from environment variables, making it ideal for containerized deployments and CI/CD pipelines.

&nbsp;

## Quick Start

```go
// Load client directly from environment variables
client, err := mongodb.NewClient(mongodb.FromEnv())
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// Override specific values while loading from environment
client, err := mongodb.NewClient(
    mongodb.FromEnv(),
    mongodb.WithMaxPoolSize(50),              // Override environment value
    mongodb.WithAppName("billing-api-v1.2"),  // For MongoDB server logs (DBA visibility)
    mongodb.WithConnectionName("api-primary"), // For application logs (dev visibility)
)
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// Use the connection name in your application logs
log.Printf("[%s] Processing payment...", client.Name())
```

🔝 [back to top](#environment-configuration)

&nbsp;

## Custom Prefix

```go
// Use custom prefix (e.g., MYAPP_MONGODB_HOSTS instead of MONGODB_HOSTS)
client, err := mongodb.NewClient(mongodb.FromEnvWithPrefix("MYAPP_"))
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// Combine custom prefix with additional options
client, err := mongodb.NewClient(
    mongodb.FromEnvWithPrefix("MYAPP_"),
    mongodb.WithAppName("billing-service"),           // For MongoDB server logs
    mongodb.WithConnectionName("main-processor"),     // For application logs
    mongodb.WithMaxPoolSize(100),
)
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// Log with connection identifier
log.Printf("[%s] Service started successfully", client.Name())
```

🔝 [back to top](#environment-configuration)

&nbsp;

## Supported Environment Variables

&nbsp;

### Connection Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `MONGODB_HOSTS` | `localhost:27017` | MongoDB hosts (comma-separated for clusters) |
| `MONGODB_USERNAME` | `""` | MongoDB username |
| `MONGODB_PASSWORD` | `""` | MongoDB password |
| `MONGODB_DATABASE` | `app` | Default database name |
| `MONGODB_AUTH_DATABASE` | `admin` | Authentication database |
| `MONGODB_REPLICA_SET` | `""` | Replica set name |
| `MONGODB_CONNECTION_NAME` | `""` | Connection identifier |
| `MONGODB_APP_NAME` | `go-mongodb-app` | Application name for MongoDB logs |
| `MONGODB_DIRECT_CONNECTION` | `false` | Enable direct connection mode (bypasses replica set discovery) |

🔝 [back to top](#environment-configuration)

&nbsp;

### Security Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `MONGODB_TLS_ENABLED` | `false` | Enable TLS/SSL connection |
| `MONGODB_TLS_INSECURE` | `false` | Skip TLS certificate verification (dev only) |
| `MONGODB_TLS_CA_FILE` | `""` | Path to CA certificate file |
| `MONGODB_TLS_CERT_FILE` | `""` | Path to client certificate file |
| `MONGODB_TLS_KEY_FILE` | `""` | Path to client private key file |

🔝 [back to top](#environment-configuration)

&nbsp;

### Document Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `MONGODB_ID_MODE` | `ulid` | ID generation strategy (ulid, objectid, custom) |

🔝 [back to top](#environment-configuration)

&nbsp;

### Pool Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `MONGODB_MAX_POOL_SIZE` | `100` | Maximum connections in pool |
| `MONGODB_MIN_POOL_SIZE` | `5` | Minimum connections in pool |
| `MONGODB_MAX_IDLE_TIME` | `5m` | Maximum connection idle time |
| `MONGODB_MAX_CONN_IDLE_TIME` | `10m` | Maximum connection idle time |

🔝 [back to top](#environment-configuration)

&nbsp;

### Timeout Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `MONGODB_CONNECT_TIMEOUT` | `10s` | Initial connection timeout |
| `MONGODB_SERVER_SELECT_TIMEOUT` | `5s` | Server selection timeout |
| `MONGODB_SOCKET_TIMEOUT` | `10s` | Socket operation timeout |

🔝 [back to top](#environment-configuration)

&nbsp;

### Reconnection Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `MONGODB_RECONNECT_ENABLED` | `true` | Enable auto-reconnection |
| `MONGODB_RECONNECT_DELAY` | `5s` | Initial reconnection delay |
| `MONGODB_MAX_RECONNECT_DELAY` | `1m` | Maximum reconnection delay |
| `MONGODB_RECONNECT_BACKOFF` | `2.0` | Reconnection backoff multiplier |
| `MONGODB_MAX_RECONNECT_ATTEMPTS` | `10` | Maximum reconnection attempts |

🔝 [back to top](#environment-configuration)

&nbsp;

### Health Check Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `MONGODB_HEALTH_CHECK_ENABLED` | `true` | Enable health checks |
| `MONGODB_HEALTH_CHECK_INTERVAL` | `30s` | Health check interval |

🔝 [back to top](#environment-configuration)

&nbsp;

### Performance Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `MONGODB_COMPRESSION_ENABLED` | `true` | Enable compression |
| `MONGODB_COMPRESSION_ALGORITHM` | `snappy` | Compression algorithm (snappy, zlib, zstd) |
| `MONGODB_READ_PREFERENCE` | `primary` | Read preference |
| `MONGODB_WRITE_CONCERN` | `majority` | Write concern |
| `MONGODB_READ_CONCERN` | `local` | Read concern |

🔝 [back to top](#environment-configuration)

&nbsp;

### Logging Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `MONGODB_LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |
| `MONGODB_LOG_FORMAT` | `json` | Log format (json, text) |

🔝 [back to top](#environment-configuration)

&nbsp;

## Environment File Support

Create a `.env` file in your project root:

```bash
# Basic connection
MONGODB_HOSTS=mongodb.example.com:27017
MONGODB_USERNAME=myuser
MONGODB_PASSWORD=mypassword
MONGODB_DATABASE=production

# Cluster connection
MONGODB_HOSTS=mongo1.cluster.com:27017,mongo2.cluster.com:27017,mongo3.cluster.com:27017
MONGODB_REPLICA_SET=rs0

# Security (production)
MONGODB_TLS_ENABLED=true
MONGODB_TLS_CA_FILE=/path/to/ca.pem

# Configuration
MONGODB_ID_MODE=ulid
MONGODB_CONNECTION_NAME=my-production-service
MONGODB_MAX_POOL_SIZE=200
MONGODB_CONNECT_TIMEOUT=15s
MONGODB_COMPRESSION_ENABLED=true
MONGODB_COMPRESSION_ALGORITHM=zstd
MONGODB_READ_PREFERENCE=primaryPreferred
MONGODB_DIRECT_CONNECTION=false
```

🔝 [back to top](#environment-configuration)

&nbsp;

## Docker/Kubernetes Integration

&nbsp;

### Docker Compose

```yaml
version: '3.8'
services:
  my-app:
    image: my-app:latest
    environment:
      MONGODB_HOSTS: mongodb:27017
      MONGODB_DATABASE: myapp
      MONGODB_ID_MODE: ulid
      MONGODB_USERNAME: app_user
      MONGODB_PASSWORD: secure_password
      MONGODB_CONNECTION_NAME: my-app-instance
      MONGODB_DIRECT_CONNECTION: false
    depends_on:
      - mongodb

  mongodb:
    image: mongo:7
    environment:
      MONGO_INITDB_ROOT_USERNAME: admin
      MONGO_INITDB_ROOT_PASSWORD: password
      MONGO_INITDB_DATABASE: myapp
```

🔝 [back to top](#environment-configuration)

&nbsp;

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  template:
    spec:
      containers:
      - name: my-app
        image: my-app:latest
        env:
        - name: MONGODB_HOSTS
          value: "mongodb-service:27017"
        - name: MONGODB_DATABASE
          value: "myapp"
        - name: MONGODB_ID_MODE
          value: "ulid"
        - name: MONGODB_USERNAME
          valueFrom:
            secretKeyRef:
              name: mongodb-secret
              key: username
        - name: MONGODB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: mongodb-secret
              key: password
        - name: MONGODB_CONNECTION_NAME
          value: "my-app-pod"
```

🔝 [back to top](#environment-configuration)

&nbsp;

## URI Construction

Environment variables automatically construct proper MongoDB URIs with priority-based configuration:

**Priority Order:**

1. **Code defaults** - Sensible defaults for development
2. **Code defaults + environment variables** - Environment variables override specific defaults
3. **All environment variables** - When all required env vars are set, they take precedence

```go
// Environment variables are automatically processed when creating a client
client, _ := mongodb.NewClient(mongodb.FromEnv())

// The client internally constructs proper MongoDB URIs from environment variables:
// mongodb://user:pass@host:27017/database?authSource=admin&replicaSet=rs0
// All configuration is handled automatically with priority-based overrides
```

🔝 [back to top](#environment-configuration)

&nbsp;

## Configuration Examples

&nbsp;

### Development Environment

```bash
# Minimal development setup
export MONGODB_HOSTS=localhost:27017
export MONGODB_DATABASE=myapp_dev
export MONGODB_ID_MODE=ulid
export MONGODB_LOG_LEVEL=debug
```

🔝 [back to top](#environment-configuration)

&nbsp;

### Production Environment

```bash
# Production cluster configuration
export MONGODB_HOSTS=mongodb1.cluster.com:27017,mongodb2.cluster.com:27017,mongodb3.cluster.com:27017
export MONGODB_USERNAME=prod_user
export MONGODB_PASSWORD=secure_production_password
export MONGODB_DATABASE=myapp_production
export MONGODB_AUTH_DATABASE=admin
export MONGODB_REPLICA_SET=rs0
export MONGODB_ID_MODE=ulid
export MONGODB_CONNECTION_NAME=myapp-production

# Security configuration
export MONGODB_TLS_ENABLED=true
export MONGODB_TLS_CA_FILE=/etc/ssl/certs/mongodb-ca.pem

# Performance configuration
export MONGODB_MAX_POOL_SIZE=200
export MONGODB_MIN_POOL_SIZE=10
export MONGODB_CONNECT_TIMEOUT=30s
export MONGODB_COMPRESSION_ENABLED=true
export MONGODB_COMPRESSION_ALGORITHM=zstd
export MONGODB_READ_PREFERENCE=primaryPreferred
export MONGODB_WRITE_CONCERN=majority
export MONGODB_HEALTH_CHECK_ENABLED=true
export MONGODB_LOG_LEVEL=info
export MONGODB_LOG_FORMAT=json
```

🔝 [back to top](#environment-configuration)

&nbsp;

### Multi-Service Environment

```bash
# Service A configuration
export PAYMENTS_MONGODB_HOSTS=mongodb-payments.internal:27017
export PAYMENTS_MONGODB_DATABASE=payments
export PAYMENTS_MONGODB_ID_MODE=objectid
export PAYMENTS_MONGODB_CONNECTION_NAME=payments-service

# Service B configuration
export ORDERS_MONGODB_HOSTS=mongodb-orders.internal:27017
export ORDERS_MONGODB_DATABASE=orders
export ORDERS_MONGODB_ID_MODE=ulid
export ORDERS_MONGODB_CONNECTION_NAME=orders-service

# Load in application
paymentsClient, _ := mongodb.NewClient(mongodb.FromEnvWithPrefix("PAYMENTS_"))
ordersClient, _ := mongodb.NewClient(mongodb.FromEnvWithPrefix("ORDERS_"))
```

🔝 [back to top](#environment-configuration)

&nbsp;

## Best Practices

&nbsp;

### Security

- **Never hardcode credentials** - Always use environment variables or secrets
- **Use authentication database** - Set `MONGODB_AUTH_DATABASE` appropriately
- **Enable TLS in production** - Include TLS options in your URI
- **Rotate passwords regularly** - Update environment variables during deployments

🔝 [back to top](#environment-configuration)

&nbsp;

### Performance

- **Size connection pools appropriately** - Set `MONGODB_MAX_POOL_SIZE` based on your workload
- **Configure timeouts** - Set realistic timeouts for your network conditions
- **Enable compression** - Use compression for network-bound workloads
- **Choose read preferences wisely** - Balance consistency and performance needs

🔝 [back to top](#environment-configuration)

&nbsp;

### Monitoring

- **Use descriptive connection names** - Set `MONGODB_CONNECTION_NAME` for better monitoring
- **Enable health checks** - Monitor connection health with `MONGODB_HEALTH_CHECK_ENABLED`
- **Configure structured logging** - Use JSON logging in production environments
- **Set appropriate log levels** - Use `debug` for development, `info` for production

🔝 [back to top](#environment-configuration)

&nbsp;

## Integration with Fluent Builders

The environment configuration works seamlessly with the modern fluent query and update builders:

&nbsp;

### Observability-First Configuration

This example demonstrates the powerful combination of environment configuration with dual-level observability:

```go
package main

import (
    "context"
    "log"

    "github.com/cloudresty/go-mongodb"
    "github.com/cloudresty/go-mongodb/filter"
    "github.com/cloudresty/go-mongodb/update"
)

func main() {
    // This client connects to a read-only replica for reporting.
    // It identifies itself to MongoDB as the "billing-api" but locally
    // as the "reporting-read-replica" instance.
    client, err := mongodb.NewClient(
        mongodb.FromEnv(), // Load shared credentials from environment

        // For the DBA: "Who is connecting to me?"
        mongodb.WithAppName("billing-api-v1.2"),

        // For my own logs: "Which client instance is this?"
        mongodb.WithConnectionName("reporting-read-replica"),

        mongodb.WithReadPreference(mongodb.SecondaryPreferred),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // The client's local name is used for application logging
    log.Printf("[%s] Starting report generation...", client.Name())

    ctx := context.Background()
    db := client.Database("") // Uses MONGODB_DATABASE from environment
    transactions := db.Collection("transactions")

    // Use fluent builders with observable client
    recentTransactions := filter.Gte("created_at", time.Now().Add(-24*time.Hour)).
        And(filter.Eq("status", "completed"))

    cursor, err := transactions.Find(ctx, recentTransactions)
    if err != nil {
        log.Printf("[%s] Failed to query transactions: %v", client.Name(), err)
        return
    }
    defer cursor.Close(ctx)

    count := 0
    for cursor.Next(ctx) {
        count++
    }

    log.Printf("[%s] Processed %d transactions in daily report", client.Name(), count)
}
```

**Key Benefits:**

- **`mongodb.WithAppName("billing-api-v1.2")`** - MongoDB server logs show this name, helping DBAs identify which service is connecting
- **`mongodb.WithConnectionName("reporting-read-replica")`** - Your application logs use this name via `client.Name()`, helping developers identify which client instance
- **Environment configuration** - Shared connection details (host, credentials, database) loaded automatically
- **Functional options** - Instance-specific overrides (read preference, pool size) applied cleanly

🔝 [back to top](#environment-configuration)

&nbsp;

### Complete Example with Fluent APIs

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/cloudresty/go-mongodb"
    "github.com/cloudresty/go-mongodb/filter"
    "github.com/cloudresty/go-mongodb/update"
)

func main() {
    // Create client from environment variables
    client, err := mongodb.NewClient(mongodb.FromEnv())
    if err != nil {
        log.Fatal("Failed to create client:", err)
    }
    defer client.Close()

    // Get database and collection - names from environment
    db := client.Database("") // Uses MONGODB_DATABASE from environment
    users := db.Collection("users")

    ctx := context.Background()

    // Use fluent filter builder with environment-configured client
    activeUsers := filter.Eq("status", "active").
        And(filter.Gt("last_login", time.Now().Add(-30*24*time.Hour)))

    // Use fluent update builder
    updateLastSeen := update.Set("last_seen", time.Now()).
        Inc("visit_count", 1)

    // Execute operations
    result, err := users.UpdateMany(ctx, activeUsers, updateLastSeen)
    if err != nil {
        log.Fatal("Update failed:", err)
    }

    log.Printf("Updated %d active users", result.ModifiedCount)
}
```

🔝 [back to top](#environment-configuration)

&nbsp;

### Multi-Service Configuration with Builders

```go
// Configure multiple services with different environment prefixes
paymentsClient, _ := mongodb.NewClient(mongodb.FromEnvWithPrefix("PAYMENTS_"))
ordersClient, _ := mongodb.NewClient(mongodb.FromEnvWithPrefix("ORDERS_"))

// Each client can use the same fluent builders
paymentFilter := filter.Eq("status", "pending").
    And(filter.Lt("created_at", time.Now().Add(-1*time.Hour)))

orderUpdate := update.Set("processed_at", time.Now()).
    Set("status", "completed")

// Operations use respective environment configurations
paymentsCollection := paymentsClient.Database("").Collection("transactions")
ordersCollection := ordersClient.Database("").Collection("orders")

// Both use the same fluent API despite different configurations
_, err1 := paymentsCollection.UpdateMany(ctx, paymentFilter, orderUpdate)
_, err2 := ordersCollection.UpdateMany(ctx, filter.Eq("payment_id", "abc123"), orderUpdate)
```

🔝 [back to top](#environment-configuration)

&nbsp;

## Connection Examples

&nbsp;

### Single Host

```bash
# Simple single host
export MONGODB_HOSTS=localhost:27017

# Single host with custom port
export MONGODB_HOSTS=mongodb.example.com:27018

# Single host with authentication
export MONGODB_HOSTS=mongodb.example.com:27017
export MONGODB_USERNAME=myuser
export MONGODB_PASSWORD=secretpass

# Single host with TLS
export MONGODB_HOSTS=mongodb.example.com:27017
export MONGODB_TLS_ENABLED=true
export MONGODB_USERNAME=prod_user
export MONGODB_PASSWORD=prod_password
```

🔝 [back to top](#environment-configuration)

&nbsp;

### Multiple Hosts (Cluster)

```bash
# Multiple hosts on same port
export MONGODB_HOSTS=mongo1.cluster.com:27017,mongo2.cluster.com:27017,mongo3.cluster.com:27017

# Multiple hosts on different ports
export MONGODB_HOSTS=mongo1.example.com:27017,mongo2.example.com:27018,mongo3.example.com:27019

# Multiple hosts with authentication and replica set
export MONGODB_HOSTS=node1.cluster.com:27017,node2.cluster.com:27017,node3.cluster.com:27017
export MONGODB_USERNAME=cluster_user
export MONGODB_PASSWORD=cluster_password
export MONGODB_REPLICA_SET=rs0
export MONGODB_TLS_ENABLED=true
```

🔝 [back to top](#environment-configuration)

&nbsp;

### MongoDB Atlas

```bash
# MongoDB Atlas cluster
export MONGODB_HOSTS=cluster0-shard-00-00.abc123.mongodb.net:27017,cluster0-shard-00-01.abc123.mongodb.net:27017,cluster0-shard-00-02.abc123.mongodb.net:27017
export MONGODB_USERNAME=atlas_user
export MONGODB_PASSWORD=atlas_password
export MONGODB_TLS_ENABLED=true
export MONGODB_AUTH_DATABASE=admin
export MONGODB_REPLICA_SET=atlas-cluster-shard-0
```

🔝 [back to top](#environment-configuration)

&nbsp;

## Environment Variables Reference

The complete list of environment variables is also available in the [Environment Variables](environment-variables.md) documentation.

🔝 [back to top](#environment-configuration)

&nbsp;

---

&nbsp;

An open source project brought to you by the [Cloudresty](https://cloudresty.com) team.

[Website](https://cloudresty.com) &nbsp;|&nbsp; [LinkedIn](https://www.linkedin.com/company/cloudresty) &nbsp;|&nbsp; [BlueSky](https://bsky.app/profile/cloudresty.com) &nbsp;|&nbsp; [GitHub](https://github.com/cloudresty)

&nbsp;
