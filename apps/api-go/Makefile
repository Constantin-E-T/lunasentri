.PHONY: run build test clean

# Build variables
BINARY_NAME=lunasentri-api
BUILD_DIR=./bin

# Default target
all: build

# Run the application
run:
	go run main.go

# Build the application
build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) .

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Run all checks
check: fmt vet test