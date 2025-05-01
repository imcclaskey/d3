.PHONY: build install clean test

# Binary name
BINARY_NAME=i3
OUTPUT_DIR=bin
BUILD_DIR=build

# Build the binary
build:
	mkdir -p $(OUTPUT_DIR)
	go build -o $(OUTPUT_DIR)/$(BINARY_NAME) ./i3

# Install the binary
install:
	go install ./i3

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf $(OUTPUT_DIR)
	rm -rf $(BUILD_DIR)

# Build for all platforms
build-all: clean
	mkdir -p $(BUILD_DIR)
	
	# Mac (amd64 and arm64)
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./i3
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./i3
	
	# Linux (amd64 and arm64)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./i3
	GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./i3
	
	# Windows (amd64)
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./i3 