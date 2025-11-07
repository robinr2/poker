#!/bin/bash

# Test script to verify all linters run successfully on clean code
# This test validates that the linting infrastructure is properly set up

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== Testing Lint Infrastructure ===${NC}"
echo ""

# Test 1: Go fmt check
echo -e "${YELLOW}Test 1: Checking Go code formatting...${NC}"
cd "$PROJECT_ROOT"
if ! go fmt ./internal/... ./cmd/...; then
    echo -e "${RED}✗ Go fmt check failed${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Go fmt check passed${NC}"
echo ""

# Test 2: Go vet
echo -e "${YELLOW}Test 2: Running Go vet...${NC}"
if ! go vet ./internal/... ./cmd/...; then
    echo -e "${RED}✗ Go vet check failed${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Go vet check passed${NC}"
echo ""

# Test 3: ESLint
echo -e "${YELLOW}Test 3: Running ESLint...${NC}"
cd "$PROJECT_ROOT/frontend"
if ! npm run lint; then
    echo -e "${RED}✗ ESLint check failed${NC}"
    exit 1
fi
echo -e "${GREEN}✓ ESLint check passed${NC}"
echo ""

echo -e "${GREEN}=== All lint tests passed ===${NC}"
