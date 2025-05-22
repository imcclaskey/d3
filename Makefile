.PHONY: build install clean test release version-patch version-minor version-major print-version update-formula publish-tap push-release

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

# Determine OS for executable suffix
GOOS_OUTPUT = $(shell go env GOOS)
EXTENSION =
ifeq ($(GOOS_OUTPUT),windows)
    EXTENSION = .exe
endif

# Print current version
print-version:
	@echo "Current version: $(VERSION)"

# Build the binary
build:
	mkdir -p $(OUTPUT_DIR)
	go build $(VERSION_FLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME)$(EXTENSION) ./d3

# Install the binary
install:
	go install $(VERSION_FLAGS) ./d3

# Detect if we're on Windows to avoid race detector issues with CGO
COVERAGE_FILE=coverage.out
GOOS=$(shell go env GOOS)
ifeq ($(GOOS),windows)
TEST_RACE_FLAG=
COVER_MODE=count
else
TEST_RACE_FLAG=-race
COVER_MODE=atomic
endif

# Run tests
test:
	@rm -f $(COVERAGE_FILE)
	@echo "Running tests on $(GOOS), race detector: $(if $(TEST_RACE_FLAG),enabled,disabled)"
	go test -v $(TEST_RACE_FLAG) -cover -covermode=$(COVER_MODE) -coverprofile=$(COVERAGE_FILE) ./...

# Generate coverage summary
coverage-summary:
	@echo "Code Coverage Summary:"
	@echo "Total Coverage: $$(go tool cover -func=$(COVERAGE_FILE) | grep total | awk '{print $$3}')"
	@go tool cover -func=$(COVERAGE_FILE)

# Show test coverage in browser (optional helper target)
coverage-html:
	go tool cover -html=$(COVERAGE_FILE)

# Clean build artifacts
clean:
	rm -rf $(OUTPUT_DIR)
	rm -rf $(BUILD_DIR)
	rm -f $(COVERAGE_FILE)

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
_bump-version:
	@sed -i.bak 's/const Version = "$(VERSION)"/const Version = "$(NEW_VERSION)"/g' $(VERSION_FILE)
	@rm -f $(VERSION_FILE).bak
	@echo "New version: $$(grep 'const Version = ' $(VERSION_FILE) | sed 's/.*"\(.*\)".*/\1/')"
	@git add $(VERSION_FILE)
	@git commit -m "Bump version to $(NEW_VERSION)"

version-patch:
	@echo "Bumping patch version: $(VERSION) -> $(NEW_VERSION_PATCH)"
	@$(MAKE) NEW_VERSION=$(NEW_VERSION_PATCH) _bump-version

version-minor:
	@echo "Bumping minor version: $(VERSION) -> $(NEW_VERSION_MINOR)"
	@$(MAKE) NEW_VERSION=$(NEW_VERSION_MINOR) _bump-version

version-major:
	@echo "Bumping major version: $(VERSION) -> $(NEW_VERSION_MAJOR)"
	@$(MAKE) NEW_VERSION=$(NEW_VERSION_MAJOR) _bump-version

# Update Homebrew formula
update-formula: build-all
	@echo "Updating Homebrew formula with version $(VERSION)"
	@./scripts/update_formula.sh
	@echo "Formula updated."

# Publish to Homebrew tap repository
publish-tap: update-formula
	@echo "Publishing to Homebrew tap repository with version $(VERSION)"
	@./scripts/update_tap.sh
	@echo "Tap repository updated."

# Create a new release (builds binaries and creates tag)
release: build-all test
	@echo "Creating release v$(VERSION)"
	@git tag -a v$(VERSION) -m "Release v$(VERSION)"
	@echo "Tag created. Run 'make push-release' to push everything to GitHub."
	@echo "After pushing the release, run 'make publish-tap' to publish to your Homebrew tap."

# Push new version
push-release:
	@echo "Pushing to GitHub..."
	@git push origin main && git push origin --tags

# Full release workflow
release-all: release push-release publish-tap
	@echo "✅ Complete release process finished!"
	@echo "Version $(VERSION) has been:"
	@echo "  • Tagged and pushed to GitHub"
	@echo "  • Released with binaries via GitHub Actions"
	@echo "  • Published to Homebrew tap" 