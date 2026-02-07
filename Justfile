# braindump justfile

RELEASE := "build/release"

# Default recipe to display help
default:
    @just --list

# Build the braindump binary
build:
    go build -o braindump ./cmd/braindump

# Build with version information
build-release version:
    go build -ldflags="-X main.Version={{version}}" -o {{RELEASE}}/braindump ./cmd/braindump

# Run all tests
test:
    go test ./...

# Run tests with verbose output
test-verbose:
    go test -v ./...

# Run tests with coverage
test-coverage:
    go test -cover ./...

# Generate coverage report
coverage:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report generated: coverage.html"

# Lint code using golangci-lint
lint:
    golangci-lint run ./...

# Format code
fmt:
    go fmt ./...

# Check if code is formatted
fmt-check:
    @if [ -n "$$(gofmt -l .)" ]; then \
        echo "The following files are not formatted:"; \
        gofmt -l .; \
        exit 1; \
    fi

# Run go vet
vet:
    go vet ./...

# Run all checks (fmt, vet, lint, test)
check: fmt-check vet lint test

# Clean build artifacts
clean:
    rm -f braindump
    rm -f coverage.out coverage.html

# Install dependencies
deps:
    go mod download
    go mod tidy

# Run the tool with pretty output
run *args:
    go run ./cmd/braindump {{args}}

# Run the tool with example filters
demo:
    @echo "=== All Claude sessions ==="
    go run ./cmd/braindump --agent claude --pretty | head -50
    @echo ""
    @echo "=== Session count ==="
    go run ./cmd/braindump | jq '.sessions | length'

# Install the binary to $GOPATH/bin
install:
    go install ./cmd/braindump

# Show project statistics
stats:
    @echo "=== Code Statistics ==="
    @echo "Go files:"
    @find . -name "*.go" | wc -l
    @echo "Lines of code:"
    @find . -name "*.go" -exec cat {} \; | wc -l
    @echo ""
    @echo "=== Test Coverage ==="
    @go test -cover ./... 2>/dev/null | grep coverage || echo "Run 'just test-coverage' for details"

# Watch for changes and run tests
watch:
    #!/usr/bin/env bash
    while true; do
        inotifywait -e modify -r . --include '.*\.go$' 2>/dev/null
        clear
        just test
    done
