#!/bin/bash

# Integration test script - validates that all documentation and commands work
# Tests that the project is properly set up and documented

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters
FAILED=0
PASSED=0

echo -e "${BLUE}================================================${NC}"
echo -e "${BLUE}  Poker Application - Integration Test Suite${NC}"
echo -e "${BLUE}================================================${NC}"
echo ""

# Test 1: Check required files exist
echo -e "${YELLOW}>> Checking required files exist${NC}"
echo ""

check_file() {
    local file="$1"
    if [ -f "$PROJECT_ROOT/$file" ]; then
        echo -e "  ${GREEN}✓${NC} $file"
        ((PASSED++))
    else
        echo -e "  ${RED}✗${NC} $file (missing)"
        ((FAILED++))
    fi
}

check_file "README.md"
check_file ".env.example"
check_file "CONTRIBUTING.md"
check_file "Makefile"
check_file "docker-compose.yml"
check_file "Dockerfile"
check_file "go.mod"
check_file "cmd/server/main.go"
check_file "frontend/package.json"
check_file "docs/ARCHITECTURE.md"
check_file "docs/DEVELOPMENT.md"
check_file "docs/WEBSOCKET_PROTOCOL.md"

echo ""

# Test 2: Verify Makefile targets
echo -e "${YELLOW}>> Verifying Makefile targets${NC}"
echo ""

check_make_target() {
    local target="$1"
    if grep -q "^${target}:" "$PROJECT_ROOT/Makefile"; then
        echo -e "  ${GREEN}✓${NC} make $target target exists"
        ((PASSED++))
    else
        echo -e "  ${RED}✗${NC} make $target target missing"
        ((FAILED++))
    fi
}

check_make_target "help"
check_make_target "dev-backend"
check_make_target "dev-frontend"
check_make_target "build"
check_make_target "test"
check_make_target "lint"
check_make_target "docker-dev"

echo ""

# Test 3: Verify environment variables are documented
echo -e "${YELLOW}>> Verifying environment variables are documented${NC}"
echo ""

check_env_var() {
    local var="$1"
    if grep -q "$var" "$PROJECT_ROOT/.env.example"; then
        echo -e "  ${GREEN}✓${NC} $var in .env.example"
        ((PASSED++))
    else
        echo -e "  ${RED}✗${NC} $var not in .env.example"
        ((FAILED++))
    fi
}

check_env_var "PORT"
check_env_var "LOG_LEVEL"
check_env_var "NODE_ENV"

echo ""

# Test 4: Verify documentation sections
echo -e "${YELLOW}>> Verifying documentation completeness${NC}"
echo ""

check_readme_section() {
    local section="$1"
    if grep -qi "$section" "$PROJECT_ROOT/README.md"; then
        echo -e "  ${GREEN}✓${NC} README.md has '$section' section"
        ((PASSED++))
    else
        echo -e "  ${RED}✗${NC} README.md missing '$section' section"
        ((FAILED++))
    fi
}

check_readme_section "Quick Start"
check_readme_section "Features"
check_readme_section "Tech Stack"
check_readme_section "Prerequisites"
check_readme_section "Development"
check_readme_section "Testing"
check_readme_section "Docker"

echo ""

check_arch_section() {
    local section="$1"
    if grep -qi "$section" "$PROJECT_ROOT/docs/ARCHITECTURE.md"; then
        echo -e "  ${GREEN}✓${NC} ARCHITECTURE.md has '$section' section"
        ((PASSED++))
    else
        echo -e "  ${RED}✗${NC} ARCHITECTURE.md missing '$section' section"
        ((FAILED++))
    fi
}

check_arch_section "Backend"
check_arch_section "Frontend"
check_arch_section "WebSocket"
check_arch_section "Communication"

echo ""

check_dev_section() {
    local section="$1"
    if grep -qi "$section" "$PROJECT_ROOT/docs/DEVELOPMENT.md"; then
        echo -e "  ${GREEN}✓${NC} DEVELOPMENT.md has '$section' section"
        ((PASSED++))
    else
        echo -e "  ${RED}✗${NC} DEVELOPMENT.md missing '$section' section"
        ((FAILED++))
    fi
}

check_dev_section "Setup"
check_dev_section "Workflow"
check_dev_section "Testing"
check_dev_section "Code Style"

echo ""

check_ws_section() {
    local section="$1"
    if grep -qi "$section" "$PROJECT_ROOT/docs/WEBSOCKET_PROTOCOL.md"; then
        echo -e "  ${GREEN}✓${NC} WEBSOCKET_PROTOCOL.md has '$section' section"
        ((PASSED++))
    else
        echo -e "  ${RED}✗${NC} WEBSOCKET_PROTOCOL.md missing '$section' section"
        ((FAILED++))
    fi
}

check_ws_section "Connection"
check_ws_section "Message"
check_ws_section "Error"

echo ""

check_contrib_section() {
    local section="$1"
    if grep -qi "$section" "$PROJECT_ROOT/CONTRIBUTING.md"; then
        echo -e "  ${GREEN}✓${NC} CONTRIBUTING.md has '$section' section"
        ((PASSED++))
    else
        echo -e "  ${RED}✗${NC} CONTRIBUTING.md missing '$section' section"
        ((FAILED++))
    fi
}

check_contrib_section "Getting Started"
check_contrib_section "Development Process"
check_contrib_section "Code Style"
check_contrib_section "Pull Request"

echo ""

# Test 5: Verify scripts are executable
echo -e "${YELLOW}>> Verifying scripts are executable${NC}"
echo ""

check_script_executable() {
    local script="$1"
    if [ -x "$PROJECT_ROOT/$script" ]; then
        echo -e "  ${GREEN}✓${NC} $script is executable"
        ((PASSED++))
    else
        echo -e "  ${RED}✗${NC} $script is not executable"
        ((FAILED++))
    fi
}

check_script_executable "scripts/test.sh"
check_script_executable "scripts/lint.sh"
check_script_executable "scripts/build.sh"
check_script_executable "scripts/docker-dev.sh"

echo ""

# Test 6: Test that make help works
echo -e "${YELLOW}>> Testing make help command${NC}"
echo ""

cd "$PROJECT_ROOT"
if make help > /dev/null 2>&1; then
    echo -e "  ${GREEN}✓${NC} make help works"
    ((PASSED++))
else
    echo -e "  ${RED}✗${NC} make help failed"
    ((FAILED++))
fi

echo ""

# Test 7: Verify backend builds
echo -e "${YELLOW}>> Testing backend build${NC}"
echo ""

cd "$PROJECT_ROOT"
if make build-backend > /dev/null 2>&1; then
    echo -e "  ${GREEN}✓${NC} make build-backend succeeds"
    ((PASSED++))
else
    echo -e "  ${RED}✗${NC} make build-backend failed"
    ((FAILED++))
fi

echo ""

# Test 8: Verify frontend builds
echo -e "${YELLOW}>> Testing frontend build${NC}"
echo ""

cd "$PROJECT_ROOT"
if make build-frontend > /dev/null 2>&1; then
    echo -e "  ${GREEN}✓${NC} make build-frontend succeeds"
    ((PASSED++))
else
    echo -e "  ${RED}✗${NC} make build-frontend failed"
    ((FAILED++))
fi

echo ""

# Test 9: Verify Docker build
echo -e "${YELLOW}>> Testing Docker build${NC}"
echo ""

cd "$PROJECT_ROOT"
if make docker-build > /dev/null 2>&1; then
    echo -e "  ${GREEN}✓${NC} make docker-build succeeds"
    ((PASSED++))
else
    echo -e "  ${RED}✗${NC} make docker-build failed"
    ((FAILED++))
fi

echo ""

# Summary
echo -e "${BLUE}================================================${NC}"
echo -e "${BLUE}  Integration Test Summary${NC}"
echo -e "${BLUE}================================================${NC}"

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All integration tests passed!${NC}"
    echo -e "  Passed: ${PASSED} | Failed: 0"
    echo -e "${BLUE}================================================${NC}"
    exit 0
else
    echo -e "${RED}✗ Some integration tests failed${NC}"
    echo -e "  Passed: ${PASSED} | Failed: ${FAILED}"
    echo -e "${BLUE}================================================${NC}"
    exit 1
fi
