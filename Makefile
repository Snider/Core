.PHONY: help build test test-coverage test-bench lint fmt vet clean example example-ui install-tools

# Default target
help:
	@echo "Available targets:"
	@echo "  make build          - Build the project"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage"
	@echo "  make test-bench     - Run benchmark tests"
	@echo "  make lint           - Run golangci-lint"
	@echo "  make fmt            - Format code with gofmt"
	@echo "  make vet            - Run go vet"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make example        - Run the example program (CLI demo)"
	@echo "  make example-ui     - Run the Wails UI example"
	@echo "  make install-tools  - Install development tools"

# Build the project
build:
	@echo "Building..."
	@go build -v .

# Run tests
test:
	@echo "Running tests..."
	@go test -v -race .

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.txt -covermode=atomic .
	@go tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run benchmark tests
test-bench:
	@echo "Running benchmark tests..."
	@go test -bench=. -benchmem .

# Run golangci-lint (requires installation)
lint:
	@echo "Running golangci-lint..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Run 'make install-tools'" && exit 1)
	@golangci-lint run .

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt .

# Run go vet
vet:
	@echo "Running go vet..."
	@go vet .

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f coverage.txt coverage.html
	@go clean -cache -testcache -modcache
	@rm -rf bin/

# Run the example program
example:
	@echo "Running example..."
	@cd examples && go run demo.go

# Run the Wails UI example (requires Wails v3 with UI support)
example-ui:
	@echo "Running Wails UI example..."
	@cd examples && go run main.go

# Install development tools
install-tools:
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Tools installed successfully"
