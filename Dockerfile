# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=1.0.0" \
    -o nomad-agent \
    ./cmd/nomad

# Final stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 nomad && \
    adduser -u 1000 -G nomad -s /bin/sh -D nomad

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/nomad-agent /app/nomad-agent

# Copy static files for webchat (if they exist)
COPY --from=builder /app/web/dist /app/web/dist 2>/dev/null || true

# Set ownership
RUN chown -R nomad:nomad /app

# Switch to non-root user
USER nomad

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget -q --spider http://localhost:8080/health || exit 1

# Run the agent
ENTRYPOINT ["/app/nomad-agent"]
