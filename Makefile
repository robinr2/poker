.PHONY: dev-backend dev-frontend build-frontend build-backend build test clean install-tools help

# Default target
help:
	@echo "Poker Application - Available Targets"
	@echo "======================================"
	@echo "  dev-backend      - Run backend with 'go run cmd/server/main.go'"
	@echo "  dev-frontend     - Run frontend dev server 'cd frontend && npm run dev'"
	@echo "  build-frontend   - Build frontend 'cd frontend && npm run build'"
	@echo "  build-backend    - Build Go binary to bin/poker"
	@echo "  build            - Build both frontend and backend"
	@echo "  test             - Run all tests (Go + Frontend)"
	@echo "  clean            - Remove bin/ and web/static/"
	@echo "  install-tools    - Install development tools (if any)"

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
