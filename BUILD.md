# Build and Run Instructions

## Prerequisites

- Go 1.21 or later
- Git (for version control)

## Quick Start

### 1. Build the Tool

```bash
# From the project root
go build -o tb ./cmd/tb

# Or build and install to $GOPATH/bin
go install ./cmd/tb
```

### 2. Run the Tool

```bash
# Run directly from project directory
./tb build --dry-run

# Or if installed to $GOPATH/bin (and $GOPATH/bin is in your PATH)
tb build --dry-run
```

### 3. Test in Different Project Types

**In a Go project (like this one):**
```bash
./tb build --dry-run    # Outputs: go build
./tb test --dry-run     # Outputs: go test ./...
./tb fmt --dry-run      # Outputs: go fmt ./...
```

**In a Node.js project:**
```bash
# Create a test Node.js project
mkdir /tmp/test-node && cd /tmp/test-node
echo '{"name":"test"}' > package.json

# Use tb with forced context
/path/to/tb build --dry-run    # Outputs: npm run build
/path/to/tb test --dry-run     # Outputs: npm test
```

**Override context manually:**
```bash
./tb build --context python --dry-run   # Uses Python context
./tb test --context rust --dry-run      # Uses Rust context
```

## Installation Options

### Option 1: Build and Copy to PATH

```bash
go build -o tb ./cmd/tb
sudo cp tb /usr/local/bin/
# Now you can use 'tb' from anywhere
```

### Option 2: Install to GOPATH

```bash
go install ./cmd/tb
# Installs to $GOPATH/bin/tb
# Make sure $GOPATH/bin is in your PATH
```

### Option 3: Build for Specific OS/Architecture

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o tb-linux-amd64 ./cmd/tb

# macOS ARM64 (M1/M2)
GOOS=darwin GOARCH=arm64 go build -o tb-darwin-arm64 ./cmd/tb

# Windows
GOOS=windows GOARCH=amd64 go build -o tb.exe ./cmd/tb
```

## Development Workflow

### Run Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests verbosely
go test -v ./...
```

### Format Code

```bash
go fmt ./...
```

### Lint Code (requires golangci-lint)

```bash
golangci-lint run
```

### Development Build

```bash
# Quick rebuild during development
go build -o tb ./cmd/tb && ./tb --help
```

## Configuration

### Default Configuration

The tool comes with built-in defaults for common contexts (Node.js, Go, Python, Rust, etc.). No configuration file is needed to get started.

### Custom Configuration

Create a `.toolbox.yaml` in your project root or `~/.toolbox/config.yaml` for global settings:

```yaml
contexts:
  node:
    commands:
      build: "npm run build"
      test: "npm test"
      custom: "npm run custom-script"
  
  myapp:
    commands:
      deploy: "kubectl apply -f deploy.yaml"
      logs: "kubectl logs -f myapp"
```

### Configuration Priority

1. File specified with `--config` flag
2. `.toolbox.yaml` in current directory
3. `~/.toolbox/config.yaml` in home directory
4. Built-in defaults

## Usage Examples

### Basic Commands

```bash
# Auto-detect context and run build
tb build

# Dry-run to see what would execute
tb build --dry-run

# Verbose output showing detection
tb test --verbose

# Force a specific context
tb build --context node

# Pass additional arguments to the underlying command
tb test --verbose -- --coverage
```

### Project-Specific Setup

```bash
# Create a local config for your project
cat > .toolbox.yaml << 'EOF'
contexts:
  go:
    commands:
      build: "go build -ldflags='-s -w' -o bin/myapp"
      deploy: "rsync -av bin/myapp user@server:/opt/myapp/"
EOF

# Now 'tb deploy' will run your custom deploy command
tb deploy
```

### Global Configuration

```bash
# Create global config directory
mkdir -p ~/.toolbox

# Create global config
cat > ~/.toolbox/config.yaml << 'EOF'
contexts:
  python:
    commands:
      test: "pytest -v"
      lint: "ruff check . && mypy ."
EOF
```

## Troubleshooting

### "No recognized project context found"

This means tb couldn't find marker files (package.json, go.mod, etc.) in the current directory or parent directories.

**Solutions:**
- Use `--context` flag to force a context: `tb build --context go`
- Ensure you're in a project directory with appropriate marker files
- Create a `.toolbox.yaml` to define a custom context

### Command not found in context

The command you're trying to run isn't defined for the detected context.

**Solutions:**
- Check available commands with `--dry-run --verbose`
- Add the command to your `.toolbox.yaml`
- Use `--context` to switch to a different context

### Module import errors during build

Run `go mod tidy` to ensure dependencies are up to date:
```bash
go mod tidy
go build -o tb ./cmd/tb
```

## Next Steps

1. Install the tool to your PATH
2. Create a `.toolbox.yaml` in your most-used projects
3. Try it with different project types
4. Customize commands to match your workflow
5. Share configurations with your team

## Uninstallation

```bash
# If installed to /usr/local/bin
sudo rm /usr/local/bin/tb

# If installed via go install
rm $GOPATH/bin/tb

# Remove global config (optional)
rm -rf ~/.toolbox
```
