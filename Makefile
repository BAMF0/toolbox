# Makefile for ToolBox (tb)

.PHONY: all build test clean install uninstall install-man uninstall-man help

# Variables
BINARY_NAME=tb
BUILD_DIR=.
INSTALL_PREFIX=/usr/local
MAN_DIR=$(INSTALL_PREFIX)/share/man
VERSION?=$(shell cat VERSION 2>/dev/null || echo "0.1.0")
GIT_COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-X 'github.com/bamf0/toolbox/internal/cli.Version=$(VERSION)' \
        -X 'github.com/bamf0/toolbox/internal/cli.GitCommit=$(GIT_COMMIT)' \
        -X 'github.com/bamf0/toolbox/internal/cli.BuildTime=$(BUILD_TIME)'

all: build

# Build the binary
build:
	@echo "Building $(BINARY_NAME) version $(VERSION)..."
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/tb

# Build for release (with optimizations)
build-release:
	@echo "Building $(BINARY_NAME) $(VERSION) for release..."
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS) -s -w" -trimpath -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/tb

# Run tests
test:
	@echo "Running tests..."
	go test ./... -v

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test ./... -cover -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BUILD_DIR)/$(BINARY_NAME)
	rm -f coverage.out coverage.html

# Install binary
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_PREFIX)/bin..."
	install -d $(INSTALL_PREFIX)/bin
	install -m 755 $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_PREFIX)/bin/

# Uninstall binary
uninstall:
	@echo "Uninstalling $(BINARY_NAME) from $(INSTALL_PREFIX)/bin..."
	rm -f $(INSTALL_PREFIX)/bin/$(BINARY_NAME)

# Install man pages (system-wide, requires sudo)
install-man:
	@echo "Installing man pages to $(MAN_DIR)..."
	install -d $(MAN_DIR)/man1
	install -d $(MAN_DIR)/man5
	install -m 644 docs/man/*.1 $(MAN_DIR)/man1/
	install -m 644 docs/man/*.5 $(MAN_DIR)/man5/
	@if command -v mandb >/dev/null 2>&1; then \
		echo "Updating man database..."; \
		mandb -q 2>/dev/null || true; \
	elif command -v makewhatis >/dev/null 2>&1; then \
		echo "Updating man database..."; \
		makewhatis 2>/dev/null || true; \
	fi
	@echo "Man pages installed. Try: man tb"

# Install man pages for current user (no root required)
install-man-user:
	@./docs/man/install-man.sh --user

# Uninstall man pages
uninstall-man:
	@echo "Uninstalling man pages from $(MAN_DIR)..."
	rm -f $(MAN_DIR)/man1/tb*.1
	rm -f $(MAN_DIR)/man5/tb*.5
	@if command -v mandb >/dev/null 2>&1; then \
		mandb -q 2>/dev/null || true; \
	elif command -v makewhatis >/dev/null 2>&1; then \
		makewhatis 2>/dev/null || true; \
	fi

# Full install (binary + man pages)
install-all: install install-man
	@echo "Installation complete!"

# Full uninstall (binary + man pages)
uninstall-all: uninstall uninstall-man
	@echo "Uninstallation complete!"

# Development: rebuild and run
dev: build
	@./$(BINARY_NAME)

# Run linter
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install from https://golangci-lint.run/"; \
		exit 1; \
	fi

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Show help
help:
	@echo "ToolBox (tb) Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build              Build the binary (version $(VERSION))"
	@echo "  make build-release      Build optimized release binary"
	@echo "  make test               Run tests"
	@echo "  make test-coverage      Run tests with coverage report"
	@echo "  make clean              Clean build artifacts"
	@echo "  make install            Install binary (requires sudo)"
	@echo "  make install-man        Install man pages (requires sudo)"
	@echo "  make install-man-user   Install man pages for current user"
	@echo "  make install-all        Install binary and man pages (requires sudo)"
	@echo "  make uninstall          Uninstall binary"
	@echo "  make uninstall-man      Uninstall man pages"
	@echo "  make uninstall-all      Uninstall binary and man pages"
	@echo "  make lint               Run golangci-lint"
	@echo "  make fmt                Format code with go fmt"
	@echo "  make dev                Build and run (for development)"
	@echo "  make help               Show this help message"
	@echo ""
	@echo "Variables:"
	@echo "  VERSION=$(VERSION)      Override version (default: 0.1.0)"
	@echo ""
	@echo "Examples:"
	@echo "  make                    # Build the binary"
	@echo "  make test               # Run tests"
	@echo "  make VERSION=0.2.0 build # Build with custom version"
	@echo "  make build-release      # Build optimized release binary"
	@echo "  sudo make install-all   # Install everything"
	@echo "  make install-man-user   # Install man pages for current user (no sudo)"
