# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install git for go mod
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o server ./cmd/server

# Runtime stage
FROM alpine:3.19 AS runtime

# Install runtime dependencies
RUN apk add --no-cache tzdata bash curl mysql-client

# Create non-root user
RUN addgroup -g 1000 appgroup && \
    adduser -u 1000 -G appgroup -s /bin/sh -D appuser

# Create necessary directories
RUN mkdir -p /app/logs && chown -R appuser:appgroup /app

# Copy binary and config from builder
COPY --from=builder /app/server /app/server
COPY --from=builder /app/config.yaml /app/config.yaml
COPY --from=builder /app/db/schema /app/db/schema

# Switch to non-root user
USER appuser

# Expose application port
EXPOSE 1234

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:1234/health || exit 1

# Set working directory
WORKDIR /app

# Run the server
ENTRYPOINT ["/app/server"]
