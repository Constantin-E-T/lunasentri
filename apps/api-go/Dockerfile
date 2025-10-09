# Build stage
FROM golang:1.24 AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o lunasentri-api .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

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

# Expose port
EXPOSE 8080

# Run the binary
ENTRYPOINT ["./lunasentri-api"]