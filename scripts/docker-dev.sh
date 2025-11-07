#!/bin/bash

# Docker development environment helper script
# Starts the poker application with hot reload support for both backend and frontend

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print styled output
print_header() {
    echo -e "${BLUE}=== $1 ===${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

# Check if Docker is installed
check_docker() {
    print_header "Checking Docker Installation"
    
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed"
        echo "Please install Docker from https://www.docker.com/products/docker-desktop"
        exit 1
    fi
    print_success "Docker is installed"
    
    if ! command -v docker-compose &> /dev/null; then
        print_error "docker-compose is not installed"
        echo "Please install docker-compose or upgrade Docker to the latest version"
        exit 1
    fi
    print_success "docker-compose is installed"
}

# Check if Docker daemon is running
check_docker_running() {
    print_header "Checking Docker Daemon"
    
    if ! docker info &> /dev/null; then
        print_error "Docker daemon is not running"
        echo "Please start Docker Desktop or the Docker daemon"
        exit 1
    fi
    print_success "Docker daemon is running"
}

# Display helpful information
show_info() {
    echo ""
    print_header "Application URLs"
    echo -e "Frontend: ${GREEN}http://localhost:5173${NC}"
    echo -e "Backend:  ${GREEN}http://localhost:8080${NC}"
    echo ""
    print_header "Logs"
    echo "View backend logs:  docker-compose logs -f backend"
    echo "View frontend logs: docker-compose logs -f frontend"
    echo "View all logs:      docker-compose logs -f"
    echo ""
    print_header "Hot Reload"
    echo "Backend (Go):   Changes to *.go files in cmd/ and internal/ will auto-rebuild"
    echo "Frontend (React): Changes to *.tsx/*.ts files will auto-reload in browser"
    echo ""
    print_header "Common Commands"
    echo "Stop services:   docker-compose down"
    echo "Rebuild images:  docker-compose build --no-cache"
    echo ""
}

# Main execution
main() {
    print_header "Poker Application - Docker Development Environment"
    
    # Check prerequisites
    check_docker
    check_docker_running
    
    print_header "Starting Services"
    print_info "Starting backend and frontend services with hot reload enabled..."
    echo ""
    
    # Start docker-compose with flags
    docker-compose up --build "$@"
}

# Trap EXIT to show goodbye message
trap 'echo ""; print_info "Docker development environment stopped"; exit 0' EXIT

# Show info after a brief delay (when containers are starting)
show_info &
INFO_PID=$!

# Run main function
main "$@"
