# Multi-stage production build for poker application

# Stage 1: Build Frontend
FROM node:24-alpine AS frontend-builder

WORKDIR /app/frontend

# Copy frontend package files
COPY frontend/package*.json ./

# Install dependencies (use ci for production builds)
RUN npm ci --only=production && npm ci --only=development

# Copy frontend source code
COPY frontend/ ./

# Build the frontend (outputs to ../web/static)
RUN echo "===== STARTING FRONTEND BUILD =====" && \
    npm run build && \
    echo "===== BUILD COMPLETE, CHECKING OUTPUT =====" && \
    ls -la ../web/ && \
    ls -la ../web/static/ && \
    ls -la ../web/static/assets/ && \
    echo "===== ASSET COUNT: $(find ../web/static/assets -type f | wc -l) files =====" && \
    echo "===== TOTAL SIZE: $(du -sh ../web/static) ====="

# Stage 2: Build Backend
FROM golang:1.24-alpine AS backend-builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/

# Build the binary (static binary for Alpine)
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o /app/poker ./cmd/server

# Stage 3: Final Runtime Image
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata wget

WORKDIR /app

# Create non-root user for security
RUN addgroup -g 1000 poker && \
    adduser -D -u 1000 -G poker poker && \
    chown -R poker:poker /app

# Copy backend binary from builder
COPY --from=backend-builder /app/poker .

# Copy healthcheck script
COPY healthcheck.sh .
RUN chmod +x healthcheck.sh

# Copy frontend build from frontend-builder
# The Vite config builds to ../web/static relative to /app/frontend
# So the build output is at /app/web/static
COPY --from=frontend-builder --chown=poker:poker /app/web/static ./web/static/

# Verify assets were copied
RUN echo "===== VERIFYING COPIED ASSETS =====" && \
    ls -la ./web/static/ && \
    ls -la ./web/static/assets/ && \
    echo "===== ASSET FILES =====" && \
    find ./web/static/assets -type f -exec ls -lh {} \; && \
    echo "===== VERIFICATION COMPLETE ====="

# Switch to non-root user
USER poker

# Expose port 8080
EXPOSE 8080

# Set environment variables
ENV PORT=8080 \
    LOG_LEVEL=info

# Health check - use PORT environment variable (defaults to 8080)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ./healthcheck.sh

# Run the binary
CMD ["./poker"]
