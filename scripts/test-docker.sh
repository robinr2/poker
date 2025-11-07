#!/bin/bash

# Integration tests for Docker containerization
# Tests verify that Docker builds correctly and services start properly

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Test results
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Helper functions
test_start() {
    TESTS_RUN=$((TESTS_RUN + 1))
    echo -e "${BLUE}[TEST $TESTS_RUN] $1${NC}"
}

test_pass() {
    TESTS_PASSED=$((TESTS_PASSED + 1))
    echo -e "${GREEN}✓ PASSED${NC}"
}

test_fail() {
    TESTS_FAILED=$((TESTS_FAILED + 1))
    echo -e "${RED}✗ FAILED: $1${NC}"
}

# Test 1: Verify Dockerfile exists and is valid
test_dockerfile_exists() {
    test_start "Dockerfile exists and is valid"
    
    if [ ! -f "$PROJECT_ROOT/Dockerfile" ]; then
        test_fail "Dockerfile not found"
        return 1
    fi
    
    if ! grep -q "FROM golang:1.24-alpine" "$PROJECT_ROOT/Dockerfile"; then
        test_fail "Dockerfile doesn't have golang:1.24-alpine base image"
        return 1
    fi
    
    if ! grep -q "FROM alpine:latest" "$PROJECT_ROOT/Dockerfile"; then
        test_fail "Dockerfile doesn't have alpine:latest runtime image"
        return 1
    fi
    
    if ! grep -q "EXPOSE 8080" "$PROJECT_ROOT/Dockerfile"; then
        test_fail "Dockerfile doesn't expose port 8080"
        return 1
    fi
    
    test_pass
    return 0
}

# Test 2: Verify .dockerignore exists
test_dockerignore_exists() {
    test_start ".dockerignore exists with proper exclusions"
    
    if [ ! -f "$PROJECT_ROOT/.dockerignore" ]; then
        test_fail ".dockerignore not found"
        return 1
    fi
    
    if ! grep -q "node_modules/" "$PROJECT_ROOT/.dockerignore"; then
        test_fail ".dockerignore doesn't exclude node_modules/"
        return 1
    fi
    
    if ! grep -q "frontend/dist/" "$PROJECT_ROOT/.dockerignore"; then
        test_fail ".dockerignore doesn't exclude frontend/dist/"
        return 1
    fi
    
    test_pass
    return 0
}

# Test 3: Verify .air.toml exists
test_air_config_exists() {
    test_start ".air.toml exists with proper configuration"
    
    if [ ! -f "$PROJECT_ROOT/.air.toml" ]; then
        test_fail ".air.toml not found"
        return 1
    fi
    
    if ! grep -q "bin = \"./tmp/main\"" "$PROJECT_ROOT/.air.toml"; then
        test_fail ".air.toml doesn't have correct bin path"
        return 1
    fi
    
    if ! grep -q "cmd/server" "$PROJECT_ROOT/.air.toml"; then
        test_fail ".air.toml build command doesn't reference cmd/server"
        return 1
    fi
    
    test_pass
    return 0
}

# Test 4: Verify docker-compose.yml exists
test_docker_compose_exists() {
    test_start "docker-compose.yml exists with required services"
    
    if [ ! -f "$PROJECT_ROOT/docker-compose.yml" ]; then
        test_fail "docker-compose.yml not found"
        return 1
    fi
    
    if ! grep -q "backend:" "$PROJECT_ROOT/docker-compose.yml"; then
        test_fail "docker-compose.yml doesn't define backend service"
        return 1
    fi
    
    if ! grep -q "frontend:" "$PROJECT_ROOT/docker-compose.yml"; then
        test_fail "docker-compose.yml doesn't define frontend service"
        return 1
    fi
    
    if ! grep -q "5173:5173" "$PROJECT_ROOT/docker-compose.yml"; then
        test_fail "docker-compose.yml doesn't expose frontend port"
        return 1
    fi
    
    if ! grep -q "8080:8080" "$PROJECT_ROOT/docker-compose.yml"; then
        test_fail "docker-compose.yml doesn't expose backend port"
        return 1
    fi
    
    test_pass
    return 0
}

# Test 5: Verify docker-compose config is valid
test_docker_compose_config() {
    test_start "docker-compose configuration is valid"
    
    if ! docker-compose config > /dev/null 2>&1; then
        test_fail "docker-compose config validation failed"
        return 1
    fi
    
    test_pass
    return 0
}

# Test 6: Verify docker-dev.sh script exists and is executable
test_docker_dev_script() {
    test_start "docker-dev.sh script exists and is executable"
    
    if [ ! -f "$PROJECT_ROOT/scripts/docker-dev.sh" ]; then
        test_fail "scripts/docker-dev.sh not found"
        return 1
    fi
    
    if [ ! -x "$PROJECT_ROOT/scripts/docker-dev.sh" ]; then
        test_fail "scripts/docker-dev.sh is not executable"
        return 1
    fi
    
    test_pass
    return 0
}

# Test 7: Verify Docker build completes successfully
test_docker_build() {
    test_start "Docker build completes successfully"
    
    if ! docker build -t poker:test . > /dev/null 2>&1; then
        test_fail "Docker build failed"
        return 1
    fi
    
    # Verify image was created
    if ! docker image ls poker:test --quiet | grep -q "."; then
        test_fail "Docker image not created"
        return 1
    fi
    
    test_pass
    
    # Clean up test image
    docker image rm poker:test > /dev/null 2>&1 || true
    return 0
}

# Test 8: Verify main.go binds to 0.0.0.0
test_server_binding() {
    test_start "Server binds to 0.0.0.0 for Docker accessibility"
    
    if ! grep -q "0.0.0.0" "$PROJECT_ROOT/cmd/server/main.go"; then
        test_fail "Server doesn't bind to 0.0.0.0"
        return 1
    fi
    
    test_pass
    return 0
}

# Test 9: Verify Dockerfile has multi-stage build
test_dockerfile_multistage() {
    test_start "Dockerfile uses multi-stage build pattern"
    
    builder_count=$(grep -c "FROM golang:1.24-alpine" "$PROJECT_ROOT/Dockerfile" || true)
    alpine_count=$(grep -c "FROM alpine:latest" "$PROJECT_ROOT/Dockerfile" || true)
    
    if [ "$builder_count" -ne 1 ] || [ "$alpine_count" -ne 1 ]; then
        test_fail "Dockerfile doesn't have proper multi-stage build (builder + runtime)"
        return 1
    fi
    
    if ! grep -q "COPY --from=builder" "$PROJECT_ROOT/Dockerfile"; then
        test_fail "Dockerfile doesn't copy from builder stage"
        return 1
    fi
    
    test_pass
    return 0
}

# Test 10: Verify Air hot reload is configured
test_air_hot_reload() {
    test_start "Air hot reload configured for Go files"
    
    if ! grep -q "include_ext.*go" "$PROJECT_ROOT/.air.toml"; then
        test_fail ".air.toml doesn't include .go extension"
        return 1
    fi
    
    if ! grep -q "exclude_dir.*frontend" "$PROJECT_ROOT/.air.toml"; then
        test_fail ".air.toml doesn't exclude frontend directory"
        return 1
    fi
    
    test_pass
    return 0
}

# Main test execution
main() {
    cd "$PROJECT_ROOT"
    
    echo -e "${BLUE}================================================${NC}"
    echo -e "${BLUE}Docker Containerization Integration Tests${NC}"
    echo -e "${BLUE}================================================${NC}"
    echo ""
    
    # Run all tests
    test_dockerfile_exists || true
    test_dockerignore_exists || true
    test_air_config_exists || true
    test_docker_compose_exists || true
    test_docker_compose_config || true
    test_docker_dev_script || true
    test_docker_build || true
    test_server_binding || true
    test_dockerfile_multistage || true
    test_air_hot_reload || true
    
    echo ""
    echo -e "${BLUE}================================================${NC}"
    echo -e "${BLUE}Test Results${NC}"
    echo -e "${BLUE}================================================${NC}"
    echo -e "Tests run:    ${TESTS_RUN}"
    echo -e "Tests passed: ${GREEN}${TESTS_PASSED}${NC}"
    echo -e "Tests failed: $([ $TESTS_FAILED -eq 0 ] && echo -e "${GREEN}${TESTS_FAILED}${NC}" || echo -e "${RED}${TESTS_FAILED}${NC}")"
    echo ""
    
    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "${GREEN}All tests passed! ✓${NC}"
        return 0
    else
        echo -e "${RED}Some tests failed! ✗${NC}"
        return 1
    fi
}

main "$@"
