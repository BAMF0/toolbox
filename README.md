# ToolBox (tb)

(Built as a test project with GitHub Copilot)

**Context-aware command aliasing for developers**

ToolBox is a CLI tool that provides intelligent command shortcuts based on your project context. Define simple commands like `tb build`, `tb test`, or `tb deploy` that automatically expand to the correct commands for your current project type.

```bash
# Instead of remembering these:
npm run build       # Node.js
go build            # Go
cargo build         # Rust
python -m build     # Python

# Just use:
tb build           # Works in any project!
```

---

## Quick Demo

```bash
# Build the tool
$ go build -o tb ./cmd/tb

# Use it in any project
$ cd ~/my-node-project
$ tb build --dry-run
Context: node
Command: npm run build

$ cd ~/my-go-api  
$ tb test --dry-run
Context: go
Command: go test ./...

# It just works!
```

---

## Features

- **Auto-Detection**: Identifies project type from marker files (package.json, go.mod, etc.)
- **Smart Aliasing**: Same command works across different project types
- **Configurable**: YAML-based config with sensible defaults
- **Extensible**: Add custom contexts without touching code
- **Cross-platform**: Linux, macOS, and Windows
- **Zero Config**: Works out-of-the-box for common projects
- **Built-in Help**: Comprehensive help for all commands with `tb help <command>`
- **Shell Autocompletion**: Tab completion with command descriptions for bash, zsh, fish

---

## Installation

### From Source

```bash
# Clone and build
git clone https://github.com/bamf0/toolbox.git
cd toolbox
go build -o tb ./cmd/tb

# Install to PATH
sudo cp tb /usr/local/bin/
# or
go install ./cmd/tb
```

### Pre-built Binaries (Coming Soon)

Download from [Releases](https://github.com/bamf0/toolbox/releases)

---

## Usage

### Basic Commands

```bash
# Run in any supported project directory
tb build              # Runs the appropriate build command
tb test               # Runs tests
tb start              # Starts the project
tb lint               # Runs linter

# Preview without executing
tb build --dry-run    # Shows what would run

# Verbose output
tb test -v            # Shows context detection

# Force a specific context
tb build --context python

# Get help for any command
tb help build              # Shows description and what it runs

# Install shell autocompletion (bash example)
tb completion bash | sudo tee /etc/bash_completion.d/tb
source ~/.bashrc
# Now use Tab to see commands with descriptions!
```

### Getting Help

```bash
# Show help for a command (uses current context)
tb help build

# Show help in a specific context
tb help --context ubuntu-packaging gbranch

# See available plugins
tb plugin list

# Get plugin details
tb plugin info ubuntu
```

### Supported Contexts (Default)

| Context | Auto-detected by | Available Commands |
|---------|------------------|-------------------|
| **Node.js** | package.json | build, test, start, dev, lint, install |
| **Go** | go.mod | build, test, run, fmt, lint, install |
| **Python** | pyproject.toml, requirements.txt | test, lint, fmt, install, run |
| **Rust** | Cargo.toml | build, test, run, lint, fmt, install |
| **Make** | Makefile | build, test, clean |
| **Ruby** | Gemfile | (extensible via config) |
| **Java** | pom.xml, build.gradle | (extensible via config) |
| **PHP** | composer.json | (extensible via config) |

---

## Configuration

### Quick Customization

Create `.toolbox.yaml` in your project root:

```yaml
contexts:
  node:
    commands:
      build: "npm run build:prod"
      test: "npm test -- --coverage"
      deploy: "npm run deploy"
  
  custom:
    commands:
      up: "docker-compose up -d"
      down: "docker-compose down"
      logs: "docker-compose logs -f"
```

### Global Configuration

Create `~/.toolbox/config.yaml` for system-wide settings:

```yaml
contexts:
  python:
    commands:
      test: "pytest -v --cov"
      lint: "ruff check . && mypy ."
      fmt: "black . && isort ."
```

### Configuration Priority

1. File specified with `--config` flag
2. `.toolbox.yaml` in current directory  
3. `~/.toolbox/config.yaml` in home directory
4. Built-in defaults

> **Tip**: For Go projects with a specific structure (like `cmd/myapp/`), customize in `.toolbox.yaml`:
> ```yaml
> contexts:
>   go:
>     commands:
>       build: "go build -o bin/myapp ./cmd/myapp"
>       run: "go run ./cmd/myapp"
> ```

---

## Architecture

ToolBox follows a clean, modular architecture:

```
User Command (tb build)
         ↓
    CLI Layer (Cobra)
         ↓
  Context Detector → Scans for package.json, go.mod, etc.
         ↓
   Config Loader → Merges YAML with defaults
         ↓
  Command Registry → Looks up command for context
         ↓
     Executor → Runs expanded shell command
```

### Project Structure

```
toolbox/
├── cmd/tb/              # CLI entry point
├── internal/
│   ├── cli/             # Cobra commands & handlers
│   ├── config/          # YAML config loading
│   ├── context/         # Project detection
│   ├── plugin/          # Plugin system
│   └── registry/        # Command lookup
├── examples/            # Example configs
└── scripts/             # Helper scripts
```

### Design Principles

1. **Modular**: Clear separation of concerns
2. **Simple**: Focused on core functionality  
3. **Extensible**: Easy to add contexts via config or plugins
4. **Idiomatic**: Follows Go best practices

---

## Plugin System

ToolBox supports a flexible plugin system for extending functionality:

```bash
# List available plugins
tb plugin list

# Get plugin information
tb plugin info ubuntu

# List all contexts (including from plugins)
tb plugin contexts
```

### Available Plugins

- **Ubuntu Packaging**: Smart PPA-based Ubuntu packaging workflow
  - Auto-detects Debian packaging directories
  - Manages git branches based on bug IDs
  - Automates changelog, builds, and uploads
  - See `tb plugin info ubuntu` for details

---

## Future Enhancements

Planned features include:

- **Interactive Mode**: Command selection UI
- **Templates**: Parameterized commands with variables
- **History**: Track and reuse frequent commands
- **Environments**: Different commands for dev/staging/prod
- **Hooks**: Pre/post command execution
- **CI/CD Integration**: Generate pipeline configs

---

## Documentation

Full documentation is available in the [docs/](docs/) directory:

- **[User Guide](docs/user-guide.md)** - Complete guide to using ToolBox
- **[Shell Autocompletion](docs/autocompletion.md)** - Setting up tab completion
- **[Plugin Development](docs/plugin-development.md)** - Creating custom plugins
- **[Configuration Guide](docs/configuration.md)** - Advanced configuration options
- **[API Reference](docs/api-reference.md)** - Plugin API documentation
- **[Command Reference](docs/command-reference.md)** - All available commands

### Quick Help

```bash
# Show general help
tb --help

# Show help for a specific command
tb help build

# Show help in a specific context
tb help --context ubuntu-packaging gbranch

# Generate shell completion
tb completion bash > /etc/bash_completion.d/tb
```

---

## Contributing

Contributions are welcome! Here's how:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Adding a New Context

Just update the config - no code changes needed:

```yaml
contexts:
  elixir:
    commands:
      build: "mix compile"
      test: "mix test"
      run: "mix run"
```

---

## License

[MIT License](LICENSE) - feel free to use in your projects!

---

## Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) - excellent CLI framework
- Inspired by Make, Task, and Just - but context-aware
- Thanks to the Go community for great tooling

---

## Support

- **Issues**: [GitHub Issues](https://github.com/bamf0/toolbox/issues)
- **Discussions**: [GitHub Discussions](https://github.com/bamf0/toolbox/discussions)  

---

## Quick Links

- [Installation](#-installation)
- [Usage](#-usage)
- [Configuration](#️-configuration)
- [Documentation](#-documentation)
- [Contributing](#-contributing)

