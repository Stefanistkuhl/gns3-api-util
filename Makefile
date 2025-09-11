.PHONY: test build clean test-release

# Variables
VERSION ?= $(shell git describe --tags --always --dirty)
BUILD_DIR = dist
BINARY_NAME = gns3util

# Test the project
test:
	go test ./...

# Build for current platform
build:
	go build -o $(BINARY_NAME) .

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 .
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	GOOS=windows GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-arm64.exe .

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f $(BINARY_NAME)

# Create a test release (dry run)
test-release:
	@echo "Creating test release artifacts..."
	@mkdir -p $(BUILD_DIR)
	@$(MAKE) build-all
	@cd $(BUILD_DIR) && for file in $(BINARY_NAME)-*; do \
		if [[ "$$file" == *".exe" ]]; then \
			tar -czf "$$file.tar.gz" "$$file" README.md LICENSE 2>/dev/null || tar -czf "$$file.tar.gz" "$$file"; \
		else \
			tar -czf "$$file.tar.gz" "$$file" README.md LICENSE 2>/dev/null || tar -czf "$$file.tar.gz" "$$file"; \
		fi; \
	done
	@cd $(BUILD_DIR) && for file in *.tar.gz; do \
		sha256sum "$$file" > "$$file.sha256"; \
	done
	@echo "Release artifacts created in $(BUILD_DIR)/"

# Help
help:
	@echo "Available targets:"
	@echo "  test        - Run tests"
	@echo "  build       - Build for current platform"
	@echo "  build-all   - Build for all platforms"
	@echo "  clean       - Clean build artifacts"
	@echo "  test-release- Create test release artifacts"
	@echo "  help        - Show this help"
