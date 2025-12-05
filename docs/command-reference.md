# Command Reference

Complete reference of all available commands and contexts in ToolBox.

## Table of Contents

- [Built-in Contexts](#built-in-contexts)
  - [Node.js](#nodejs)
  - [Go](#go)
  - [Python](#python)
  - [Rust](#rust)
  - [Make](#make)
- [Plugin Contexts](#plugin-contexts)
  - [Ubuntu Packaging](#ubuntu-packaging)
- [Meta Commands](#meta-commands)
- [Global Flags](#global-flags)

## Built-in Contexts

These contexts are built into ToolBox and available by default.

### Node.js

**Detected by**: `package.json`

**Context name**: `node`

#### Commands

| Command | Shell Command | Description |
|---------|--------------|-------------|
| `build` | `npm run build` | Build the project |
| `test` | `npm test` | Run tests |
| `start` | `npm start` | Start the application |
| `dev` | `npm run dev` | Start development server |
| `lint` | `npm run lint` | Run linter |
| `install` | `npm install` | Install dependencies |

#### Usage Examples

```bash
# Build the project
tb build

# Run tests
tb test

# Start development server
tb dev
```

---

### Go

**Detected by**: `go.mod`

**Context name**: `go`

#### Commands

| Command | Shell Command | Description |
|---------|--------------|-------------|
| `build` | `go build ./...` | Build all packages |
| `test` | `go test ./...` | Run all tests |
| `run` | `go run ./cmd/...` | Run main program |
| `install` | `go mod download` | Download dependencies |
| `lint` | `golangci-lint run` | Run golangci-lint |
| `fmt` | `go fmt ./...` | Format code |

#### Usage Examples

```bash
# Build all packages
tb build

# Run tests with verbose output
tb test -v

# Format code
tb fmt
```

---

### Python

**Detected by**: `pyproject.toml` or `requirements.txt`

**Context name**: `python`

#### Commands

| Command | Shell Command | Description |
|---------|--------------|-------------|
| `test` | `pytest` | Run tests with pytest |
| `lint` | `ruff check .` | Check code with ruff |
| `fmt` | `black .` | Format code with black |
| `install` | `pip install -r requirements.txt` | Install dependencies |
| `run` | `python main.py` | Run main.py |

#### Usage Examples

```bash
# Run tests
tb test

# Lint and format
tb lint && tb fmt

# Install dependencies
tb install
```

---

### Rust

**Detected by**: `Cargo.toml`

**Context name**: `rust`

#### Commands

| Command | Shell Command | Description |
|---------|--------------|-------------|
| `build` | `cargo build` | Build the project |
| `test` | `cargo test` | Run tests |
| `run` | `cargo run` | Run the binary |
| `install` | `cargo fetch` | Fetch dependencies |
| `lint` | `cargo clippy` | Run clippy linter |
| `fmt` | `cargo fmt` | Format code |

#### Usage Examples

```bash
# Build and run
tb build && tb run

# Run clippy
tb lint

# Format code
tb fmt
```

---

### Make

**Detected by**: `Makefile`

**Context name**: `make`

#### Commands

| Command | Shell Command | Description |
|---------|--------------|-------------|
| `build` | `make` | Build using Makefile |
| `test` | `make test` | Run tests |
| `clean` | `make clean` | Clean build artifacts |

#### Usage Examples

```bash
# Build project
tb build

# Run tests
tb test

# Clean
tb clean
```

---

## Plugin Contexts

These contexts are provided by plugins and may need to be enabled.

### Ubuntu Packaging

**Detected by**: `debian/control` or `debian/changelog`

**Context name**: `ubuntu-packaging`

**Plugin**: `ubuntu`

#### Branch Management Commands

| Command | Shell Command | Description |
|---------|--------------|-------------|
| `gbranch` | `bash ubuntu_helpers.sh gbranch` | Create/checkout git branch for bug fixes |
| `ppa-status` | `bash ubuntu_helpers.sh ppa-status` | Show PPA information from current branch |

##### gbranch Usage

```bash
# Create merge branch for bug LP#2133493
tb gbranch myproject 2133493 merge

# Create SRU branch with description
tb gbranch myproject 2127080 sru escape-equals

# Create bug fix branch
tb gbranch myproject 2100000 bug my-fix
```

**Format**: `tb gbranch <project> <bug-id> [merge|sru|bug] [description]`

Branch naming:
- Merge: `merge-lp<bug>`
- SRU: `sru-lp<bug>-<release>`
- Bug: `bug-lp<bug>-<release>`

#### Changelog Commands

| Command | Shell Command | Description |
|---------|--------------|-------------|
| `dch-auto` | `bash ubuntu_helpers.sh dch-auto` | Auto-update changelog with version suffix |
| `dch` | `dch -i` | Add new changelog entry manually |
| `dch-release` | `dch -r` | Mark changelog entry as released |
| `changelog` | `dpkg-parsechangelog` | Display full changelog |
| `version` | `dpkg-parsechangelog -S Version` | Show current package version |

##### dch-auto Usage

Automatically updates `debian/changelog` based on current git branch:

```bash
# On branch merge-lp2133493
tb dch-auto
# Adds entry with version suffix ~noble1
```

#### Build Commands

| Command | Shell Command | Description |
|---------|--------------|-------------|
| `ubuild` | `bash ubuntu_helpers.sh ubuild` | Complete build and upload workflow |
| `sb-auto` | `bash ubuntu_helpers.sh sb-auto` | Build source package with sbuild |
| `dput-auto` | `bash ubuntu_helpers.sh dput-auto` | Upload to PPA inferred from branch |
| `build` | `dpkg-buildpackage -us -uc` | Build binary package |
| `build-source` | `dpkg-buildpackage -S -us -uc` | Build source package only |

##### Build Workflow

```bash
# Full workflow: build and upload
tb ubuild

# Or step-by-step:
tb dch-auto       # Update changelog
tb sb-auto        # Build with sbuild
tb dput-auto      # Upload to PPA
```

#### Linting Commands

| Command | Shell Command | Description |
|---------|--------------|-------------|
| `lint` | `lintian` | Run lintian on built packages |
| `lint-source` | `lintian --pedantic *.dsc` | Run lintian on source package |
| `lint-changes` | `lintian --pedantic *.changes` | Run lintian on .changes file |

#### Clean Commands

| Command | Shell Command | Description |
|---------|--------------|-------------|
| `clean` | `debian/rules clean` | Clean build artifacts |
| `distclean` | `fakeroot debian/rules clean` | Deep clean using fakeroot |

#### Complete Ubuntu Workflow Example

```bash
# 1. Create branch for bug fix
cd my-package
tb gbranch mypackage 2133493 merge

# 2. Make your changes
vim src/file.c

# 3. Update changelog
tb dch-auto
# or manually
tb dch

# 4. Build and test locally
tb build
tb lint

# 5. Build and upload to PPA
tb ubuild

# 6. Check PPA status
tb ppa-status
```

---

## Meta Commands

These commands control ToolBox itself.

### help

Show help for commands.

```bash
# General help
tb --help
tb -h

# Help for a specific command (uses current context)
tb help build

# Help in a specific context
tb help --context python test
```

### completion

Generate shell completion scripts.

```bash
# Generate for Bash
tb completion bash

# Generate for Zsh
tb completion zsh

# Generate for Fish
tb completion fish

# Generate for PowerShell
tb completion powershell
```

See [Autocompletion Guide](autocompletion.md) for installation instructions.

### plugin

Manage and view plugins.

```bash
# List all plugins
tb plugin list

# Show plugin information
tb plugin info ubuntu

# List all contexts (including from plugins)
tb plugin contexts
```

---

## Global Flags

These flags work with any command.

### --context

Force a specific context.

```bash
# Use Node.js context even if not detected
tb --context node build

# Use Python context
tb --context python test
```

### --config

Use a custom configuration file.

```bash
# Use project-specific config
tb --config .toolbox.dev.yaml build

# Use production config
tb --config ~/.toolbox/prod.yaml deploy
```

### --dry-run

Preview command without executing.

```bash
# See what would run
tb build --dry-run
# Output:
# Context: node
# Command: npm run build

# Combine with context override
tb --context go build --dry-run
```

### --verbose / -v

Show detailed output.

```bash
# Verbose mode
tb -v build
tb --verbose test
```

### --help / -h

Show help.

```bash
tb --help
tb -h
tb build --help
```

### --version

Show version information.

```bash
tb --version
```

---

## Command Patterns

### Common Command Combinations

```bash
# Lint, test, and build
tb lint && tb test && tb build

# Clean and rebuild
tb clean && tb build

# Install dependencies and run
tb install && tb run

# Format, lint, and test
tb fmt && tb lint && tb test
```

### Using with Other Tools

```bash
# Watch and rebuild
watchexec -e go tb build

# Parallel builds
parallel 'cd {} && tb build' ::: dir1 dir2 dir3

# With environment variables
ENV=production tb build

# Pipe output
tb test | tee test-results.log
```

### Script Integration

```bash
#!/bin/bash
# build.sh

set -e  # Exit on error

tb install
tb lint
tb test
tb build

echo "Build complete!"
```

---

## Context Detection Priority

When multiple marker files exist, ToolBox uses this priority:

1. **Plugin contexts** (checked first)
   - Ubuntu packaging (`debian/control`, `debian/changelog`)
   
2. **Built-in contexts** (checked in order)
   - Node.js (`package.json`)
   - Go (`go.mod`)
   - Python (`pyproject.toml`, `requirements.txt`)
   - Rust (`Cargo.toml`)
   - Make (`Makefile`)

To override detection, use `--context` flag:

```bash
tb --context python build
```

---

## Custom Commands

You can add custom commands via configuration:

```yaml
# .toolbox.yaml
contexts:
  node:
    commands:
      deploy: "npm run deploy"
      docker: "docker build -t myapp ."
    descriptions:
      deploy: "Deploy to production"
      docker: "Build Docker image"
```

Then use them like built-in commands:

```bash
tb deploy
tb docker
```

See [Configuration Guide](configuration.md) for more details.

---

## Next Steps

- **Getting Started**: Read the [User Guide](user-guide.md)
- **Autocompletion**: Set up [Shell Autocompletion](autocompletion.md)
- **Customization**: Learn about [Configuration](configuration.md)
- **Extend ToolBox**: Try [Plugin Development](plugin-development.md)
