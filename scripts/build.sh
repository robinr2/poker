#!/bin/bash
set -e

echo "=== Poker Application Build Script ==="
echo ""

# Build frontend
echo "Step 1: Building frontend..."
cd frontend
npm run build
cd ..
echo "✓ Frontend build complete"
echo ""

# Build backend
echo "Step 2: Building backend..."
go build -o bin/poker cmd/server/main.go
echo "✓ Backend build complete"
echo ""

echo "=== Build successful ==="
echo "Binary location: bin/poker"
echo "Frontend location: web/static/"
