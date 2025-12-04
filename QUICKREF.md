# ToolBox Quick Reference

## Command Syntax

```bash
tb <command> [args...] [flags]
```

## Flags

| Flag | Shorthand | Description |
|------|-----------|-------------|
| `--config <file>` | | Use specific config file |
| `--context <name>` | | Force specific context (node, go, python, etc.) |
| `--dry-run` | | Print command without executing |
| `--verbose` | `-v` | Show detailed output (context detection, etc.) |
| `--help` | `-h` | Show help message |

## Supported Contexts (Default)

| Context | Marker Files | Example Commands |
|---------|--------------|------------------|
| **node** | package.json, package-lock.json, yarn.lock | build, test, start, dev, lint, install |
| **go** | go.mod, go.sum | build, test, run, install, lint, fmt |
| **python** | pyproject.toml, setup.py, requirements.txt | test, lint, fmt, install, run |
| **rust** | Cargo.toml, Cargo.lock | build, test, run, install, lint, fmt |
| **make** | Makefile | build, test, clean |
| **ruby** | Gemfile, Gemfile.lock | - |
| **java** | pom.xml, build.gradle | - |
| **php** | composer.json | - |

## Common Usage Patterns

### Auto-Detection (Recommended)

```bash
# In any project directory
tb build        # Automatically uses correct build command
tb test         # Automatically uses correct test command
```

### Dry-Run (Preview)

```bash
tb build --dry-run       # See what would run without executing
tb test -v --dry-run     # Verbose + dry-run
```

### Force Context

```bash
tb build --context python    # Use Python context even in Go project
tb test --context rust       # Use Rust context
```

### Pass Arguments

```bash
tb test -- --coverage        # Pass --coverage to underlying test command
tb run -- arg1 arg2          # Pass arguments to run command
```

## Configuration Examples

### Project Config (.toolbox.yaml)

```yaml
contexts:
  node:
    commands:
      build: "npm run build:prod"
      deploy: "npm run deploy"
  
  custom:
    commands:
      start: "docker-compose up"
      stop: "docker-compose down"
```

### Global Config (~/.toolbox/config.yaml)

```yaml
contexts:
  python:
    commands:
      test: "pytest -v --cov"
      lint: "ruff check . && mypy ."
```

## Real-World Examples

### Node.js Project

```bash
cd ~/projects/my-webapp
tb install              # → npm install
tb dev                  # → npm run dev
tb build                # → npm run build
tb test                 # → npm test
```

### Go Project

```bash
cd ~/projects/my-api
tb build                # → go build
tb test                 # → go test ./...
tb fmt                  # → go fmt ./...
tb run                  # → go run .
```

### Python Project

```bash
cd ~/projects/ml-model
tb install              # → pip install -r requirements.txt
tb test                 # → pytest
tb lint                 # → ruff check .
```

### Multi-Tool Project

```yaml
# .toolbox.yaml
contexts:
  fullstack:
    commands:
      dev: "docker-compose up"
      test: "npm test && pytest"
      deploy: "./deploy.sh prod"
```

```bash
tb dev                  # → docker-compose up
tb test                 # → npm test && pytest
tb deploy               # → ./deploy.sh prod
```

## Troubleshooting

| Problem | Solution |
|---------|----------|
| "No recognized project context found" | Use `--context` flag or create `.toolbox.yaml` |
| "Command not found in context" | Add command to `.toolbox.yaml` or check spelling |
| Command runs wrong thing | Check with `--dry-run -v` to see detection |
| Need to run in different dir | `cd` to project root or use custom config |

## Tips & Tricks

### Create Project Aliases

```bash
# In .toolbox.yaml
contexts:
  myapp:
    commands:
      up: "docker-compose up -d && npm run dev"
      down: "docker-compose down"
      logs: "docker-compose logs -f"
      shell: "docker-compose exec app bash"
```

### Chain Commands

```yaml
commands:
  ci: "npm run lint && npm test && npm run build"
```

### Environment Variables

```yaml
commands:
  deploy-prod: "ENV=production ./deploy.sh"
  deploy-staging: "ENV=staging ./deploy.sh"
```

### Complex Commands

```yaml
commands:
  backup: |
    timestamp=$(date +%Y%m%d_%H%M%S)
    tar -czf backup_${timestamp}.tar.gz ./data
    echo "Backup created: backup_${timestamp}.tar.gz"
```

## Quick Setup

```bash
# 1. Build and install
go build -o tb ./cmd/tb
sudo cp tb /usr/local/bin/

# 2. Test in current project
tb build --dry-run

# 3. Create project config
cat > .toolbox.yaml << 'EOF'
contexts:
  go:
    commands:
      build: "go build -o bin/myapp"
      deploy: "scp bin/myapp server:/opt/myapp/"
EOF

# 4. Use it
tb build
tb deploy
```

## Getting Help

```bash
tb --help               # Show main help
tb --version            # Show version (future)
```

## Configuration File Locations (Priority Order)

1. File specified with `--config` flag
2. `.toolbox.yaml` (current directory)
3. `~/.toolbox/config.yaml` (global)
4. Built-in defaults

---

**Pro Tip:** Start with defaults, then customize only what you need in `.toolbox.yaml`
