# Configuration Guide

Advanced configuration options for ToolBox.

## Table of Contents

- [Configuration Overview](#configuration-overview)
- [Configuration File Format](#configuration-file-format)
- [Loading Priority](#loading-priority)
- [Context Configuration](#context-configuration)
- [Command Customization](#command-customization)
- [Security Considerations](#security-considerations)
- [Examples](#examples)

## Configuration Overview

ToolBox uses YAML configuration files to define contexts and commands. You can have multiple configuration files with a clear priority system.

### Configuration Locations

1. **Project-local**: `.toolbox.yaml` in current directory
2. **User-global**: `~/.toolbox/config.yaml` in home directory
3. **Custom**: Specified with `--config` flag
4. **Built-in**: Hard-coded defaults in the binary

### When to Use Each

- **Project-local**: Project-specific customizations
- **User-global**: Personal preferences across all projects
- **Custom**: Temporary overrides or environment-specific configs
- **Built-in**: Always available, good for standard workflows

## Configuration File Format

### Basic Structure

```yaml
contexts:
  <context-name>:
    commands:
      <command-name>: "<shell-command>"
    descriptions:
      <command-name>: "<description>"
```

### Example

```yaml
contexts:
  node:
    commands:
      build: "npm run build"
      test: "npm test"
      start: "npm start"
    descriptions:
      build: "Build the project"
      test: "Run tests"
      start: "Start the application"
```

## Loading Priority

Configuration files are loaded in this order (first match wins):

### 1. Command-Line Config

Highest priority:

```bash
tb --config my-config.yaml build
```

Use when:
- Testing new configurations
- Environment-specific deployments
- Temporary overrides

### 2. Local Project Config

Create `.toolbox.yaml` in project root:

```bash
cd ~/my-project
cat > .toolbox.yaml <<EOF
contexts:
  node:
    commands:
      build: "npm run build:prod"
EOF
```

Use when:
- Project has special build requirements
- Team wants consistent commands
- Sharing config via git

### 3. Global User Config

Create `~/.toolbox/config.yaml`:

```bash
mkdir -p ~/.toolbox
cat > ~/.toolbox/config.yaml <<EOF
contexts:
  python:
    commands:
      test: "pytest -v --cov"
EOF
```

Use when:
- Personal preferences for all projects
- Custom contexts you use everywhere
- Override defaults without changing projects

### 4. Built-in Defaults

Always available as fallback. See [config.go](../internal/config/config.go) for current defaults.

## Context Configuration

### Context Structure

```yaml
contexts:
  <context-name>:
    commands: { }       # Required: at least one command
    descriptions: { }   # Optional: command descriptions
```

### Context Naming

**Rules**:
- Lowercase with hyphens
- Max 50 characters
- Alphanumeric, dash, and underscore only

**Examples**:
```yaml
# Good
contexts:
  node:
  python-django:
  my_custom_context:
  
# Bad
contexts:
  Node:                    # Don't use uppercase
  python django:           # No spaces
  very-long-context-name-that-exceeds-fifty-characters:  # Too long
```

### Multiple Contexts

Define multiple contexts in one file:

```yaml
contexts:
  frontend:
    commands:
      dev: "npm run dev"
      build: "npm run build"
  
  backend:
    commands:
      dev: "go run ./cmd/server"
      build: "go build -o bin/server ./cmd/server"
  
  docker:
    commands:
      up: "docker-compose up -d"
      down: "docker-compose down"
```

### Extending Built-in Contexts

Add commands to existing contexts:

```yaml
contexts:
  # Add to built-in 'node' context
  node:
    commands:
      # Override default build
      build: "npm run build:production"
      # Add new commands
      deploy: "npm run deploy"
      docker: "docker build -t myapp ."
    descriptions:
      deploy: "Deploy to production"
      docker: "Build Docker image"
```

## Command Customization

### Simple Commands

```yaml
contexts:
  node:
    commands:
      build: "npm run build"
```

### Commands with Arguments

```yaml
contexts:
  python:
    commands:
      test: "pytest -v"
      test-cov: "pytest -v --cov --cov-report=html"
```

### Multi-Step Commands

Use shell operators:

```yaml
contexts:
  node:
    commands:
      # Sequential execution
      full-test: "npm run lint && npm test && npm run build"
      
      # Pipeline
      deploy: "npm run build | tar czf - dist/ | ssh user@server 'tar xzf -'"
      
      # Background process
      dev: "npm run dev &"
```

### Conditional Commands

Use shell conditionals:

```yaml
contexts:
  go:
    commands:
      build: "[ -d cmd ] && go build -o bin/app ./cmd/... || go build ./..."
```

### Commands with Environment Variables

```yaml
contexts:
  node:
    commands:
      build-dev: "NODE_ENV=development npm run build"
      build-prod: "NODE_ENV=production npm run build"
      
      # Use existing env vars
      deploy: "npm run deploy -- --env=${DEPLOY_ENV:-staging}"
```

### Script-Based Commands

Call external scripts:

```yaml
contexts:
  custom:
    commands:
      build: "bash scripts/build.sh"
      deploy: "./deploy.sh production"
      init: "python scripts/init.py"
```

### Command Descriptions

Always add descriptions for better UX:

```yaml
contexts:
  node:
    commands:
      build: "npm run build"
      test: "npm test"
      e2e: "npm run test:e2e"
    descriptions:
      build: "Build for production"
      test: "Run unit tests"
      e2e: "Run end-to-end tests with Playwright"
```

Descriptions appear in:
- Shell autocompletion
- Help output (`tb help <command>`)
- Command listings

## Security Considerations

### File Size Limits

Configuration files are limited to 1MB to prevent memory exhaustion.

### Command Length Limits

Individual commands are limited to 4096 characters.

### Path Validation

- Relative paths only for `--config` flag
- No directory traversal (`..`)
- Must be `.yaml` or `.yml` extension

### Command Injection

Be careful with user input in commands:

```yaml
# Dangerous - don't do this
contexts:
  custom:
    commands:
      # Never interpolate unvalidated user input
      deploy: "ssh user@${USER_PROVIDED_HOST} 'deploy.sh'"

# Safe alternatives
contexts:
  custom:
    commands:
      # Use fixed values
      deploy-staging: "ssh user@staging.example.com 'deploy.sh'"
      deploy-prod: "ssh user@prod.example.com 'deploy.sh'"
```

### Sensitive Data

Never commit secrets to configuration files:

```yaml
# Bad - secrets in config
contexts:
  deploy:
    commands:
      push: "docker login -u user -p MyS3cr3tP@ss && docker push myimage"

# Good - use environment variables or secret managers
contexts:
  deploy:
    commands:
      push: "docker login -u user -p $DOCKER_PASSWORD && docker push myimage"
```

## Examples

### Monorepo Configuration

Project structure:
```
monorepo/
├── .toolbox.yaml          # Root config
├── frontend/
│   ├── package.json
│   └── .toolbox.yaml      # Frontend overrides
└── backend/
    ├── go.mod
    └── .toolbox.yaml      # Backend overrides
```

Root `.toolbox.yaml`:
```yaml
contexts:
  monorepo:
    commands:
      build-all: "cd frontend && npm run build && cd ../backend && go build ./..."
      test-all: "cd frontend && npm test && cd ../backend && go test ./..."
```

Frontend `.toolbox.yaml`:
```yaml
contexts:
  node:
    commands:
      build: "npm run build -- --base=/app"
      dev: "npm run dev -- --port 3001"
```

### Environment-Specific Configuration

`.toolbox.dev.yaml`:
```yaml
contexts:
  node:
    commands:
      build: "NODE_ENV=development npm run build"
      start: "npm run dev"
```

`.toolbox.prod.yaml`:
```yaml
contexts:
  node:
    commands:
      build: "NODE_ENV=production npm run build"
      start: "npm start"
```

Usage:
```bash
tb --config .toolbox.dev.yaml build
tb --config .toolbox.prod.yaml build
```

### CI/CD Configuration

`.toolbox.ci.yaml`:
```yaml
contexts:
  node:
    commands:
      install: "npm ci"  # Use ci instead of install
      test: "npm test -- --ci --coverage"
      build: "npm run build -- --production"
      
  go:
    commands:
      test: "go test -v -race -coverprofile=coverage.out ./..."
      build: "CGO_ENABLED=0 go build -ldflags='-s -w' ./..."
```

### Docker Workflow

```yaml
contexts:
  docker:
    commands:
      build: "docker build -t myapp:latest ."
      run: "docker run --rm -p 8080:8080 myapp:latest"
      push: "docker push myapp:latest"
      compose-up: "docker-compose up -d"
      compose-down: "docker-compose down"
      compose-logs: "docker-compose logs -f"
    descriptions:
      build: "Build Docker image"
      run: "Run container locally"
      push: "Push to Docker registry"
      compose-up: "Start all services"
      compose-down: "Stop all services"
      compose-logs: "Follow service logs"
```

### Multi-Stage Build

```yaml
contexts:
  node:
    commands:
      install: "npm install"
      lint: "npm run lint"
      test: "npm test"
      build: "npm run build"
      package: "tar czf dist.tar.gz dist/"
      deploy: "scp dist.tar.gz user@server:/var/www/ && ssh user@server 'cd /var/www && tar xzf dist.tar.gz'"
      
      # Combined pipeline
      ci: "npm ci && npm run lint && npm test && npm run build"
      cd: "npm run build && tar czf dist.tar.gz dist/ && scp dist.tar.gz user@server:/var/www/"
    descriptions:
      ci: "Full CI pipeline (install, lint, test, build)"
      cd: "Build and deploy to server"
```

### Python Data Science

```yaml
contexts:
  python-ds:
    commands:
      notebook: "jupyter notebook"
      lab: "jupyter lab"
      
      test: "pytest tests/"
      test-cov: "pytest --cov=src --cov-report=html tests/"
      
      lint: "ruff check src/ && mypy src/"
      fmt: "black src/ tests/ && isort src/ tests/"
      
      train: "python scripts/train.py"
      evaluate: "python scripts/evaluate.py"
      
      docker-build: "docker build -t ml-model:latest ."
      docker-run: "docker run --gpus all -v $(pwd)/data:/data ml-model:latest"
    descriptions:
      notebook: "Start Jupyter Notebook"
      lab: "Start Jupyter Lab"
      train: "Train ML model"
      evaluate: "Evaluate model performance"
      docker-run: "Run model in Docker with GPU support"
```

### Rust Embedded

```yaml
contexts:
  rust-embedded:
    commands:
      build: "cargo build --release --target thumbv7em-none-eabihf"
      flash: "cargo flash --chip STM32F411RETx --release"
      debug: "cargo embed --chip STM32F411RETx"
      
      test-host: "cargo test"
      test-target: "cargo test --target thumbv7em-none-eabihf"
      
      size: "cargo size --release --target thumbv7em-none-eabihf -- -A"
      objdump: "cargo objdump --release --target thumbv7em-none-eabihf -- -d"
    descriptions:
      flash: "Flash firmware to target device"
      debug: "Start debugging session"
      size: "Show binary size breakdown"
      objdump: "Disassemble binary"
```

## Validation

ToolBox validates configuration files on load:

### Limits

- Max file size: 1MB
- Max contexts: 100
- Max commands per context: 50
- Max command length: 4096 characters
- Context name max length: 50 characters

### Error Messages

```bash
$ tb build
Error: invalid configuration: context "my-context" has too many commands (max: 50, got: 75)

$ tb --config huge.yaml build
Error: config file exceeds maximum size of 1048576 bytes (got 2000000 bytes)
```

## Best Practices

### 1. Keep It Simple

```yaml
# Good - clear and simple
contexts:
  node:
    commands:
      build: "npm run build"

# Avoid - overly complex
contexts:
  node:
    commands:
      build: "if [ $ENV = 'prod' ]; then npm run build:prod; elif [ $ENV = 'dev' ]; then npm run build:dev; else npm run build; fi"
```

### 2. Use Descriptions

```yaml
# Good - includes descriptions
contexts:
  go:
    commands:
      build: "go build ./..."
    descriptions:
      build: "Build all packages"

# Less helpful - no descriptions
contexts:
  go:
    commands:
      build: "go build ./..."
```

### 3. Version Control

```bash
# Commit project-specific config
git add .toolbox.yaml

# Don't commit user-specific config
echo "~/.toolbox/" >> ~/.gitignore
```

### 4. Document Custom Commands

Add comments to complex configurations:

```yaml
contexts:
  custom:
    commands:
      # Multi-stage deployment pipeline
      # 1. Build optimized production bundle
      # 2. Run smoke tests
      # 3. Deploy to staging
      # 4. Run integration tests
      deploy-staging: "npm run build && npm run test:smoke && ./deploy.sh staging && npm run test:integration"
```

### 5. Test Your Configuration

```bash
# Dry-run to verify commands
tb build --dry-run

# Test in isolation
tb --config .toolbox.yaml build --dry-run
```

## Next Steps

- Learn about [Plugin Development](plugin-development.md) for advanced customization
- Set up [Autocompletion](autocompletion.md) to use descriptions effectively
- Check [User Guide](user-guide.md) for everyday usage patterns
- See [Example Configurations](../examples/example-config.yaml)
