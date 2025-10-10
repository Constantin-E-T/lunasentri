# Build stage
FROM golang:1.24 AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application from cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o lunasentri-api ./cmd/api

# Final stage
FROM alpine:latest

# Install ca-certificates and curl for HTTPS requests and healthcheck
RUN apk --no-cache add ca-certificates curl

WORKDIR /app

# Create non-root user and data directory
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    mkdir -p /app/data && \
    chown -R appuser:appuser /app

# Copy the binary from builder stage
COPY --from=builder /app/lunasentri-api .

# Switch to non-root user
USER appuser

# Expose port 80
EXPOSE 80

# Health check for CapRover
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
  CMD curl -fsS http://localhost:80/health || exit 1

# Run the binary
ENTRYPOINT ["./lunasentri-api"]