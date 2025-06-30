# Environment Variables

[Home](../README.md) &nbsp;/&nbsp; [Docs](README.md) &nbsp;/&nbsp; Environment Variables

&nbsp;

The go-mongodb package supports comprehensive configuration through environment variables, making it ideal for containerized and cloud-native deployments.

&nbsp;

## Core Connection Variables

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `MONGODB_HOSTS` | MongoDB host addresses (comma-separated) | `localhost:27017` | `mongo1:27017,mongo2:27017` |
| `MONGODB_USERNAME` | Authentication username | _(none)_ | `myuser` |
| `MONGODB_PASSWORD` | Authentication password | _(none)_ | `mypassword` |
| `MONGODB_DATABASE` | Default database name | `app` | `production` |
| `MONGODB_AUTH_DATABASE` | Authentication database | `admin` | `admin` |

🔝 [back to top](#environment-variables)

&nbsp;

## Application Settings

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `MONGODB_APP_NAME` | Application identifier for logging | `go-mongodb-app` | `my-service` |
| `MONGODB_ID_MODE` | ID generation strategy | `ulid` | `ulid`, `objectid`, `custom` |

🔝 [back to top](#environment-variables)

&nbsp;

## Connection Pool Settings

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `MONGODB_MAX_POOL_SIZE` | Maximum connections in pool | `100` | `50` |
| `MONGODB_MIN_POOL_SIZE` | Minimum connections in pool | `5` | `10` |
| `MONGODB_MAX_IDLE_TIME` | Connection idle timeout | `30m` | `15m` |

🔝 [back to top](#environment-variables)

&nbsp;

## Timeout Settings

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `MONGODB_TIMEOUT` | Default operation timeout | `30s` | `60s` |
| `MONGODB_CONNECT_TIMEOUT` | Connection establishment timeout | `10s` | `5s` |
| `MONGODB_SERVER_SELECTION_TIMEOUT` | Server selection timeout | `30s` | `15s` |

🔝 [back to top](#environment-variables)

&nbsp;

## Replica Set Configuration

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `MONGODB_REPLICA_SET` | Replica set name | _(none)_ | `rs0` |
| `MONGODB_READ_PREFERENCE` | Read preference | `primary` | `secondaryPreferred` |
| `MONGODB_WRITE_CONCERN` | Write concern | `majority` | `1` |

🔝 [back to top](#environment-variables)

&nbsp;

## Security Settings

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `MONGODB_TLS` | Enable TLS/SSL | `false` | `true` |
| `MONGODB_TLS_CERT_FILE` | TLS certificate file path | _(none)_ | `/certs/client.pem` |
| `MONGODB_TLS_KEY_FILE` | TLS key file path | _(none)_ | `/certs/client-key.pem` |
| `MONGODB_TLS_CA_FILE` | TLS CA certificate file path | _(none)_ | `/certs/ca.pem` |

🔝 [back to top](#environment-variables)

&nbsp;

## Health Check Settings

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `MONGODB_HEALTH_CHECK_ENABLED` | Enable automatic health checks | `true` | `false` |
| `MONGODB_HEALTH_CHECK_INTERVAL` | Health check interval | `30s` | `60s` |

🔝 [back to top](#environment-variables)

&nbsp;

## Reconnection Settings

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `MONGODB_RECONNECT_ENABLED` | Enable automatic reconnection | `true` | `false` |
| `MONGODB_RECONNECT_DELAY` | Initial reconnect delay | `5s` | `10s` |
| `MONGODB_MAX_RECONNECT_DELAY` | Maximum reconnect delay | `1m` | `5m` |
| `MONGODB_RECONNECT_BACKOFF` | Backoff multiplier | `2.0` | `1.5` |
| `MONGODB_MAX_RECONNECT_ATTEMPTS` | Maximum reconnection attempts | `10` | `5` |

🔝 [back to top](#environment-variables)

&nbsp;

## Example Configurations

&nbsp;

### Development Environment

```bash
# .env file for development
MONGODB_HOSTS=localhost:27017
MONGODB_DATABASE=myapp_dev
MONGODB_APP_NAME=myapp-dev
MONGODB_ID_MODE=ulid
MONGODB_TIMEOUT=30s
```

🔝 [back to top](#environment-variables)

&nbsp;

### Production Environment

```bash
# Production environment variables
MONGODB_HOSTS=mongo1.prod:27017,mongo2.prod:27017,mongo3.prod:27017
MONGODB_USERNAME=produser
MONGODB_PASSWORD=secure_password_from_secret
MONGODB_DATABASE=production
MONGODB_APP_NAME=myapp-prod
MONGODB_REPLICA_SET=rs0
MONGODB_READ_PREFERENCE=secondaryPreferred
MONGODB_WRITE_CONCERN=majority
MONGODB_TLS=true
MONGODB_MAX_POOL_SIZE=200
MONGODB_TIMEOUT=60s
MONGODB_HEALTH_CHECK_INTERVAL=30s
```

🔝 [back to top](#environment-variables)

&nbsp;

### Cloud/Atlas Environment

```bash
# MongoDB Atlas configuration
MONGODB_HOSTS=cluster0.mongodb.net:27017
MONGODB_USERNAME=atlasuser
MONGODB_PASSWORD=atlas_password
MONGODB_DATABASE=myapp
MONGODB_TLS=true
MONGODB_AUTH_DATABASE=admin
MONGODB_APP_NAME=myapp-cloud
```

🔝 [back to top](#environment-variables)

&nbsp;

### Docker Environment

```bash
# Docker Compose .env
MONGODB_HOSTS=mongodb:27017
MONGODB_DATABASE=app
MONGODB_APP_NAME=dockerized-app
MONGODB_MAX_POOL_SIZE=50
MONGODB_TIMEOUT=30s
```

🔝 [back to top](#environment-variables)

&nbsp;

### Kubernetes Environment

```yaml
# ConfigMap for Kubernetes
apiVersion: v1
kind: ConfigMap
metadata:
  name: mongodb-config
data:
  MONGODB_HOSTS: "mongodb-service:27017"
  MONGODB_DATABASE: "production"
  MONGODB_APP_NAME: "k8s-app"
  MONGODB_MAX_POOL_SIZE: "100"
  MONGODB_TIMEOUT: "60s"
  MONGODB_READ_PREFERENCE: "secondaryPreferred"
```

🔝 [back to top](#environment-variables)

&nbsp;

## Custom Prefixes

You can use custom prefixes for environment variables to avoid conflicts:

```go
// Use custom prefix
client, err := mongodb.NewClient(
    mongodb.FromEnvWithPrefix("MYAPP"),
)
```

This will look for variables like:
- `MYAPP_MONGODB_HOSTS`
- `MYAPP_MONGODB_DATABASE`
- `MYAPP_MONGODB_USERNAME`
- etc.

🔝 [back to top](#environment-variables)

&nbsp;

## Variable Loading Order

The package loads configuration in this order (later values override earlier ones):

1. Package defaults
2. Environment variables (with optional prefix)
3. Functional options passed to `NewClient()`

```go
// Example: Environment sets database, but override with option
client, err := mongodb.NewClient(
    mongodb.FromEnv(),                    // Loads MONGODB_DATABASE=prod
    mongodb.WithDatabase("test_override"), // Overrides to use "test_override"
)
```

🔝 [back to top](#environment-variables)

&nbsp;

## Validation

The package validates environment variables at startup and provides clear error messages:

```bash
# Invalid timeout format
MONGODB_TIMEOUT=invalid

# Error: invalid duration format for MONGODB_TIMEOUT: time: invalid duration "invalid"
```

```bash
# Invalid ID mode
MONGODB_ID_MODE=invalid

# Error: invalid ID mode "invalid", must be one of: ulid, objectid, custom
```

🔝 [back to top](#environment-variables)

&nbsp;

## Best Practices

&nbsp;

### Security

- **Never commit credentials** to version control
- **Use secrets management** in production (Kubernetes secrets, AWS Secrets Manager, etc.)
- **Rotate passwords regularly** and update environment variables accordingly
- **Enable TLS** in production environments

🔝 [back to top](#environment-variables)

&nbsp;

### Performance

- **Set appropriate pool sizes** based on your application's concurrency needs
- **Tune timeouts** based on your network latency and operation complexity
- **Use read preferences** to distribute load in replica set deployments
- **Enable health checks** for automatic problem detection

🔝 [back to top](#environment-variables)

&nbsp;

### Monitoring

- **Set meaningful app names** for easier log correlation
- **Enable health checks** for monitoring integration
- **Use structured logging** with the app name for better observability

🔝 [back to top](#environment-variables)

&nbsp;

---

&nbsp;

An open source project brought to you by the [Cloudresty](https://cloudresty.com) team.

[Website](https://cloudresty.com) &nbsp;|&nbsp; [LinkedIn](https://www.linkedin.com/company/cloudresty) &nbsp;|&nbsp; [BlueSky](https://bsky.app/profile/cloudresty.com) &nbsp;|&nbsp; [GitHub](https://github.com/cloudresty)

&nbsp;
