# Multi-stage build for poker application

# Stage 1: Builder
FROM golang:1.24-alpine AS builder

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

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/poker ./cmd/server

# Stage 2: Runtime
FROM alpine:latest

# Install runtime dependencies (none required for static binary)
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Create non-root user for security
RUN addgroup -g 1000 poker && \
    adduser -D -u 1000 -G poker poker && \
    chown -R poker:poker /app

# Copy binary from builder
COPY --from=builder /app/poker .

# Copy static assets
COPY --chown=poker:poker web/ ./web/

# Switch to non-root user
USER poker

# Expose port 8080
EXPOSE 8080

# Set environment variables
ENV PORT=8080 \
    LOG_LEVEL=info

# Run the binary
CMD ["./poker"]
