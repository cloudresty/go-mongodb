.PHONY: help test test-integration build clean lint fmt vet tidy examples docker-mongodb

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development
fmt: ## Format Go code
	go fmt ./...

vet: ## Run go vet
	go vet ./...

lint: ## Run golangci-lint
	golangci-lint run

tidy: ## Tidy go modules
	go mod tidy
	go mod verify

# Testing
test: ## Run unit tests
	go test -v -race -short ./...

test-integration: ## Run integration tests (requires MongoDB with test user)
	env -i PATH=$(PATH) HOME=$(HOME) \
	MONGODB_HOST=localhost \
	MONGODB_PORT=27017 \
	MONGODB_USERNAME=testuser \
	MONGODB_PASSWORD=testpass \
	MONGODB_DATABASE=testdb \
	go test -v -race ./...

test-coverage: ## Run tests with coverage
	env -i PATH=$(PATH) HOME=$(HOME) \
	MONGODB_HOST=localhost \
	MONGODB_PORT=27017 \
	MONGODB_USERNAME=testuser \
	MONGODB_PASSWORD=testpass \
	MONGODB_DATABASE=testdb \
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Build
build: ## Build examples
	@mkdir -p bin/
	go build -o bin/basic-client examples/basic-client/main.go
	go build -o bin/env-config examples/env-config/main.go
	go build -o bin/transactions examples/transactions/main.go
	go build -o bin/ulid-demo examples/ulid-demo/main.go
	go build -o bin/reconnection-test examples/reconnection-test/main.go

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html

# Docker
docker-mongodb: ## Start MongoDB in Docker
	docker run -d --name mongodb-dev \
		-p 27017:27017 \
		-e MONGO_INITDB_ROOT_USERNAME=admin \
		-e MONGO_INITDB_ROOT_PASSWORD=password \
		mongo:8

docker-setup-test-user: ## Create test database user (run after docker-mongodb)
	@echo "Waiting for MongoDB to be ready..."
	@for i in $$(seq 1 60); do \
		if nc -z localhost 27017; then \
			echo "MongoDB is ready!"; \
			break; \
		fi; \
		echo "Waiting... ($$i/60)"; \
		sleep 1; \
	done
	@sleep 5
	@echo "Creating test database user..."
	docker exec mongodb-dev mongosh --username admin --password password --authenticationDatabase admin --eval " \
		use admin; \
		db.createUser({ \
			user: 'testuser', \
			pwd: 'testpass', \
			roles: [ \
				{ role: 'readWrite', db: 'testdb' }, \
				{ role: 'dbAdmin', db: 'testdb' }, \
				{ role: 'readWrite', db: 'app' }, \
				{ role: 'dbAdmin', db: 'app' } \
			] \
		}); \
		db.runCommand({ connectionStatus: 1 }); \
	"

docker-mongodb-full: docker-mongodb docker-setup-test-user ## Start MongoDB and create test user

docker-stop: ## Stop MongoDB Docker container
	docker stop mongodb-dev || true
	docker rm mongodb-dev || true

# Examples
run-basic-client: ## Run basic client example
	go run examples/basic-client/main.go

run-env-config: ## Run environment config example
	go run examples/env-config/main.go

run-transactions: ## Run transactions example
	go run examples/transactions/main.go

run-ulid-demo: ## Run ULID demonstration example
	go run examples/ulid-demo/main.go

run-reconnection-test: ## Run auto-reconnection test example
	go run examples/reconnection-test/main.go

# CI
ci: tidy fmt vet test ## Run CI pipeline

# Install tools
install-tools: ## Install development tools
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Security
security: ## Run security checks
	gosec ./...

# Benchmarks
bench: ## Run benchmarks
	go test -bench=. -benchmem ./...

# All tests
test-all: test test-integration bench ## Run all tests including benchmarks

test-local: docker-stop docker-mongodb-full test-integration ## Setup local MongoDB and run integration tests

# Documentation
docs: ## Generate documentation
	@echo "Generating documentation..."
	@go doc -all . > docs/api-reference.txt

# Release
tag: ## Create a new git tag (requires VERSION env var)
	@if [ -z "$(VERSION)" ]; then echo "VERSION is required. Usage: make tag VERSION=v1.0.0"; exit 1; fi
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)
