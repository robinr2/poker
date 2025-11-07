.PHONY: dev-backend dev-frontend build-frontend build-backend build test clean install-tools help
.PHONY: docker-dev docker-down docker-build docker-test docker-clean docker-logs

# Default target
help:
	@echo "Poker Application - Available Targets"
	@echo "======================================"
	@echo ""
	@echo "Local Development:"
	@echo "  dev-backend      - Run backend with 'go run cmd/server/main.go'"
	@echo "  dev-frontend     - Run frontend dev server 'cd frontend && npm run dev'"
	@echo "  build-frontend   - Build frontend 'cd frontend && npm run build'"
	@echo "  build-backend    - Build Go binary to bin/poker"
	@echo "  build            - Build both frontend and backend"
	@echo "  test             - Run all tests (Go + Frontend)"
	@echo "  clean            - Remove bin/ and web/static/"
	@echo "  install-tools    - Install development tools (if any)"
	@echo ""
	@echo "Docker Development:"
	@echo "  docker-dev       - Start Docker development environment with hot reload"
	@echo "  docker-down      - Stop Docker development environment"
	@echo "  docker-build     - Build production Docker image"
	@echo "  docker-test      - Run all tests in Docker containers"
	@echo "  docker-clean     - Clean Docker images, containers, and volumes"
	@echo "  docker-logs      - View Docker container logs"

# Run backend in development mode
dev-backend:
	go run cmd/server/main.go

# Run frontend dev server
dev-frontend:
	cd frontend && npm run dev

# Build frontend
build-frontend:
	cd frontend && npm run build

# Build backend binary
build-backend: bin
	go build -o bin/poker cmd/server/main.go

# Build both frontend and backend
build: build-frontend build-backend

# Run all tests
test:
	go test ./internal/... -v
	cd frontend && npm test

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf web/static/

# Install development tools (if needed)
install-tools:
	@echo "Installing development tools..."
	go mod download
	cd frontend && npm install

# Create bin directory if it doesn't exist
bin:
	mkdir -p bin

# ============================================
# Docker Development Targets
# ============================================

# Start Docker development environment with hot reload
docker-dev:
	@echo "Starting Docker development environment..."
	docker-compose up --build

# Stop Docker development environment
docker-down:
	@echo "Stopping Docker development environment..."
	docker-compose down

# Build production Docker image
docker-build:
	@echo "Building production Docker image..."
	docker build -t poker:latest .
	@echo "✓ Docker image built: poker:latest"

# Run tests in Docker containers
docker-test:
	@echo "Running Go tests in Docker..."
	docker run --rm -v $(PWD):/app -w /app golang:1.24-alpine sh -c "go test ./internal/... -v"
	@echo ""
	@echo "Running frontend tests in Docker..."
	docker run --rm -v $(PWD)/frontend:/app -w /app node:24-alpine sh -c "npm install && npm test"
	@echo ""
	@echo "Running Docker integration tests..."
	./scripts/test-docker.sh

# Clean Docker images, containers, and volumes
docker-clean:
	@echo "Cleaning Docker resources..."
	docker-compose down -v
	docker image rm poker:latest 2>/dev/null || true
	docker image rm poker:test 2>/dev/null || true
	@echo "✓ Docker resources cleaned"

# View Docker container logs
docker-logs:
	docker-compose logs -f
