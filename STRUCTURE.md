# ToolBox Project Structure

```
toolbox/
├── cmd/
│   └── tb/
│       └── main.go              # CLI entry point
├── internal/
│   ├── cli/
│   │   └── root.go              # Cobra root command and dynamic handler
│   ├── config/
│   │   └── config.go            # Configuration loading and parsing
│   ├── context/
│   │   └── detector.go          # Project context detection
│   └── registry/
│       └── registry.go          # Command registry and lookup
├── examples/
│   └── example-config.yaml      # Example configuration file
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

## Directory Breakdown

### `/cmd/tb`
Contains the main application entry point. Following Go conventions, the `main.go` file here is minimal and delegates to the internal packages.

### `/internal`
Private application code that cannot be imported by other projects. This enforces encapsulation.

- **`cli/`**: Cobra-based CLI command structure
  - Handles command-line argument parsing
  - Implements dynamic command routing
  - Manages flags (--dry-run, --verbose, --context)

- **`config/`**: Configuration management
  - Loads YAML configuration files
  - Merges user config with defaults
  - Provides sensible defaults for common contexts

- **`context/`**: Project context detection
  - Scans filesystem for marker files (package.json, go.mod, etc.)
  - Returns detected context type
  - Supports parent directory traversal

- **`registry/`**: Command registry
  - Maps commands to implementations per context
  - Provides command lookup functionality
  - Supports listing available commands

### `/examples`
Sample configuration files for reference.

## Design Philosophy

1. **Separation of Concerns**: Each package has a single, well-defined responsibility
2. **Testability**: Pure functions and minimal dependencies make testing straightforward
3. **Extensibility**: Adding new contexts or commands only requires config changes
4. **Idiomatic Go**: Follows standard project layout and naming conventions
