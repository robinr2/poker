#!/bin/bash

# Comprehensive linting script for the poker application
# Runs all linters for Go backend and React frontend

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

echo -e "${BLUE}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║          Poker Application - Lint Check                ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════╝${NC}"
echo ""

# Go linting
echo -e "${YELLOW}→ Linting Go Backend${NC}"
echo ""

# Go fmt
echo -e "  ${YELLOW}Running gofmt...${NC}"
cd "$PROJECT_ROOT"
go fmt ./internal/... ./cmd/... > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo -e "    ${GREEN}✓ gofmt passed${NC}"
    ((PASSED++))
else
    echo -e "    ${RED}✗ gofmt failed${NC}"
    ((FAILED++))
fi

# Go vet
echo -e "  ${YELLOW}Running go vet...${NC}"
cd "$PROJECT_ROOT"
go vet ./internal/... ./cmd/... > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo -e "    ${GREEN}✓ go vet passed${NC}"
    ((PASSED++))
else
    echo -e "    ${RED}✗ go vet failed${NC}"
    go vet ./internal/... ./cmd/...
    ((FAILED++))
fi

echo ""
echo -e "${YELLOW}→ Linting Frontend (React/TypeScript)${NC}"
echo ""

# ESLint
echo -e "  ${YELLOW}Running ESLint...${NC}"
cd "$PROJECT_ROOT/frontend"
npm run lint > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo -e "    ${GREEN}✓ ESLint passed${NC}"
    ((PASSED++))
else
    echo -e "    ${RED}✗ ESLint failed${NC}"
    npm run lint
    ((FAILED++))
fi

echo ""
echo -e "${BLUE}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║                   Lint Summary                         ║${NC}"
echo -e "${BLUE}╠════════════════════════════════════════════════════════╣${NC}"

if [ $FAILED -eq 0 ]; then
    echo -e "${BLUE}║  ${GREEN}✓ All linters passed${NC}${BLUE}                          ║${NC}"
    echo -e "${BLUE}║  Passed: ${GREEN}${PASSED}${NC}${BLUE} | Failed: 0                            ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════════════════╝${NC}"
    exit 0
else
    echo -e "${BLUE}║  ${RED}✗ Some linters failed${NC}${BLUE}                       ║${NC}"
    echo -e "${BLUE}║  Passed: ${PASSED}${NC}${BLUE} | Failed: ${RED}${FAILED}${NC}${BLUE}                            ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════════════════╝${NC}"
    exit 1
fi
