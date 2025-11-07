#!/bin/bash
set -e
echo "=== Building Frontend ==="
cd frontend
npm run build
cd ..
echo ""
echo "=== Starting Backend Server ==="
echo "Server will serve frontend at http://localhost:8080"
echo "Press Ctrl+C to stop"
go run cmd/server/main.go
