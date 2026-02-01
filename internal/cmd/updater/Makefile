.PHONY: build dev release-local test coverage

BINARY_NAME=updater
CMD_PATH=./cmd/updater

# Default LDFLAGS to empty
LDFLAGS = ""

# If VERSION is set, override LDFLAGS
ifdef VERSION
	LDFLAGS = -ldflags "-X 'github.com/snider/updater.Version=$(VERSION)'"
endif

.PHONY: generate
generate:
	@echo "Generating code..."
	@go generate ./...

build: generate
	@echo "Building $(BINARY_NAME)..."
	@cd $(CMD_PATH) && go build $(LDFLAGS)

dev: build
	@echo "Running $(BINARY_NAME)..."
	@$(CMD_PATH)/$(BINARY_NAME) --check-update

release-local:
	@echo "Running local release with GoReleaser..."
	@~/go/bin/goreleaser release --snapshot --clean

test:
	@echo "Running tests..."
	@go test ./...

coverage:
	@echo "Generating code coverage report..."
	@go test -coverprofile=coverage.out ./...
	@echo "Coverage report generated: coverage.out"
	@echo "To view in browser: go tool cover -html=coverage.out"
	@echo "To upload to Codecov, ensure you have the Codecov CLI installed (e.g., 'go install github.com/codecov/codecov-cli@latest') and run: codecov -f coverage.out"
