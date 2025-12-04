# Future Extensions and Architecture Notes

## Current Architecture Overview

The ToolBox MVP is built with extensibility in mind. Here's how the components work together:

```
User Input (tb build)
    ↓
CLI Layer (Cobra)
    ↓
Context Detector → Scans filesystem for markers
    ↓
Config Loader → Loads YAML config (with defaults)
    ↓
Command Registry → Maps context + command → full command
    ↓
Executor → Runs the expanded command in shell
```

## Design Decisions

### 1. Dynamic Command Handling

Instead of defining each command (build, test, etc.) as separate Cobra commands, we use a **dynamic handler** that intercepts unknown commands. This provides:

- **Unlimited commands**: Add new commands via config without code changes
- **Context-aware behavior**: Same command name works differently per context
- **Simplicity**: Single code path for all dynamic commands

### 2. Configuration Merging

The config system merges user configs with built-in defaults:

```
User Config + Default Config → Final Config
```

This means:
- Users only override what they need
- New default contexts are automatically available
- Zero-config experience for common setups

### 3. Context Detection Priority

Contexts are detected in a specific order (Node.js, Go, Python, Rust, etc.):

- **Why?** Some projects have multiple marker files (e.g., both Makefile and package.json)
- **Solution:** Priority order ensures consistent detection
- **Extensible:** Users can override with `--context` flag

### 4. Modular Package Structure

Each component is isolated:

- `cli/`: Command-line interface only
- `config/`: Configuration logic only
- `context/`: Detection logic only
- `registry/`: Command lookup only

**Benefits:**
- Easy to test each component independently
- Easy to replace/extend individual pieces
- Clear separation of concerns

## Future Enhancement Ideas

### 1. Plugin System

**Goal:** Allow third-party plugins to add custom contexts and commands

**Implementation approach:**
```go
// Plugin interface
type Plugin interface {
    Name() string
    Contexts() map[string]ContextConfig
    Detect(dir string) (string, bool)
}

// Plugin loader
func LoadPlugins(dir string) ([]Plugin, error)
```

**Config:**
```yaml
plugins:
  - ~/.toolbox/plugins/docker-plugin.so
  - ~/.toolbox/plugins/k8s-plugin.so
```

### 2. Interactive Mode

**Goal:** Show available commands and let users select

**Example:**
```bash
$ tb --interactive
? Select a command: 
  > build
    test
    lint
    deploy
    
? Command: build
Running: npm run build
```

**Implementation:** Use a library like `survey` or `bubbletea`

### 3. Command Templates with Variables

**Goal:** Support parameterized commands

**Config:**
```yaml
contexts:
  node:
    commands:
      deploy: "npm run deploy -- --env={{.env}} --region={{.region}}"
```

**Usage:**
```bash
tb deploy --env=prod --region=us-east-1
# Expands to: npm run deploy -- --env=prod --region=us-east-1
```

**Implementation:** Use Go's `text/template`

### 4. Command History and Favorites

**Goal:** Track frequently used commands and provide quick access

**Features:**
- Show recent commands with `tb history`
- Mark favorites with `tb favorite add deploy`
- Quick execute with `tb !3` (run command #3 from history)

**Storage:** SQLite database or JSON file in `~/.toolbox/history.db`

### 5. Environment-Specific Overrides

**Goal:** Different commands for dev/staging/prod environments

**Config:**
```yaml
contexts:
  node:
    commands:
      deploy: "npm run deploy"
    environments:
      prod:
        deploy: "npm run deploy:prod"
      staging:
        deploy: "npm run deploy:staging"
```

**Usage:**
```bash
tb deploy --env=prod
```

### 6. Shell Completion

**Goal:** Tab completion for commands and flags

**Implementation:** Cobra has built-in completion support

```bash
# Generate completion
tb completion bash > /etc/bash_completion.d/tb

# Now works:
tb bui<TAB>  → tb build
```

### 7. Multi-Context Projects

**Goal:** Support projects with multiple contexts (e.g., monorepo with Node + Python)

**Config:**
```yaml
contexts:
  monorepo:
    commands:
      build: |
        npm run build && 
        pip install -r requirements.txt
      test: |
        npm test &&
        pytest
```

### 8. Command Hooks (Pre/Post)

**Goal:** Run commands before/after main command

**Config:**
```yaml
contexts:
  node:
    commands:
      test:
        pre: "npm run lint"
        command: "npm test"
        post: "npm run coverage"
```

### 9. Remote Context Detection

**Goal:** Detect context from remote repositories before cloning

**Usage:**
```bash
tb detect github.com/user/repo
# Output: This is a Node.js project
```

**Implementation:** Use GitHub API to check for marker files

### 10. CI/CD Integration

**Goal:** Generate CI/CD configs based on detected context

**Usage:**
```bash
tb generate-ci
# Creates .github/workflows/build.yml based on detected context
```

## Extension Points in Current Code

### Adding New Contexts

**No code changes needed!** Just update config:

```yaml
contexts:
  dart:
    commands:
      build: "dart compile exe bin/main.dart"
      test: "dart test"
```

### Adding Context Markers

**Code change needed** (but simple):

```go
// In internal/context/detector.go
func NewDetector() *Detector {
    return &Detector{
        markers: map[string][]string{
            // ... existing markers ...
            "dart": {"pubspec.yaml"},
            "swift": {"Package.swift"},
        },
    }
}
```

### Custom Command Execution

**Current:** Commands run in shell (sh/bash/cmd)

**Future:** Support for different executors:

```go
type Executor interface {
    Execute(command string) error
}

type ShellExecutor struct{}
type DockerExecutor struct{}
type SSHExecutor struct{}
```

## Testing Strategy

### Unit Tests

Each package should have tests:

```
internal/config/config_test.go
internal/context/detector_test.go
internal/registry/registry_test.go
```

### Integration Tests

Test the full flow:

```go
func TestBuildCommand(t *testing.T) {
    // Create temp project with go.mod
    // Run tb build --dry-run
    // Verify output is "go build"
}
```

### Table-Driven Tests

For context detection:

```go
tests := []struct {
    files    []string
    expected string
}{
    {[]string{"package.json"}, "node"},
    {[]string{"go.mod"}, "go"},
    {[]string{"Cargo.toml"}, "rust"},
}
```

## Performance Considerations

### Current Performance

- Context detection: < 1ms (filesystem scans)
- Config loading: < 5ms (YAML parsing)
- Command execution: Depends on underlying command

### Optimization Opportunities

1. **Cache context detection results** per directory
2. **Lazy load config** only when needed
3. **Parallel detection** for multiple markers

## Security Considerations

### Current

- Commands run in user's shell (same permissions as user)
- No validation of command strings
- Config files can contain arbitrary commands

### Future Improvements

1. **Command validation**: Warn on dangerous commands (rm -rf, etc.)
2. **Sandboxing**: Option to run commands in containers
3. **Config signing**: Verify config files haven't been tampered with

## Contribution Guidelines

### Adding a New Context

1. Add default commands to `internal/config/config.go`
2. Add marker files to `internal/context/detector.go`
3. Update `examples/example-config.yaml`
4. Update `README.md` with example usage

### Adding a New Feature

1. Create issue describing the feature
2. Implement in appropriate package
3. Add tests
4. Update documentation
5. Submit PR

## Release Process

### Versioning

Follow semantic versioning:
- MAJOR: Breaking API changes
- MINOR: New features (backward compatible)
- PATCH: Bug fixes

### Distribution

Build binaries for multiple platforms:

```bash
# Use goreleaser or custom script
GOOS=linux GOARCH=amd64 go build -o dist/tb-linux-amd64
GOOS=darwin GOARCH=amd64 go build -o dist/tb-darwin-amd64
GOOS=windows GOARCH=amd64 go build -o dist/tb-windows-amd64.exe
```

## Community & Support

### Documentation

- README.md: Quick start and overview
- BUILD.md: Build and installation instructions
- STRUCTURE.md: Architecture and design
- EXTENSIONS.md: This file - future plans

### Issue Labels

- `enhancement`: New feature requests
- `bug`: Something isn't working
- `documentation`: Docs improvements
- `good-first-issue`: Good for newcomers

---

**The MVP provides a solid foundation. The architecture supports all these extensions without major refactoring.**
