#!/bin/sh
# Healthcheck script that uses the PORT environment variable
wget --no-verbose --tries=1 -O /dev/null "http://localhost:${PORT:-8080}/health" || exit 1
