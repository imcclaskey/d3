.PHONY: build install clean test release version-patch version-minor version-major print-version

# Binary name
BINARY_NAME=d3
OUTPUT_DIR=bin
BUILD_DIR=build
VERSION_FILE=internal/version/version.go

# Extract version components
VERSION=$(shell grep 'const Version = ' $(VERSION_FILE) | sed 's/.*"\(.*\)".*/\1/')
VERSION_MAJOR=$(shell echo $(VERSION) | cut -d. -f1)
VERSION_MINOR=$(shell echo $(VERSION) | cut -d. -f2)
VERSION_PATCH=$(shell echo $(VERSION) | cut -d. -f3)

# New versions
NEW_VERSION_PATCH=$(VERSION_MAJOR).$(VERSION_MINOR).$(shell expr $(VERSION_PATCH) + 1)
NEW_VERSION_MINOR=$(VERSION_MAJOR).$(shell expr $(VERSION_MINOR) + 1).0
NEW_VERSION_MAJOR=$(shell expr $(VERSION_MAJOR) + 1).0.0

# Version flags
VERSION_FLAGS=-ldflags "-X github.com/imcclaskey/d3/internal/version.Version=$(VERSION)"

# Print current version
print-version:
	@echo "Current version: $(VERSION)"

# Build the binary
build:
	mkdir -p $(OUTPUT_DIR)
	go build $(VERSION_FLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME) ./d3

# Install the binary
install:
	go install $(VERSION_FLAGS) ./d3

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
	GOOS=darwin GOARCH=amd64 go build $(VERSION_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./d3
	GOOS=darwin GOARCH=arm64 go build $(VERSION_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./d3
	
	# Linux (amd64 and arm64)
	GOOS=linux GOARCH=amd64 go build $(VERSION_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./d3
	GOOS=linux GOARCH=arm64 go build $(VERSION_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./d3
	
	# Windows (amd64)
	GOOS=windows GOARCH=amd64 go build $(VERSION_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./d3

# Version bumping targets
version-patch:
	@echo "Bumping patch version: $(VERSION) -> $(NEW_VERSION_PATCH)"
	@sed -i.bak 's/const Version = "$(VERSION)"/const Version = "$(NEW_VERSION_PATCH)"/g' $(VERSION_FILE)
	@rm -f $(VERSION_FILE).bak
	@$(MAKE) post-version

version-minor:
	@echo "Bumping minor version: $(VERSION) -> $(NEW_VERSION_MINOR)"
	@sed -i.bak 's/const Version = "$(VERSION)"/const Version = "$(NEW_VERSION_MINOR)"/g' $(VERSION_FILE)
	@rm -f $(VERSION_FILE).bak
	@$(MAKE) post-version

version-major:
	@echo "Bumping major version: $(VERSION) -> $(NEW_VERSION_MAJOR)"
	@sed -i.bak 's/const Version = "$(VERSION)"/const Version = "$(NEW_VERSION_MAJOR)"/g' $(VERSION_FILE)
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

# Push new version
push-release:
	@echo "Pushing to GitHub..."
	@git push origin main && git push origin --tags 