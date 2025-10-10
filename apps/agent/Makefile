.PHONY: build clean test install docker-build docker-run help

# Build variables
BINARY_NAME=lunasentri-agent
VERSION=1.0.0
BUILD_DIR=dist
GO=go
DOCKER=docker

# Build the binary
build:
	@echo "Building $(BINARY_NAME) v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for Linux (useful when building from macOS)
build-linux:
	@echo "Building $(BINARY_NAME) v$(VERSION) for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64"

# Run tests
test:
	@echo "Running tests..."
	$(GO) test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	@echo "Clean complete"

# Install to system (requires sudo)
install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "Installation complete"

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	$(DOCKER) build -t lunasentri/agent:$(VERSION) -t lunasentri/agent:latest .
	@echo "Docker build complete"

# Run in Docker (interactive)
docker-run:
	@echo "Running agent in Docker..."
	$(DOCKER) run --rm -it \
		-e LUNASENTRI_SERVER_URL=${LUNASENTRI_SERVER_URL} \
		-e LUNASENTRI_API_KEY=${LUNASENTRI_API_KEY} \
		lunasentri/agent:latest

# Help
help:
	@echo "LunaSentri Agent - Build Commands"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build         Build the agent binary"
	@echo "  build-linux   Build for Linux (cross-compile)"
	@echo "  test          Run tests"
	@echo "  clean         Remove build artifacts"
	@echo "  install       Install to /usr/local/bin (requires sudo)"
	@echo "  docker-build  Build Docker image"
	@echo "  docker-run    Run agent in Docker container"
	@echo "  help          Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make test"
	@echo "  LUNASENTRI_API_KEY=xxx make docker-run"
