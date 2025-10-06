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

WORKDIR /


# Copy the binary from builder stage
COPY --from=builder /app/lunasentri-api .

# Use non-root user
USER nonroot:nonroot

# Expose port
EXPOSE 8080

# Run the binary
ENTRYPOINT ["./lunasentri-api"]