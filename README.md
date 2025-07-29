# Go MongoDB

[Home](README.md) &nbsp;/

&nbsp;

A modern, production-ready Go package for MongoDB operations with environment-first configuration, ULID IDs, and comprehensive production features.

&nbsp;

[![Go Reference](https://pkg.go.dev/badge/github.com/cloudresty/go-mongodb.svg)](https://pkg.go.dev/github.com/cloudresty/go-mongodb)
[![Go Tests](https://github.com/cloudresty/go-mongodb/actions/workflows/ci.yaml/badge.svg)](https://github.com/cloudresty/go-mongodb/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/cloudresty/go-mongodb)](https://goreportcard.com/report/github.com/cloudresty/go-mongodb)
[![GitHub Tag](https://img.shields.io/github/v/tag/cloudresty/go-mongodb?label=Version)](https://github.com/cloudresty/go-mongodb/tags)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

&nbsp;

## Table of Contents

- [Key Features](#key-features)
- [Quick Start](#quick-start)
  - [Installation](#installation)
  - [Basic Usage](#basic-usage)
  - [Environment Configuration](#environment-configuration)
- [Documentation](#documentation)
- [Why This Package?](#why-this-package)
- [Production Usage](#production-usage)
- [Requirements](#requirements)
- [Contributing](#contributing)
- [Security](#security)
- [License](#license)

&nbsp;

## Key Features

- **Environment-First**: Configure via environment variables for cloud-native deployments
- **ULID IDs**: 6x faster generation, database-optimized, lexicographically sortable
- **Auto-Reconnection**: Intelligent retry with configurable backoff
- **Production-Ready**: Graceful shutdown, timeouts, health checks, transaction support
- **Pluggable Logging**: Silent by default, integrate with any logging framework
- **High Performance**: Optimized for throughput with efficient ULID generation
- **Fully Tested**: Comprehensive test coverage with CI/CD pipeline

🔝 [back to top](#go-mongodb)

&nbsp;

## Quick Start

&nbsp;

### Installation

```bash
go get github.com/cloudresty/go-mongodb
```

🔝 [back to top](#go-mongodb)

&nbsp;

### Basic Usage

```go
package main

import (
    "context"
    "log"
    "github.com/cloudresty/go-mongodb"
    "github.com/cloudresty/go-mongodb/filter"
    "github.com/cloudresty/go-mongodb/update"
)

type User struct {
    Name   string `bson:"name" json:"name"`
    Email  string `bson:"email" json:"email"`
    Status string `bson:"status" json:"status"`
}

func main() {
    // Client - uses MONGODB_* environment variables
    client, err := mongodb.NewClient()
    if err != nil {
        panic(err)
    }
    defer client.Close()

    // Insert a document with auto-generated ULID ID using type-safe struct
    collection := client.Collection("users")
    user := User{
        Name:  "John Doe",
        Email: "john@example.com",
    }

    result, err := collection.InsertOne(context.Background(), user)
    if err != nil {
        log.Fatal(err)
    }

    // Find documents using type-safe filter
    var foundUser User
    err = collection.FindOne(context.Background(),
        filter.Eq("email", "john@example.com")).Decode(&foundUser)
    if err != nil {
        log.Fatal(err)
    }

    // Update documents using type-safe update builder
    _, err = collection.UpdateOne(context.Background(),
        filter.Eq("email", "john@example.com"),
        update.Set("status", "active"))
    if err != nil {
        log.Fatal(err)
    }
}
```

🔝 [back to top](#go-mongodb)

&nbsp;

### Environment Configuration

Set environment variables for your deployment:

```bash
export MONGODB_HOSTS=localhost:27017
export MONGODB_DATABASE=myapp
export MONGODB_CONNECTION_NAME=my-service
```

🔝 [back to top](#go-mongodb)

&nbsp;

### Pluggable Logging

The library is **silent by default** (zero-allocation) but supports pluggable logging for any framework:

```go
// Silent by default - no logging output
client, err := mongodb.NewClient()

// Integrate with your logging framework
client, err := mongodb.NewClient(
    mongodb.WithLogger(YourLoggerAdapter{}),
)
```

**Standard Logger Example:**
```go
type StandardLoggerAdapter struct {
    logger *log.Logger
}

func (s StandardLoggerAdapter) Info(msg string, fields ...any) {
    s.logger.Printf("INFO: %s %v", msg, fields)
}
// ... implement Warn, Error, Debug methods
```

**Emit Logger Example:**
See `examples/custom-logger-emit/` for a complete integration with [cloudresty/emit](https://github.com/cloudresty/emit).

🔝 [back to top](#go-mongodb)

&nbsp;

## Documentation

| Document | Description |
|----------|-------------|
| [API Reference](docs/api-reference.md) | Complete function reference and usage patterns |
| [Environment Configuration](docs/environment-configuration.md) | Environment variables and deployment configurations |
| [Production Features](docs/production-features.md) | Auto-reconnection, graceful shutdown, health checks, transactions |
| [ID Generation](docs/id-generation.md) | High-performance, database-optimized document identifiers |
| [Examples](docs/examples.md) | Comprehensive examples and usage patterns |

🔝 [back to top](#go-mongodb)

&nbsp;

## Why This Package?

This package is designed for modern cloud-native applications that require robust, high-performance MongoDB operations. It leverages the power of MongoDB while providing a developer-friendly API that integrates seamlessly with environment-based configurations.

🔝 [back to top](#go-mongodb)

&nbsp;

### Environment-First Design

Perfect for modern cloud deployments with Docker, Kubernetes, and CI/CD pipelines. No more hardcoded connection strings.

🔝 [back to top](#go-mongodb)

&nbsp;

### ULID IDs

Get 6x faster document ID generation with better database performance compared to UUIDs. Natural time-ordering and collision resistance.

🔝 [back to top](#go-mongodb)

&nbsp;

### Production-Ready

Built-in support for high availability, graceful shutdown, automatic reconnection, and comprehensive timeout controls.

🔝 [back to top](#go-mongodb)

&nbsp;

### Performance Optimized

Pluggable logging framework, efficient ULID generation, and optimized for high-throughput scenarios.

🔝 [back to top](#go-mongodb)

&nbsp;

## Production Usage

```go
package main

import (
    "context"
    "log"
    "time"
    "github.com/cloudresty/go-mongodb"
)

func main() {
    // Use custom environment prefix for multi-service deployments
    client, err := mongodb.NewClient(mongodb.FromEnvWithPrefix("PAYMENTS_"))
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Health checks and monitoring - use Ping to verify connectivity
    if err := client.Ping(context.Background()); err != nil {
        log.Fatalf("Health check failed: MongoDB is not reachable: %v", err)
    }
    log.Println("Health check passed: MongoDB is reachable.")

    // For detailed metrics, get the stats
    stats := client.Stats()
    log.Printf("Client stats: %+v", stats)

    // Graceful shutdown with signal handling
    shutdownManager := mongodb.NewShutdownManager(&mongodb.ShutdownConfig{
        Timeout: 30 * time.Second,
    })
    shutdownManager.SetupSignalHandler()
    shutdownManager.Register(client)
    shutdownManager.Wait() // Blocks until SIGINT/SIGTERM
}
```

🔝 [back to top](#go-mongodb)

&nbsp;

## Requirements

- Go 1.24+ (recommended)
- MongoDB 8.0+ (recommended)

🔝 [back to top](#go-mongodb)

&nbsp;

## Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

1. Fork the repository
2. Create a feature branch
3. Add tests for your changes
4. Ensure all tests pass
5. Submit a pull request

🔝 [back to top](#go-mongodb)

&nbsp;

## Security

If you discover a security vulnerability, please report it via email to [security@cloudresty.com](mailto:security@cloudresty.com).

🔝 [back to top](#go-mongodb)

&nbsp;

## License

This project is licensed under the MIT License - see the [LICENSE.txt](LICENSE.txt) file for details.

🔝 [back to top](#go-mongodb)

&nbsp;

---

&nbsp;

An open source project brought to you by the [Cloudresty](https://cloudresty.com) team.

[Website](https://cloudresty.com) &nbsp;|&nbsp; [LinkedIn](https://www.linkedin.com/company/cloudresty) &nbsp;|&nbsp; [BlueSky](https://bsky.app/profile/cloudresty.com) &nbsp;|&nbsp; [GitHub](https://github.com/cloudresty) &nbsp;|&nbsp; [Docker Hub](https://hub.docker.com/u/cloudresty)

&nbsp;
