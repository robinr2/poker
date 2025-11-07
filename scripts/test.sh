#!/bin/bash

# Comprehensive testing script for the poker application
# Runs all tests for Go backend and React frontend

set -e

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
TOTAL_TESTS=0

echo -e "${BLUE}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║        Poker Application - Test Suite                  ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════╝${NC}"
echo ""

# Go tests
echo -e "${YELLOW}→ Running Go Backend Tests${NC}"
echo ""

cd "$PROJECT_ROOT"
if go test ./internal/... ./cmd/... -v -race 2>&1 | tee /tmp/go_test.log; then
    GO_TEST_RESULT=0
    GO_PASS_COUNT=$(grep -c "^=== RUN" /tmp/go_test.log || echo "?")
    echo -e "    ${GREEN}✓ Go tests passed${NC}"
else
    GO_TEST_RESULT=1
    echo -e "    ${RED}✗ Go tests failed${NC}"
    ((FAILED++))
fi

echo ""
echo -e "${YELLOW}→ Running Frontend Tests${NC}"
echo ""

cd "$PROJECT_ROOT/frontend"
if npm test -- --run 2>&1 | tee /tmp/frontend_test.log; then
    FRONTEND_TEST_RESULT=0
    # Extract test count from vitest output
    FRONTEND_PASS_COUNT=$(grep -oP "Tests\s+\K\d+" /tmp/frontend_test.log | head -1 || echo "?")
    echo -e "    ${GREEN}✓ Frontend tests passed${NC}"
else
    FRONTEND_TEST_RESULT=1
    echo -e "    ${RED}✗ Frontend tests failed${NC}"
    ((FAILED++))
fi

echo ""
echo -e "${BLUE}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║                   Test Summary                         ║${NC}"
echo -e "${BLUE}╠════════════════════════════════════════════════════════╣${NC}"

if [ $FAILED -eq 0 ] && [ $GO_TEST_RESULT -eq 0 ] && [ $FRONTEND_TEST_RESULT -eq 0 ]; then
    echo -e "${BLUE}║  Backend Tests (Go):                                 ║${NC}"
    echo -e "${BLUE}║    ${GREEN}✓ Passed${NC}${BLUE}                                   ║${NC}"
    echo -e "${BLUE}║                                                    ║${NC}"
    echo -e "${BLUE}║  Frontend Tests (Vitest):                            ║${NC}"
    echo -e "${BLUE}║    ${GREEN}✓ Passed (${FRONTEND_PASS_COUNT} tests)${NC}${BLUE}                           ║${NC}"
    echo -e "${BLUE}║                                                    ║${NC}"
    echo -e "${BLUE}║  ${GREEN}✓ All tests passed!${NC}${BLUE}                          ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════════════════╝${NC}"
    exit 0
else
    echo -e "${BLUE}║  ${RED}✗ Some tests failed${NC}${BLUE}                         ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════════════════╝${NC}"
    exit 1
fi
