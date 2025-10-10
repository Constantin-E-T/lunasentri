# Stage 1: Build
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w -s' -o lunasentri-agent .

# Stage 2: Run
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -S lunasentri && adduser -S -G lunasentri lunasentri

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/lunasentri-agent .

# Change ownership
RUN chown -R lunasentri:lunasentri /app

# Switch to non-root user
USER lunasentri

# Set environment variables (can be overridden)
ENV LUNASENTRI_SERVER_URL=https://api.lunasentri.com
ENV LUNASENTRI_INTERVAL=10s

ENTRYPOINT ["/app/lunasentri-agent"]
