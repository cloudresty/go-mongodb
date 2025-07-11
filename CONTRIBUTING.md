# Contributing to Go MongoDB

[Home](README.md) &nbsp;/&nbsp; Contributing

&nbsp;

Thank you for your interest in contributing to Go MongoDB! We welcome
contributions from the community and are grateful for your help in making this package
better.

&nbsp;

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Code Style](#code-style)
- [Documentation](#documentation)
- [Community](#community)

🔝 [back to top](#contributing-to-go-mongodb)

&nbsp;

## Code of Conduct

This project and everyone participating in it is governed by our Code of
Conduct. By participating, you are expected to uphold this code. Please report
unacceptable behavior to [conduct@cloudresty.com](mailto:conduct@cloudresty.com).

🔝 [back to top](#contributing-to-go-mongodb)

&nbsp;

## Getting Started

&nbsp;

### Prerequisites

- Go 1.21+ (1.24+ recommended)
- MongoDB server (for integration tests)
- Git
- Make (optional, for convenience commands)

🔝 [back to top](#contributing-to-go-mongodb)

&nbsp;

### Ways to Contribute

- **Bug Reports**: Help us identify and fix issues
- **Feature Requests**: Suggest new functionality
- **Code Contributions**: Implement features, fix bugs, improve performance
- **Documentation**: Improve guides, examples, and API documentation
- **Testing**: Add test coverage, improve test reliability
- **Examples**: Create usage examples for common scenarios

🔝 [back to top](#contributing-to-go-mongodb)

&nbsp;

## Development Setup

&nbsp;

### 1. Fork and Clone

```bash
# Fork the repository on GitHub, then clone your fork
git clone https://github.com/YOUR_USERNAME/go-mongodb.git
cd go-mongodb

# Add the original repository as upstream
git remote add upstream https://github.com/cloudresty/go-mongodb.git
```

🔝 [back to top](#contributing-to-go-mongodb)

&nbsp;

### 2. Install Dependencies

```bash
# Download Go modules
go mod download

# Verify everything builds
go build ./...
```

🔝 [back to top](#contributing-to-go-mongodb)

&nbsp;

### 3. Start MongoDB (for testing)

```bash
# Using Docker (recommended)
make docker-mongodb

# Or install locally
# macOS: brew install mongodb/brew/mongodb-community
# Ubuntu: sudo apt-get install mongodb
# Or use Docker directly:
docker run -d --name mongodb -p 27017:27017 mongo:7.0
```

🔝 [back to top](#contributing-to-go-mongodb)

&nbsp;

### 4. Run Tests

```bash
# Unit tests only (no MongoDB required)
make test

# All tests including integration tests (requires MongoDB)
make test-integration

# Run tests with coverage
make test-coverage
```

🔝 [back to top](#contributing-to-go-mongodb)

&nbsp;

## Making Changes

&nbsp;

### 1. Create a Branch

```bash
# Create a descriptive branch name
git checkout -b feature/add-batch-operations
git checkout -b fix/connection-leak
git checkout -b docs/improve-examples
```

🔝 [back to top](#contributing-to-go-mongodb)

&nbsp;

### 2. Branch Naming Conventions

- Features: `feature/description`
- Bug fixes: `fix/description`
- Documentation: `docs/description`
- Tests: `test/description`
- Chores: `chore/description`

🔝 [back to top](#contributing-to-go-mongodb)

&nbsp;

### 3. Commit Guidelines

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```bash
# Format: type(scope): description
git commit -m "feat(collection): add batch insert operations"
git commit -m "fix(client): resolve connection leak in error handling"
git commit -m "docs(examples): add graceful shutdown example"
git commit -m "test(ulid): add performance benchmark tests"
```

**Types:**

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Adding or modifying tests
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `chore`: Build process or auxiliary tools

🔝 [back to top](#contributing-to-go-mongodb)

&nbsp;

## Testing

&nbsp;

### Test Types

1. **Unit Tests**: Test individual functions and methods
2. **Integration Tests**: Test with real MongoDB server
3. **Performance Tests**: Benchmark critical paths
4. **Example Tests**: Ensure examples compile and run

🔝 [back to top](#contributing-to-go-mongodb)

&nbsp;

### Writing Tests

```go
func TestNewClient(t *testing.T) {
    // Use testify for assertions
    client, err := mongodb.NewClient()
    require.NoError(t, err)
    assert.NotNil(t, client)

    // Clean up
    defer client.Close()
}

// Integration tests should check for MongoDB availability
func TestInsertIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    // Test with real MongoDB
}
```

🔝 [back to top](#contributing-to-go-mongodb)

&nbsp;

### Test Environment Variables

```bash
# For integration tests
export MONGODB_HOSTS=localhost:27017
export MONGODB_DATABASE=test

# For testing different configurations
export TEST_MONGODB_USERNAME="test_user"
export TEST_MONGODB_PASSWORD="test_pass"
```

🔝 [back to top](#contributing-to-go-mongodb)

&nbsp;

### Running Specific Tests

```bash
# Run specific test
go test -v -run TestClient

# Run with race detection
go test -race ./...

# Run all tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

🔝 [back to top](#contributing-to-go-mongodb)

&nbsp;

## Submitting Changes

&nbsp;

### 1. Before Submitting

✅ All tests pass (`make test-integration`)
✅ Code follows style guidelines (`make lint`)
✅ Documentation is updated
✅ Examples work and compile
✅ CHANGELOG.md is updated (if applicable)
✅ Commit messages follow conventions

🔝 [back to top](#contributing-to-go-mongodb)

&nbsp;

### 2. Create Pull Request

1. Push your branch to your fork
2. Create a pull request against `main` branch
3. Fill out the pull request template
4. Request review from maintainers

🔝 [back to top](#contributing-to-go-mongodb)

&nbsp;

### 3. Pull Request Template

```markdown
## Description
Brief description of the changes.

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] New tests added for new functionality

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] Examples updated (if applicable)
```

🔝 [back to top](#contributing-to-go-mongodb)

&nbsp;

### 4. Review Process

1. **Automated Checks**: CI/CD pipeline runs tests and linting
2. **Code Review**: Maintainers review for quality, style, and correctness
3. **Testing**: Verify functionality and performance
4. **Approval**: At least one maintainer approval required
5. **Merge**: Squash and merge to main branch

🔝 [back to top](#contributing-to-go-mongodb)

&nbsp;

## Code Style

&nbsp;

### Go Style Guide

We follow the standard Go style guidelines:

- Use `gofmt` for formatting
- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use [golangci-lint](https://golangci-lint.run/) for linting
- Write clear, self-documenting code

🔝 [back to top](#contributing-to-go-mongodb)

&nbsp;

### Code Formatting

```bash
# Format all Go files
go fmt ./...

# Run linter
make lint

# Fix common issues
make lint-fix
```

🔝 [back to top](#contributing-to-go-mongodb)

&nbsp;

### Variable and Function Naming

```go
// Good: Clear, descriptive names
func (c *Client) InsertWithRetry(ctx context.Context, config InsertConfig) error
var connectionTimeout time.Duration

// Avoid: Unclear abbreviations
func (c *Client) InsWithRetry(ctx context.Context, cfg InsCfg) error
var connTO time.Duration
```

🔝 [back to top](#contributing-to-go-mongodb)

&nbsp;

### Error Handling

```go
// Good: Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to insert document into collection %s: %w", collection, err)
}

// Good: Use structured logging
emit.Error.StructuredFields("Failed to connect to MongoDB",
    emit.ZString("host", sanitizeHost(host)),
    emit.ZString("error", err.Error()))
```

🔝 [back to top](#contributing-to-go-mongodb)

&nbsp;

### Comments and Documentation

```go
// InsertConfig holds configuration for inserting a single document.
// All fields are optional except Document.
type InsertConfig struct {
    // Collection is the MongoDB collection to insert into.
    // If empty, uses the default collection.
    Collection string

    // Document is the document to insert.
    // Must not be nil.
    Document interface{}

    // Options provides additional insert options.
    Options *options.InsertOneOptions
}
```

🔝 [back to top](#contributing-to-go-mongodb)

&nbsp;

## Documentation

&nbsp;

### Types of Documentation

1. **API Documentation**: Go doc comments for all public APIs
2. **User Guides**: Comprehensive usage guides in `docs/`
3. **Examples**: Working code examples in `examples/`
4. **README**: Quick start and overview

🔝 [back to top](#contributing-to-go-mongodb)

&nbsp;

### Documentation Guidelines

- Write clear, concise documentation
- Include code examples for complex functionality
- Update documentation when changing APIs
- Ensure examples compile and run

🔝 [back to top](#contributing-to-go-mongodb)

&nbsp;

### Adding Examples

```bash
# Create new example
mkdir examples/my-feature
cd examples/my-feature

# Create main.go with working example
cat > main.go << 'EOF'
package main

import (
    "context"
    "github.com/cloudresty/go-mongodb"
)

func main() {
    // Your example code here
}
EOF

# Test the example
go run main.go
```

🔝 [back to top](#contributing-to-go-mongodb)

&nbsp;

## Community

&nbsp;

### Getting Help

- **GitHub Issues**: For bugs and feature requests
- **GitHub Discussions**: For questions and community discussion
- **Documentation**: Check the `docs/` directory
- **Examples**: See `examples/` for usage patterns

🔝 [back to top](#contributing-to-go-mongodb)

&nbsp;

### Reporting Issues

When reporting issues, please include:

1. **Go version**: `go version`
2. **Package version**: Git commit or tag
3. **MongoDB version**: `mongod --version`
4. **Operating system**: OS and version
5. **Minimal reproduction case**: Code that demonstrates the issue
6. **Expected vs actual behavior**: What should happen vs what does happen
7. **Logs**: Relevant log output (sanitize sensitive information)

🔝 [back to top](#contributing-to-go-mongodb)

&nbsp;

### Issue Template

```markdown
## Bug Report

**Go Version:**
**Package Version:**
**MongoDB Version:**
**OS:**

**Description:**
A clear description of the bug.

**Reproduction Steps:**
1. Step one
2. Step two
3. Bug occurs

**Expected Behavior:**
What should happen.

**Actual Behavior:**
What actually happens.

**Code Example:**

```go
// Minimal code that reproduces the issue
```

**Logs:**

```text
Relevant log output (sanitize sensitive data)
```

🔝 [back to top](#contributing-to-go-mongodb)

&nbsp;

### Feature Requests

For feature requests, please provide:

1. **Use case**: What problem does this solve?
2. **Proposed solution**: How should it work?
3. **Alternatives considered**: Other approaches you've considered
4. **Breaking changes**: Would this break existing code?

🔝 [back to top](#contributing-to-go-mongodb)

&nbsp;

## Release Process

(For maintainers)

1. Update CHANGELOG.md
2. Update version in appropriate files
3. Create and push git tag
4. GitHub Actions will create release
5. Announce on relevant channels

🔝 [back to top](#contributing-to-go-mongodb)

&nbsp;

## Recognition

Contributors are recognized in:

- CHANGELOG.md for significant contributions
- GitHub releases for version contributions
- Special thanks in README for major features

🔝 [back to top](#contributing-to-go-mongodb)

&nbsp;

---

&nbsp;

Thank you for contributing to go-mongodb! Your efforts help make this package
better for everyone.

Questions? Feel free to open a GitHub Discussion or contact us at [contribute@cloudresty.com](mailto:contribute@cloudresty.com).

&nbsp;

---

&nbsp;

An open source project brought to you by the [Cloudresty](https://cloudresty.com/) team.

[Website](https://cloudresty.com/) | [LinkedIn](https://www.linkedin.com/company/cloudresty) | [BlueSky](https://bsky.app/profile/cloudresty.com) | [GitHub](https://github.com/cloudresty)

&nbsp;
