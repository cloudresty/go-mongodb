# Environment Configuration

[Home](../README.md) &nbsp;/&nbsp; [Docs](README.md) &nbsp;/&nbsp; Environment Configuration

&nbsp;

The package supports loading configuration from environment variables, making it ideal for containerized deployments and CI/CD pipelines.

&nbsp;

## Quick Start

```go
// Load with default MONGODB_ prefix
client, err := mongodb.NewClient()

// Load configuration and customize before use
config, err := mongodb.LoadConfig()
if err != nil {
    log.Fatal(err)
}

// Customize after loading
config.ConnectionName = "my-custom-service"
client, err := mongodb.NewClientWithConfig(config)
```

🔝 [back to top](#environment-configuration)

&nbsp;

## Custom Prefix

```go
// Use custom prefix (e.g., MYAPP_MONGODB_HOST instead of MONGODB_HOST)
client, err := mongodb.NewClientWithPrefix("MYAPP_")

// Load with custom prefix
config, err := mongodb.LoadConfigWithPrefix("MYAPP_")
```

🔝 [back to top](#environment-configuration)

&nbsp;

## Supported Environment Variables

&nbsp;

### Connection Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `MONGODB_HOST` | `localhost` | MongoDB host |
| `MONGODB_PORT` | `27017` | MongoDB port |
| `MONGODB_USERNAME` | `""` | MongoDB username |
| `MONGODB_PASSWORD` | `""` | MongoDB password |
| `MONGODB_DATABASE` | `app` | Default database name |
| `MONGODB_AUTH_DATABASE` | `admin` | Authentication database |
| `MONGODB_REPLICA_SET` | `""` | Replica set name |
| `MONGODB_CONNECTION_NAME` | `""` | Connection identifier |
| `MONGODB_APP_NAME` | `go-mongodb-app` | Application name for MongoDB logs |

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
MONGODB_HOST=mongodb.example.com
MONGODB_PORT=27017
MONGODB_USERNAME=myuser
MONGODB_PASSWORD=mypassword
MONGODB_DATABASE=production
MONGODB_ID_MODE=ulid
MONGODB_CONNECTION_NAME=my-production-service
MONGODB_MAX_POOL_SIZE=200
MONGODB_CONNECT_TIMEOUT=15s
MONGODB_COMPRESSION_ENABLED=true
MONGODB_COMPRESSION_ALGORITHM=zstd
MONGODB_READ_PREFERENCE=primaryPreferred
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
      MONGODB_HOST: mongodb
      MONGODB_DATABASE: myapp
      MONGODB_ID_MODE: ulid
      MONGODB_USERNAME: app_user
      MONGODB_PASSWORD: secure_password
      MONGODB_CONNECTION_NAME: my-app-instance
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
        - name: MONGODB_HOST
          value: "mongodb-service"
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
config, _ := mongodb.LoadConfig()

// Always builds URI from components using priority-based configuration:
// mongodb://user:pass@host:27017/database?authSource=admin&replicaSet=rs0
uri := config.BuildConnectionURI()
```

🔝 [back to top](#environment-configuration)

&nbsp;

## Configuration Examples

&nbsp;

### Development Environment

```bash
# Minimal development setup
export MONGODB_HOST=localhost
export MONGODB_DATABASE=myapp_dev
export MONGODB_ID_MODE=ulid
export MONGODB_LOG_LEVEL=debug
```

🔝 [back to top](#environment-configuration)

&nbsp;

### Production Environment

```bash
# Production configuration
export MONGODB_HOST=mongodb-cluster.example.com
export MONGODB_PORT=27017
export MONGODB_USERNAME=prod_user
export MONGODB_PASSWORD=secure_production_password
export MONGODB_DATABASE=myapp_production
export MONGODB_AUTH_DATABASE=admin
export MONGODB_REPLICA_SET=rs0
export MONGODB_ID_MODE=ulid
export MONGODB_CONNECTION_NAME=myapp-production
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
export PAYMENTS_MONGODB_HOST=mongodb-payments.internal
export PAYMENTS_MONGODB_DATABASE=payments
export PAYMENTS_MONGODB_ID_MODE=objectid
export PAYMENTS_MONGODB_CONNECTION_NAME=payments-service

# Service B configuration
export ORDERS_MONGODB_HOST=mongodb-orders.internal
export ORDERS_MONGODB_DATABASE=orders
export ORDERS_MONGODB_ID_MODE=ulid
export ORDERS_MONGODB_CONNECTION_NAME=orders-service

# Load in application
paymentsClient, _ := mongodb.NewClientWithPrefix("PAYMENTS_")
ordersClient, _ := mongodb.NewClientWithPrefix("ORDERS_")
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

---

&nbsp;

An open source project brought to you by the [Cloudresty](https://cloudresty.com) team.

[Website](https://cloudresty.com) &nbsp;|&nbsp; [LinkedIn](https://www.linkedin.com/company/cloudresty) &nbsp;|&nbsp; [BlueSky](https://bsky.app/profile/cloudresty.com) &nbsp;|&nbsp; [GitHub](https://github.com/cloudresty)

&nbsp;
