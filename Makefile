.PHONY: build install clean test release version-patch version-minor version-major

# Binary name
BINARY_NAME=i3
OUTPUT_DIR=bin
BUILD_DIR=build
VERSION=$(shell grep 'const Version = ' internal/version/version.go | sed 's/.*"\(.*\)".*/\1/')
VERSION_FILE=internal/version/version.go

# Version components
VERSION_MAJOR=$(shell echo $(VERSION) | cut -d. -f1)
VERSION_MINOR=$(shell echo $(VERSION) | cut -d. -f2)
VERSION_PATCH=$(shell echo $(VERSION) | cut -d. -f3)

# Version flags
VERSION_FLAGS=-ldflags "-X github.com/imcclaskey/i3/internal/version.Version=$(VERSION)"

# Build the binary
build:
	mkdir -p $(OUTPUT_DIR)
	go build $(VERSION_FLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME) ./i3

# Install the binary
install:
	go install $(VERSION_FLAGS) ./i3

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
	GOOS=darwin GOARCH=amd64 go build $(VERSION_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./i3
	GOOS=darwin GOARCH=arm64 go build $(VERSION_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./i3
	
	# Linux (amd64 and arm64)
	GOOS=linux GOARCH=amd64 go build $(VERSION_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./i3
	GOOS=linux GOARCH=arm64 go build $(VERSION_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./i3
	
	# Windows (amd64)
	GOOS=windows GOARCH=amd64 go build $(VERSION_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./i3

# Version bumping targets
version-patch:
	@echo "Bumping patch version: $(VERSION) -> $(VERSION_MAJOR).$(VERSION_MINOR).$$(( $(VERSION_PATCH) + 1 ))"
	@sed -i.bak 's/const Version = "$(VERSION)"/const Version = "$(VERSION_MAJOR).$(VERSION_MINOR).$$(( $(VERSION_PATCH) + 1 ))"/g' $(VERSION_FILE)
	@rm -f $(VERSION_FILE).bak
	@$(MAKE) post-version

version-minor:
	@echo "Bumping minor version: $(VERSION) -> $(VERSION_MAJOR).$$(( $(VERSION_MINOR) + 1 )).0"
	@sed -i.bak 's/const Version = "$(VERSION)"/const Version = "$(VERSION_MAJOR).$$(( $(VERSION_MINOR) + 1 )).0"/g' $(VERSION_FILE)
	@rm -f $(VERSION_FILE).bak
	@$(MAKE) post-version

version-major:
	@echo "Bumping major version: $(VERSION) -> $$(( $(VERSION_MAJOR) + 1 )).0.0"
	@sed -i.bak 's/const Version = "$(VERSION)"/const Version = "$$(( $(VERSION_MAJOR) + 1 )).0.0"/g' $(VERSION_FILE)
	@rm -f $(VERSION_FILE).bak
	@$(MAKE) post-version

# Common post-version tasks
post-version: build test
	@echo "New version: $$(grep 'const Version = ' $(VERSION_FILE) | sed 's/.*"\(.*\)".*/\1/')"
	@read -p "Commit and tag new version? [y/N] " confirm; \
	if [ "$$confirm" = "y" ] || [ "$$confirm" = "Y" ]; then \
		NEW_VERSION=$$(grep 'const Version = ' $(VERSION_FILE) | sed 's/.*"\(.*\)".*/\1/'); \
		git add $(VERSION_FILE); \
		git commit -m "Bump version to $$NEW_VERSION"; \
		git tag -a "v$$NEW_VERSION" -m "Release v$$NEW_VERSION"; \
		echo "Version $$NEW_VERSION ready to release. Run:"; \
		echo "  git push origin main && git push origin v$$NEW_VERSION"; \
	else \
		echo "Version update cancelled."; \
	fi

# Create a new release
release: build-all
	@echo "Creating release v$(VERSION)"
	@git tag -a v$(VERSION) -m "Release v$(VERSION)"
	@echo "Tag created. Run 'git push origin v$(VERSION)' to trigger the GitHub workflow" 