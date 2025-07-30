.PHONY: build run test clean docker-build docker-run help stop status build-clean run-clean

# Default target
all: build-clean

# Build the application (clean architecture)
build-clean:
	@echo "Building funding monitor (clean architecture)..."
	go build -o fundingmonitor_clean main_clean.go

# Build the application (legacy)
build:
	@echo "Building funding monitor (legacy)..."
	go build -o fundingmonitor .

# Run the application (clean architecture)
run-clean: build-clean
	@echo "Starting funding monitor (clean architecture)..."
	@echo "Web interface: http://localhost:8080"
	@echo "API endpoints:"
	@echo "  - GET /api/funding"
	@echo "  - GET /api/funding/{exchange}"
	@echo "  - GET /api/health"
	@echo "  - GET /api/logs"
	@echo "  - GET /api/logs/{symbol}"
	@echo ""
	@echo "Press Ctrl+C to stop"
	./fundingmonitor_clean

# Run the application (legacy)
run: build
	@echo "Starting funding monitor (legacy)..."
	@echo "Web interface: http://localhost:8080"
	@echo "API endpoints:"
	@echo "  - GET /api/funding"
	@echo "  - GET /api/funding/{exchange}"
	@echo "  - GET /api/health"
	@echo ""
	./fundingmonitor

# Stop the running application
stop:
	@echo "Stopping funding monitor..."
	@if pgrep -f "fundingmonitor" > /dev/null; then \
		pkill -f "fundingmonitor"; \
		echo "Application stopped."; \
	else \
		echo "No funding monitor process found."; \
	fi

# Check application status
status:
	@echo "Checking funding monitor status..."
	@if pgrep -f "fundingmonitor" > /dev/null; then \
		echo "✅ Funding monitor is running"; \
		ps aux | grep fundingmonitor | grep -v grep; \
	else \
		echo "❌ Funding monitor is not running"; \
	fi

# Run tests
test:
	@echo "Running tests..."
	./run_tests.sh all

# Run unit tests only
test-unit:
	@echo "Running unit tests..."
	./run_tests.sh unit

# Run integration tests only
test-integration:
	@echo "Running integration tests..."
	./run_tests.sh integration

# Run E2E tests only
test-e2e:
	@echo "Running E2E tests..."
	./run_tests.sh e2e

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f fundingmonitor fundingmonitor_clean

# Clean logs
clean-logs:
	@echo "Cleaning log files..."
	rm -rf funding_logs/*

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t fundingmonitor .

# Run with Docker
docker-run: docker-build
	@echo "Running with Docker..."
	docker run -p 8080:8080 fundingmonitor

# Run with docker-compose
docker-compose:
	@echo "Running with docker-compose..."
	docker-compose up --build

# Stop docker-compose
docker-compose-down:
	@echo "Stopping docker-compose..."
	docker-compose down

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	golangci-lint run

# Show help
help:
	@echo "Available targets:"
	@echo "  build-clean      - Build the application (clean architecture)"
	@echo "  build            - Build the application (legacy)"
	@echo "  run-clean        - Build and run the application (clean architecture)"
	@echo "  run              - Build and run the application (legacy)"
	@echo "  stop             - Stop the running application"
	@echo "  status           - Check application status"
	@echo "  test             - Run all tests"
	@echo "  test-unit        - Run unit tests only"
	@echo "  test-integration - Run integration tests only"
	@echo "  test-e2e         - Run E2E tests only"
	@echo "  clean            - Clean build artifacts"
	@echo "  clean-logs       - Clean log files"
	@echo "  docker-build     - Build Docker image"
	@echo "  docker-run       - Run with Docker"
	@echo "  docker-compose   - Run with docker-compose"
	@echo "  deps             - Install dependencies"
	@echo "  fmt              - Format code"
	@echo "  lint             - Lint code"
	@echo "  help             - Show this help" 