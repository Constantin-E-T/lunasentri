# Build stage
FROM golang:1.22 AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o lunasentri-api .

# Final stage
FROM gcr.io/distroless/static:nonroot

WORKDIR /app

# Create data directory for SQLite database
# Note: In production, mount a volume to /app/data for persistence
USER nonroot:nonroot

# Copy the binary from builder stage
COPY --from=builder /app/lunasentri-api .

# Expose port
EXPOSE 8080

# Run the binary
ENTRYPOINT ["./lunasentri-api"]