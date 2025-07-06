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

# Testing variables (can be overridden)
# Default credentials match the Docker MongoDB setup (admin/password)
TEST_MONGODB_HOSTS ?= localhost:27017
TEST_MONGODB_DATABASE ?= app
TEST_MONGODB_USERNAME ?= admin
TEST_MONGODB_PASSWORD ?= password

# Testing
test: ## Run unit tests (with authentication support)
	@echo "Running unit tests with authentication..."
	@env -i PATH="$(PATH)" HOME="$(HOME)" \
	MONGODB_HOSTS=$(TEST_MONGODB_HOSTS) \
	MONGODB_USERNAME=$(TEST_MONGODB_USERNAME) \
	MONGODB_PASSWORD=$(TEST_MONGODB_PASSWORD) \
	MONGODB_DATABASE=$(TEST_MONGODB_DATABASE) \
	go test -v -race -short $(shell go list ./... | grep -v '/examples/')

test-integration: ## Run integration tests (always uses authentication)
	@echo "Running integration tests with authentication..."
	@mkdir -p test-results
	@if [ -n "$(TEST_MONGODB_USERNAME)" ] && [ -n "$(TEST_MONGODB_PASSWORD)" ]; then \
		echo "Using custom credentials: $(TEST_MONGODB_USERNAME)@$(TEST_MONGODB_HOSTS)/$(TEST_MONGODB_DATABASE)"; \
	else \
		echo "Using default credentials: admin@$(TEST_MONGODB_HOSTS)/$(TEST_MONGODB_DATABASE)"; \
	fi
	@echo "Output will be saved to test-results/integration.txt"
	@if env -i PATH="$(PATH)" HOME="$(HOME)" \
	MONGODB_HOSTS=$(TEST_MONGODB_HOSTS) \
	MONGODB_USERNAME=$(TEST_MONGODB_USERNAME) \
	MONGODB_PASSWORD=$(TEST_MONGODB_PASSWORD) \
	MONGODB_DATABASE=$(TEST_MONGODB_DATABASE) \
	go test -v -race $$(go list ./... | grep -v '/examples/') > test-results/integration.txt 2>&1; then \
		echo "Test completed successfully. Check test-results/integration.txt for results."; \
	else \
		echo "Test failed. Error output:"; \
		cat test-results/integration.txt; \
		exit 1; \
	fi

test-integration-auth: test-integration ## Alias for test-integration (for backward compatibility)

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage and authentication..."
	@env -i PATH="$(PATH)" HOME="$(HOME)" \
	MONGODB_HOSTS=$(TEST_MONGODB_HOSTS) \
	MONGODB_USERNAME=$(TEST_MONGODB_USERNAME) \
	MONGODB_PASSWORD=$(TEST_MONGODB_PASSWORD) \
	MONGODB_DATABASE=$(TEST_MONGODB_DATABASE) \
	go test -v -race -coverprofile=coverage.out $$(go list ./... | grep -v '/examples/')
	go tool cover -html=coverage.out -o coverage.html

# Build
build: ## Build examples
	@mkdir -p bin/
	go build -o bin/basic-client examples/basic-client/main.go
	go build -o bin/env-config examples/env-config/main.go
	go build -o bin/transactions examples/transactions/main.go
	go build -o bin/ulid-demo examples/ulid-demo/main.go
	go build -o bin/reconnection-test examples/reconnection-test/main.go
	go build -o bin/direct-connection examples/direct-connection/main.go

clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf test-results/
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
	@echo "Creating app database..."
	@docker exec mongodb-dev mongosh --username admin --password password --authenticationDatabase admin --eval \
		"db.getSiblingDB('app').init.insertOne({created: new Date()}); print('✓ App database created');"
	@echo "Database setup complete"

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

run-direct-connection: ## Run direct connection example
	go run examples/direct-connection/main.go

# CI
ci: tidy fmt vet test-integration ## Run CI pipeline with authentication

# Install tools
install-tools: ## Install development tools
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Security
security: ## Run security checks
	gosec ./...

# All tests
test-all: test test-integration ## Run all tests

test-results: ## Show latest test results
	@echo "=== Latest Test Results ==="
	@if [ -f test-results/integration.txt ]; then \
		echo "--- Integration Tests (test-results/integration.txt) ---"; \
		tail -n 20 test-results/integration.txt; \
		echo ""; \
	fi
	@echo "Use 'cat test-results/*.txt' to see full results"

test-status: ## Check if tests passed or failed
	@echo "=== Test Status Summary ==="
	@if [ -f test-results/integration.txt ]; then \
		if grep -q "PASS" test-results/integration.txt && ! grep -q "FAIL" test-results/integration.txt; then \
			echo "✅ Integration tests: PASSED"; \
		else \
			echo "❌ Integration tests: FAILED"; \
		fi; \
	else \
		echo "ℹ️  No integration test results found. Run 'make test-integration' first."; \
	fi

test-local: docker-stop docker-mongodb-full test-integration ## Setup local MongoDB with auth and run integration tests

test-local-simple: ## Run tests against existing MongoDB (detects auth automatically)
	@$(MAKE) test-integration

# Documentation
docs: ## Generate documentation
	@echo "Generating documentation..."
	@go doc -all . > docs/api-reference.txt

# Release
tag: ## Create a new git tag (requires VERSION env var)
	@if [ -z "$(VERSION)" ]; then echo "VERSION is required. Usage: make tag VERSION=v1.0.0"; exit 1; fi
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)

test-connection: ## Test MongoDB connection with credentials
	@echo "Testing MongoDB connection..."
	@docker exec mongodb-dev mongosh --username admin --password password --authenticationDatabase admin --eval \
		"try { \
			db.runCommand({ping: 1}); \
			print('✓ Authentication successful'); \
			db.getSiblingDB('app').test.insertOne({test: 'connection'}); \
			print('✓ Write operation successful'); \
			db.getSiblingDB('app').test.deleteMany({test: 'connection'}); \
			print('✓ Delete operation successful'); \
		} catch(e) { \
			print('✗ Connection failed: ' + e); \
		}"

# CI targets for GitHub Actions
ci-test: ## Run all tests for CI (unit + integration with auth)
	@echo "=== Running CI Test Suite ==="
	@echo "1/3 Running unit tests..."
	@$(MAKE) test
	@echo ""
	@echo "2/3 Running integration tests with authentication..."
	@$(MAKE) test-integration
	@echo ""
	@echo "3/3 Checking test results..."
	@$(MAKE) test-status
	@echo ""
	@echo "=== CI Test Suite Complete ==="

ci-setup-mongodb: ## Setup MongoDB for CI (Docker-based)
	@echo "Setting up MongoDB for CI..."
	@$(MAKE) docker-mongodb-full
	@echo "MongoDB setup complete"

ci-cleanup: ## Cleanup CI environment
	@echo "Cleaning up CI environment..."
	@$(MAKE) docker-stop clean
	@echo "Cleanup complete"

# Complete local testing workflow
test-full-local: ## Complete local test workflow (setup + test + cleanup)
	@echo "=== Full Local Test Workflow ==="
	@$(MAKE) ci-setup-mongodb
	@$(MAKE) ci-test
	@echo "=== Test Workflow Complete ==="
	@echo "Tip: Run 'make ci-cleanup' to stop MongoDB and clean up"

test-clean: ## Clean test results and cache
	@echo "Cleaning test results and cache..."
	@rm -rf test-results/
	@go clean -testcache
	@echo "Test cleanup complete"
