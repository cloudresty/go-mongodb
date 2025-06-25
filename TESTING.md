# Testing Guide

This document explains how to run tests for the go-mongodb package.

## Overview

The package includes comprehensive testing with both unit tests and integration tests that work with real MongoDB instances.

## Test Categories

### Unit Tests
- **Purpose**: Test core functionality without requiring MongoDB
- **Speed**: Fast execution
- **Command**: `make test`
- **Coverage**: Configuration, validation, ID generation logic, environment loading

### Integration Tests

- **Purpose**: Test actual MongoDB operations
- **Speed**: Slower execution (requires MongoDB connection)
- **Command**: `make test-integration`
- **Coverage**: CRUD operations, transactions, indexes, ULID generation, connection handling
- **Authentication**: Always requires MongoDB authentication (testuser/testpass by default)

## Prerequisites

### For Unit Tests
- Go 1.23+ installed
- No additional dependencies

### For Integration Tests
- Docker installed (for local MongoDB)
- OR existing MongoDB instance with authentication

## Running Tests Locally

### Quick Start (Automated Setup)
```bash
# Complete workflow: setup MongoDB + run all tests
make test-full-local

# When done, cleanup
make ci-cleanup
```

### Manual Setup
```bash
# 1. Start MongoDB with authentication
make docker-mongodb-full

# 2. Run unit tests
make test

# 3. Run integration tests
make test-integration

# 4. Check test status
make test-status

# 5. View detailed results
make test-results
```

### Individual Test Commands
```bash
# Unit tests only
make test

# Integration tests with auto-detection
make test-integration

# Test specific functionality
go test -v -run TestClientCreation
go test -v -run TestULID
```

## MongoDB Configuration

### Default Test Configuration
- **Host**: localhost
- **Port**: 27017
- **Username**: testuser
- **Password**: testpass
- **Database**: testdb

### Custom Configuration
You can override defaults using environment variables:
```bash
TEST_MONGODB_HOST=custom-host \
TEST_MONGODB_PORT=27017 \
TEST_MONGODB_USERNAME=myuser \
TEST_MONGODB_PASSWORD=mypass \
TEST_MONGODB_DATABASE=mydb \
make test-integration
```

### Authentication Support

MongoDB always requires authentication in this setup:

- **Default credentials**: testuser/testpass for testdb database
- **Custom credentials**: Set `TEST_MONGODB_USERNAME`, `TEST_MONGODB_PASSWORD`, `TEST_MONGODB_DATABASE` environment variables
- **No unauthenticated mode**: All tests require valid MongoDB credentials

## Available Makefile Targets

### Core Testing
- `make test` - Run unit tests (fast)
- `make test-integration` - Run integration tests with auto-detection
- `make test-integration-auth` - Run integration tests with authentication
- `make test-all` - Run both unit and integration tests

### MongoDB Management
- `make docker-mongodb` - Start MongoDB container
- `make docker-setup-test-user` - Create test user
- `make docker-mongodb-full` - Start MongoDB + create test user
- `make docker-stop` - Stop and remove MongoDB container

### Test Utilities
- `make test-status` - Show pass/fail status summary
- `make test-results` - Show detailed test output
- `make test-connection` - Test MongoDB connection
- `make clean` - Clean test artifacts

### CI/Development
- `make ci-test` - Complete CI test suite
- `make ci-setup-mongodb` - Setup MongoDB for CI
- `make test-full-local` - Full local test workflow

## Test Output

Tests output results to files in `test-results/`:

- `integration.txt` - Integration tests (with authentication)

## Continuous Integration

The CI pipeline (`.github/workflows/ci.yaml`) uses the same MongoDB setup and test commands as local development, ensuring consistency between local and CI environments.

### CI Workflow
1. Start MongoDB service with authentication
2. Create test user
3. Run unit tests
4. Run integration tests
5. Generate coverage reports

## Troubleshooting

### Common Issues

**"Authentication failed"**
- Ensure MongoDB is running: `docker ps`
- Check test user exists: `make test-connection`
- Recreate user: `make docker-setup-test-user`

**"Connection refused"**
- Start MongoDB: `make docker-mongodb-full`
- Check port 27017 is available: `nc -z localhost 27017`

**"Tests are cached"**
- Clear test cache: `go clean -testcache`
- Force fresh test run: `make test-integration-auth`

**Test failures**
- Check detailed output: `make test-results`
- Run specific test: `go test -v -run TestSpecificName`
- Verify MongoDB data: `make test-connection`

### MongoDB Docker Issues
```bash
# Stop and restart MongoDB
make docker-stop
make docker-mongodb-full

# Check container logs
docker logs mongodb-dev

# Manually test connection
docker exec mongodb-dev mongosh --username admin --password password
```

## Test Coverage

Run tests with coverage reporting:
```bash
make test-coverage
open coverage.html  # View coverage report
```

## ID Mode Testing

The package includes comprehensive testing for all ID modes:
- **ULID mode** (default): Tests ULID generation, uniqueness, temporal ordering
- **ObjectID mode**: Tests MongoDB ObjectID compatibility
- **Custom mode**: Tests user-provided IDs

All modes are tested in both unit and integration scenarios.

## Performance Testing

Integration tests include performance validation:
- Bulk operations (1000+ documents)
- Concurrent ULID generation
- Transaction performance
- Connection pooling behavior

## Best Practices

1. **Always run unit tests first** - they're fast and catch basic issues
2. **Use Docker for consistent testing** - avoid local MongoDB configuration issues
3. **Check test status before pushing** - ensure all tests pass locally
4. **Clean up after testing** - stop Docker containers to free resources
5. **Test with realistic data** - integration tests use production-like scenarios

## Examples

The `examples/` directory contains working code samples that can be run for manual testing:
```bash
make run-basic-client
make run-ulid-demo
make run-transactions
```

---

For more details, see the main [README.md](README.md) and [API documentation](docs/api-reference.md).
