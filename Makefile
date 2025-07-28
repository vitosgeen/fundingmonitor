.PHONY: build run test clean docker-build docker-run help

# Default target
all: build

# Build the application
build:
	@echo "Building funding monitor..."
	go build -o fundingmonitor .

# Run the application
run: build
	@echo "Starting funding monitor..."
	@echo "Web interface: http://localhost:8080"
	@echo "API endpoints:"
	@echo "  - GET /api/funding"
	@echo "  - GET /api/funding/{exchange}"
	@echo "  - GET /api/health"
	@echo ""
	./fundingmonitor

# Run tests
test:
	@echo "Running tests..."
	go test -v

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f fundingmonitor

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
	@echo "  build           - Build the application"
	@echo "  run             - Build and run the application"
	@echo "  test            - Run tests"
	@echo "  clean           - Clean build artifacts"
	@echo "  docker-build    - Build Docker image"
	@echo "  docker-run      - Run with Docker"
	@echo "  docker-compose  - Run with docker-compose"
	@echo "  deps            - Install dependencies"
	@echo "  fmt             - Format code"
	@echo "  lint            - Lint code"
	@echo "  help            - Show this help" 