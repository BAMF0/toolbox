# User Guide

Complete guide to using ToolBox (tb) for context-aware command management.

## Table of Contents

- [Introduction](#introduction)
- [Installation](#installation)
- [Basic Usage](#basic-usage)
- [Supported Contexts](#supported-contexts)
- [Command Flags](#command-flags)
- [Configuration](#configuration)
- [Common Workflows](#common-workflows)
- [Troubleshooting](#troubleshooting)

## Introduction

ToolBox is a context-aware CLI tool that provides intelligent command shortcuts based on your project type. Instead of remembering different commands for different projects, just use `tb <command>` and ToolBox automatically runs the right command for your project.

### Key Benefits

- **Consistent Interface**: Same commands across all project types
- **Zero Configuration**: Works out-of-the-box for common projects
- **Context-Aware**: Automatically detects your project type
- **Extensible**: Easily add custom commands and contexts
- **Developer Friendly**: Built-in help, autocompletion, and dry-run mode

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/bamf0/toolbox.git
cd toolbox

# Build the binary
go build -o tb ./cmd/tb

# Install to your PATH
sudo cp tb /usr/local/bin/
# or
go install ./cmd/tb
```

### Verify Installation

```bash
tb --version
tb --help
```

### Set Up Autocompletion (Optional but Recommended)

See [Autocompletion Guide](autocompletion.md) for detailed setup instructions.

Quick setup for Bash:
```bash
tb completion bash | sudo tee /etc/bash_completion.d/tb
source ~/.bashrc
```

## Basic Usage

### Running Commands

The simplest way to use ToolBox:

```bash
# Navigate to your project directory
cd ~/my-project

# Run a command - ToolBox automatically detects your project type
tb build
tb test
tb start
```

### Dry Run Mode

Preview what command would execute without running it:

```bash
tb build --dry-run
# Output:
# Context: node
# Command: npm run build
```

This is great for:
- Learning what ToolBox does
- Debugging configuration issues
- Verifying context detection

### Verbose Mode

See detailed information about context detection:

```bash
tb test -v
# Output:
# Detected context: go (marker: go.mod)
# Running command: go test ./...
```

### Getting Help

```bash
# General help
tb --help

# Help for a specific command in your current context
tb help build

# Help for a command in a specific context
tb help --context python test
```

## Supported Contexts

ToolBox automatically detects your project type based on marker files:

### Node.js

**Detected by**: `package.json`

**Available commands**:
```bash
tb build    # npm run build
tb test     # npm test
tb start    # npm start
tb dev      # npm run dev
tb lint     # npm run lint
tb install  # npm install
```

### Go

**Detected by**: `go.mod`

**Available commands**:
```bash
tb build    # go build ./...
tb test     # go test ./...
tb run      # go run ./cmd/...
tb fmt      # go fmt ./...
tb lint     # golangci-lint run
tb install  # go mod download
```

### Python

**Detected by**: `pyproject.toml` or `requirements.txt`

**Available commands**:
```bash
tb test     # pytest
tb lint     # ruff check .
tb fmt      # black .
tb install  # pip install -r requirements.txt
tb run      # python main.py
```

### Rust

**Detected by**: `Cargo.toml`

**Available commands**:
```bash
tb build    # cargo build
tb test     # cargo test
tb run      # cargo run
tb lint     # cargo clippy
tb fmt      # cargo fmt
tb install  # cargo fetch
```

### Make

**Detected by**: `Makefile`

**Available commands**:
```bash
tb build    # make
tb test     # make test
tb clean    # make clean
```

### Ubuntu Packaging (Plugin)

**Detected by**: `debian/control` or `debian/changelog`

**Available commands**:
```bash
tb gbranch      # Create git branch for bug fix
tb ppa-status   # Show PPA information
tb dch-auto     # Auto-update changelog
tb ubuild       # Build and upload to PPA
tb build        # dpkg-buildpackage -us -uc
tb lint         # lintian
```

See the [Ubuntu Plugin documentation](../internal/plugin/ubuntu.go) for details.

## Command Flags

### Global Flags

Available with any command:

```bash
--context <name>     # Force a specific context
--config <file>      # Use a custom config file
--dry-run            # Preview command without executing
-v, --verbose        # Show detailed output
-h, --help          # Show help
```

### Examples

```bash
# Force Python context in a Node.js project
tb --context python test

# Use custom config file
tb --config .tb-custom.yaml build

# Combine flags
tb --context go --dry-run build
```

## Configuration

### Configuration Hierarchy

ToolBox loads configuration in this order (first found wins):

1. File specified with `--config` flag
2. `.toolbox.yaml` in current directory
3. `~/.toolbox/config.yaml` in home directory
4. Built-in defaults

### Creating a Local Config

Create `.toolbox.yaml` in your project root:

```yaml
contexts:
  node:
    commands:
      build: "npm run build:prod"
      test: "npm test -- --coverage"
      deploy: "npm run deploy"
    descriptions:
      build: "Production build with optimization"
      test: "Run tests with coverage report"
      deploy: "Deploy to production server"
```

### Creating a Global Config

Create `~/.toolbox/config.yaml` for user-wide settings:

```bash
mkdir -p ~/.toolbox
cat > ~/.toolbox/config.yaml <<EOF
contexts:
  python:
    commands:
      test: "pytest -v --cov"
      lint: "ruff check . && mypy ."
      fmt: "black . && isort ."
EOF
```

### Custom Contexts

Add your own contexts for specialized workflows:

```yaml
contexts:
  docker:
    commands:
      up: "docker-compose up -d"
      down: "docker-compose down"
      logs: "docker-compose logs -f"
      build: "docker-compose build"
    descriptions:
      up: "Start containers in background"
      down: "Stop and remove containers"
      logs: "Follow container logs"
      build: "Build or rebuild services"
```

Then use it:

```bash
tb --context docker up
```

### Override Built-in Commands

Customize default commands for your workflow:

```yaml
contexts:
  go:
    commands:
      # Override default build command
      build: "go build -o bin/myapp ./cmd/myapp"
      # Add custom commands
      docker-build: "docker build -t myapp:latest ."
```

## Common Workflows

### Monorepo with Multiple Projects

Create `.toolbox.yaml` in each subproject:

```
monorepo/
├── frontend/
│   ├── package.json
│   └── .toolbox.yaml    # Node.js config
├── backend/
│   ├── go.mod
│   └── .toolbox.yaml    # Go config
└── services/
    ├── Dockerfile
    └── .toolbox.yaml    # Docker config
```

Navigate to each directory and use the same commands:
```bash
cd frontend && tb build
cd backend && tb build
cd services && tb build
```

### CI/CD Integration

Use ToolBox in CI pipelines for consistency:

```yaml
# .github/workflows/ci.yml
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Install ToolBox
        run: go install github.com/bamf0/toolbox/cmd/tb@latest
      - name: Build
        run: tb build
      - name: Test
        run: tb test
```

### Multi-Language Projects

Force context when you have multiple languages:

```bash
# Build frontend (Node.js)
cd frontend && tb build

# Build backend (Go)
cd backend && tb build

# Or from root with explicit context
tb --context node build
cd backend && tb --context go build
```

### Development Workflow

Typical day-to-day usage:

```bash
# Start development
cd ~/project
tb install    # Install dependencies
tb dev        # Start dev server

# Make changes, then test
tb lint       # Check code style
tb test       # Run tests
tb build      # Build for production

# Deploy
tb deploy     # Deploy (if configured)
```

## Troubleshooting

### Command Not Found

**Problem**: `tb: command not found`

**Solutions**:
1. Verify installation:
   ```bash
   which tb
   ```

2. Add to PATH:
   ```bash
   export PATH="$PATH:/usr/local/bin"
   # Add to ~/.bashrc or ~/.zshrc to persist
   ```

3. Reinstall:
   ```bash
   go install github.com/bamf0/toolbox/cmd/tb@latest
   ```

### Context Not Detected

**Problem**: ToolBox doesn't detect your project type

**Solutions**:
1. Check for marker files:
   ```bash
   # Node.js needs package.json
   # Go needs go.mod
   # Python needs pyproject.toml or requirements.txt
   # etc.
   ```

2. Use verbose mode to debug:
   ```bash
   tb build -v
   ```

3. Force context manually:
   ```bash
   tb --context node build
   ```

4. Create `.toolbox.yaml` with explicit context:
   ```yaml
   contexts:
     my-context:
       commands:
         build: "make"
   ```

### Wrong Command Executed

**Problem**: ToolBox runs the wrong command

**Solutions**:
1. Check context detection:
   ```bash
   tb build --dry-run
   ```

2. Verify configuration:
   ```bash
   # Check local config
   cat .toolbox.yaml
   
   # Check global config
   cat ~/.toolbox/config.yaml
   ```

3. Use specific config file:
   ```bash
   tb --config custom-config.yaml build
   ```

### Command Fails

**Problem**: Command executes but fails

**Solutions**:
1. Run in dry-run mode to see the actual command:
   ```bash
   tb build --dry-run
   ```

2. Run the command directly to verify it works:
   ```bash
   npm run build  # or whatever the actual command is
   ```

3. Check command output for errors:
   ```bash
   tb build -v
   ```

### Autocompletion Not Working

See the [Autocompletion Guide](autocompletion.md#troubleshooting) for solutions.

### Performance Issues

**Problem**: ToolBox is slow

**Solutions**:
1. Use local `.toolbox.yaml` instead of global config
2. Reduce number of contexts in config
3. Simplify detection logic in custom plugins
4. Check for slow network drives or NFS mounts

## Advanced Tips

### Aliases for Shorter Commands

Add shell aliases for even faster access:

```bash
# Add to ~/.bashrc or ~/.zshrc
alias tbb='tb build'
alias tbt='tb test'
alias tbr='tb run'
alias tbd='tb deploy'
```

### Environment-Specific Configs

Use different configs for different environments:

```bash
# Development
tb --config .toolbox.dev.yaml build

# Production
tb --config .toolbox.prod.yaml build
```

Or use environment variables:

```yaml
contexts:
  node:
    commands:
      build: "npm run build:${ENV:-dev}"
```

### Chaining Commands

Run multiple commands together:

```bash
# In shell
tb lint && tb test && tb build

# Or add as a custom command in .toolbox.yaml
contexts:
  node:
    commands:
      ci: "npm run lint && npm test && npm run build"
```

### Integration with Other Tools

Combine with other CLI tools:

```bash
# Use with watch for auto-rebuild
watchexec -e go tb build

# Use with parallel for multi-project builds
parallel 'cd {} && tb build' ::: frontend backend services

# Use with git hooks
# .git/hooks/pre-commit
#!/bin/bash
tb lint || exit 1
```

## Next Steps

- Set up [Shell Autocompletion](autocompletion.md) for faster workflows
- Learn about [Plugin Development](plugin-development.md) to create custom contexts
- Read the [Configuration Guide](configuration.md) for advanced configuration
- Check out [Example Configurations](../examples/example-config.yaml)

## Getting Help

- Use `tb help <command>` for command-specific help
- Check [GitHub Issues](https://github.com/bamf0/toolbox/issues) for known issues
- Join [GitHub Discussions](https://github.com/bamf0/toolbox/discussions) for community help
