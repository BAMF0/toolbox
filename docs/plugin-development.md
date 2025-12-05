# Plugin Development Guide

ToolBox supports a flexible plugin system that allows you to extend functionality with custom contexts and commands.

## Table of Contents

- [Plugin Architecture](#plugin-architecture)
- [Creating a Plugin](#creating-a-plugin)
- [Plugin Interface](#plugin-interface)
- [Example Plugins](#example-plugins)
- [Plugin Registration](#plugin-registration)
- [Best Practices](#best-practices)
- [Testing Plugins](#testing-plugins)
- [Security Considerations](#security-considerations)

## Plugin Architecture

ToolBox uses a **registry-based plugin system** for security and portability. Unlike dynamic plugin loading (`.so`/`.dll` files), plugins are Go packages compiled directly into the ToolBox binary.

### Why Registry-Based?

- **Security**: No arbitrary code execution from external files
- **Portability**: Single binary, no external dependencies
- **Performance**: No runtime plugin loading overhead
- **Type Safety**: Full Go compile-time type checking

### Plugin Lifecycle

```
1. Plugin implementation (Go code)
       ↓
2. Register in PluginManager (compile time)
       ↓
3. Context detection at runtime
       ↓
4. Commands available via tb CLI
```

## Creating a Plugin

### Step 1: Define Your Plugin Struct

Create a new file in `internal/plugin/`:

```go
package plugin

import (
    "os"
    "path/filepath"
    "github.com/bamf0/toolbox/internal/config"
)

// MyPlugin provides custom functionality
type MyPlugin struct {
    name    string
    version string
}

// NewMyPlugin creates a new instance
func NewMyPlugin() *MyPlugin {
    return &MyPlugin{
        name:    "my-plugin",
        version: "1.0.0",
    }
}
```

### Step 2: Implement the Plugin Interface

Every plugin must implement the `Plugin` interface:

```go
type Plugin interface {
    Name() string
    Version() string
    Contexts() map[string]config.ContextConfig
    Detect(dir string) (string, bool)
    Validate() error
}
```

#### Name() and Version()

```go
func (p *MyPlugin) Name() string {
    return p.name
}

func (p *MyPlugin) Version() string {
    return p.version
}
```

#### Contexts()

Define the contexts and commands your plugin provides:

```go
func (p *MyPlugin) Contexts() map[string]config.ContextConfig {
    return map[string]config.ContextConfig{
        "my-context": {
            Commands: map[string]string{
                "build":   "my-build-command",
                "test":    "my-test-command",
                "deploy":  "my-deploy-command",
            },
            Descriptions: map[string]string{
                "build":   "Build the project",
                "test":    "Run tests",
                "deploy":  "Deploy to production",
            },
        },
    }
}
```

#### Detect()

Implement context detection logic:

```go
func (p *MyPlugin) Detect(dir string) (string, bool) {
    // Check for marker files
    markerFile := filepath.Join(dir, "my-config.yml")
    if _, err := os.Stat(markerFile); err == nil {
        return "my-context", true
    }
    
    // Check for directory structure
    if _, err := os.Stat(filepath.Join(dir, "src", "main")); err == nil {
        return "my-context", true
    }
    
    return "", false
}
```

#### Validate()

Perform self-validation:

```go
func (p *MyPlugin) Validate() error {
    if p.name == "" {
        return fmt.Errorf("plugin name cannot be empty")
    }
    if p.version == "" {
        return fmt.Errorf("plugin version cannot be empty")
    }
    
    contexts := p.Contexts()
    if len(contexts) == 0 {
        return fmt.Errorf("plugin must provide at least one context")
    }
    
    for ctxName, ctxConfig := range contexts {
        if len(ctxConfig.Commands) == 0 {
            return fmt.Errorf("context %q has no commands", ctxName)
        }
    }
    
    return nil
}
```

### Step 3: Register Your Plugin

In `internal/cli/root.go` or where plugins are initialized:

```go
func init() {
    pm := plugin.NewPluginManager("")
    
    // Register built-in plugins
    pm.RegisterPlugin(plugin.NewUbuntuPlugin())
    pm.RegisterPlugin(plugin.NewMyPlugin())  // Add your plugin
    
    pluginManager = pm
}
```

## Plugin Interface

### Complete Interface Definition

```go
type Plugin interface {
    // Name returns unique plugin identifier
    // Convention: lowercase with hyphens (e.g., "my-plugin")
    Name() string
    
    // Version returns semantic version (e.g., "1.0.0")
    Version() string
    
    // Contexts returns all contexts this plugin provides
    // Key: context name, Value: context configuration
    Contexts() map[string]config.ContextConfig
    
    // Detect checks if plugin's context applies to directory
    // Returns: (context-name, true) if detected, ("", false) otherwise
    Detect(dir string) (string, bool)
    
    // Validate performs plugin self-validation
    // Called during registration to catch issues early
    Validate() error
}
```

### ContextConfig Structure

```go
type ContextConfig struct {
    Commands     map[string]string  // command-name → shell-command
    Descriptions map[string]string  // command-name → description
}
```

## Example Plugins

### Simple Static Plugin

A plugin with fixed commands:

```go
type StaticPlugin struct {
    name    string
    version string
}

func NewStaticPlugin() *StaticPlugin {
    return &StaticPlugin{
        name:    "static-example",
        version: "1.0.0",
    }
}

func (p *StaticPlugin) Contexts() map[string]config.ContextConfig {
    return map[string]config.ContextConfig{
        "docker": {
            Commands: map[string]string{
                "up":    "docker-compose up -d",
                "down":  "docker-compose down",
                "logs":  "docker-compose logs -f",
                "build": "docker-compose build",
            },
            Descriptions: map[string]string{
                "up":    "Start containers in background",
                "down":  "Stop and remove containers",
                "logs":  "Follow container logs",
                "build": "Build or rebuild services",
            },
        },
    }
}

func (p *StaticPlugin) Detect(dir string) (string, bool) {
    if _, err := os.Stat(filepath.Join(dir, "docker-compose.yml")); err == nil {
        return "docker", true
    }
    return "", false
}

func (p *StaticPlugin) Name() string    { return p.name }
func (p *StaticPlugin) Version() string { return p.version }
func (p *StaticPlugin) Validate() error { return nil }
```

### Dynamic Plugin with Helper Scripts

A plugin that uses external scripts (like the Ubuntu plugin):

```go
type ScriptedPlugin struct {
    name    string
    version string
}

func NewScriptedPlugin() *ScriptedPlugin {
    return &ScriptedPlugin{
        name:    "scripted-example",
        version: "1.0.0",
    }
}

func (p *ScriptedPlugin) Contexts() map[string]config.ContextConfig {
    // Locate helper script
    scriptPath := p.findHelperScript()
    
    return map[string]config.ContextConfig{
        "my-workflow": {
            Commands: map[string]string{
                "init":   fmt.Sprintf("bash %s init", scriptPath),
                "build":  fmt.Sprintf("bash %s build", scriptPath),
                "deploy": fmt.Sprintf("bash %s deploy", scriptPath),
            },
            Descriptions: map[string]string{
                "init":   "Initialize workflow",
                "build":  "Build with custom script",
                "deploy": "Deploy using helper script",
            },
        },
    }
}

func (p *ScriptedPlugin) findHelperScript() string {
    // Try installed location first
    installedPath := filepath.Join(os.Getenv("HOME"), ".toolbox", "scripts", "my_helper.sh")
    if _, err := os.Stat(installedPath); err == nil {
        return installedPath
    }
    
    // Try relative to executable
    exePath, err := os.Executable()
    if err == nil {
        relPath := filepath.Join(filepath.Dir(exePath), "scripts", "my_helper.sh")
        if absPath, err := filepath.Abs(relPath); err == nil {
            if _, err := os.Stat(absPath); err == nil {
                return absPath
            }
        }
    }
    
    // Fallback
    return "my_helper.sh"
}

func (p *ScriptedPlugin) Detect(dir string) (string, bool) {
    if _, err := os.Stat(filepath.Join(dir, ".myworkflow")); err == nil {
        return "my-workflow", true
    }
    return "", false
}

func (p *ScriptedPlugin) Name() string    { return p.name }
func (p *ScriptedPlugin) Version() string { return p.version }
func (p *ScriptedPlugin) Validate() error { return nil }
```

### Multi-Context Plugin

A single plugin providing multiple contexts:

```go
func (p *MultiPlugin) Contexts() map[string]config.ContextConfig {
    return map[string]config.ContextConfig{
        "frontend": {
            Commands: map[string]string{
                "dev":   "npm run dev",
                "build": "npm run build",
                "test":  "npm run test",
            },
        },
        "backend": {
            Commands: map[string]string{
                "dev":   "go run ./cmd/server",
                "build": "go build -o bin/server ./cmd/server",
                "test":  "go test ./...",
            },
        },
    }
}

func (p *MultiPlugin) Detect(dir string) (string, bool) {
    // Detect frontend
    if _, err := os.Stat(filepath.Join(dir, "src", "App.tsx")); err == nil {
        return "frontend", true
    }
    
    // Detect backend
    if _, err := os.Stat(filepath.Join(dir, "cmd", "server")); err == nil {
        return "backend", true
    }
    
    return "", false
}
```

## Plugin Registration

### Manual Registration

In `internal/cli/root.go` or similar initialization code:

```go
var pluginManager *plugin.PluginManager

func init() {
    pm := plugin.NewPluginManager("")
    
    // Register all plugins
    pm.RegisterPlugin(plugin.NewUbuntuPlugin())
    pm.RegisterPlugin(plugin.NewMyPlugin())
    
    pluginManager = pm
}

func getPluginManager() *plugin.PluginManager {
    return pluginManager
}
```

### Automatic Plugin Discovery (Future)

For future plugin discovery from a plugins directory:

```go
func loadPlugins() error {
    pm := plugin.NewPluginManager("~/.toolbox/plugins")
    
    // Auto-register plugins from directory
    plugins := []plugin.Plugin{
        plugin.NewUbuntuPlugin(),
        plugin.NewDockerPlugin(),
        plugin.NewKubernetesPlugin(),
    }
    
    for _, p := range plugins {
        if err := pm.RegisterPlugin(p); err != nil {
            return err
        }
    }
    
    return nil
}
```

## Best Practices

### 1. Naming Conventions

- **Plugin names**: lowercase-with-hyphens (e.g., `ubuntu-packaging`)
- **Context names**: descriptive and unique (e.g., `ubuntu-packaging`, not just `ubuntu`)
- **Command names**: short and intuitive (e.g., `build`, `test`, `deploy`)

### 2. Context Detection

- **Fast detection**: Use simple file existence checks
- **Specific markers**: Choose unique files to avoid false positives
- **Multiple indicators**: Check several files for reliability

```go
// Good: Fast and specific
func (p *Plugin) Detect(dir string) (string, bool) {
    markers := []string{"my.config", ".myproject", "MyProject.yml"}
    for _, marker := range markers {
        if _, err := os.Stat(filepath.Join(dir, marker)); err == nil {
            return "my-context", true
        }
    }
    return "", false
}

// Avoid: Slow or too generic
func (p *Plugin) Detect(dir string) (string, bool) {
    // Bad: Walking entire directory tree is slow
    filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
        // ...
    })
}
```

### 3. Command Design

- **Idempotent**: Commands should be safe to run multiple times
- **Clear output**: Provide feedback about what's happening
- **Fail fast**: Exit early on errors with clear messages
- **Dry-run support**: Honor `--dry-run` flag when possible

### 4. Descriptions

Always provide helpful descriptions:

```go
Descriptions: map[string]string{
    "build":  "Build project for production",
    "test":   "Run unit and integration tests",
    "deploy": "Deploy to staging environment",
}
```

### 5. Version Management

Use semantic versioning:

```go
version: "1.0.0"  // Initial release
version: "1.1.0"  // Added new command
version: "1.1.1"  // Bug fix
version: "2.0.0"  // Breaking change
```

## Testing Plugins

### Unit Tests

Create tests in `internal/plugin/my_plugin_test.go`:

```go
package plugin

import (
    "testing"
    "os"
    "path/filepath"
)

func TestMyPlugin_Detect(t *testing.T) {
    // Create temporary test directory
    tmpDir, err := os.MkdirTemp("", "plugin-test-*")
    if err != nil {
        t.Fatal(err)
    }
    defer os.RemoveAll(tmpDir)
    
    plugin := NewMyPlugin()
    
    // Test: Should not detect empty directory
    ctx, found := plugin.Detect(tmpDir)
    if found {
        t.Errorf("Expected no detection, got context: %s", ctx)
    }
    
    // Test: Should detect with marker file
    markerFile := filepath.Join(tmpDir, "my-marker.yml")
    if err := os.WriteFile(markerFile, []byte{}, 0644); err != nil {
        t.Fatal(err)
    }
    
    ctx, found = plugin.Detect(tmpDir)
    if !found {
        t.Error("Expected detection, got none")
    }
    if ctx != "my-context" {
        t.Errorf("Expected context 'my-context', got '%s'", ctx)
    }
}

func TestMyPlugin_Validate(t *testing.T) {
    plugin := NewMyPlugin()
    
    if err := plugin.Validate(); err != nil {
        t.Errorf("Validation failed: %v", err)
    }
}

func TestMyPlugin_Contexts(t *testing.T) {
    plugin := NewMyPlugin()
    contexts := plugin.Contexts()
    
    if len(contexts) == 0 {
        t.Error("Expected at least one context")
    }
    
    ctx, exists := contexts["my-context"]
    if !exists {
        t.Error("Expected 'my-context' to exist")
    }
    
    if len(ctx.Commands) == 0 {
        t.Error("Expected at least one command")
    }
}
```

### Integration Tests

Test plugin with actual ToolBox commands:

```bash
#!/bin/bash
# test_plugin.sh

# Create test project
mkdir -p /tmp/test-project
cd /tmp/test-project
echo "test: true" > my-marker.yml

# Test detection
output=$(tb build --dry-run 2>&1)
if echo "$output" | grep -q "Context: my-context"; then
    echo "✓ Plugin detection works"
else
    echo "✗ Plugin detection failed"
    exit 1
fi

# Test command execution
if tb build --dry-run | grep -q "my-build-command"; then
    echo "✓ Plugin command works"
else
    echo "✗ Plugin command failed"
    exit 1
fi

# Cleanup
cd -
rm -rf /tmp/test-project
```

## Security Considerations

### Input Validation

Always validate inputs in your detect and command logic:

```go
func (p *Plugin) Detect(dir string) (string, bool) {
    // Clean path to prevent directory traversal
    dir = filepath.Clean(dir)
    
    // Validate it's a directory
    info, err := os.Stat(dir)
    if err != nil || !info.IsDir() {
        return "", false
    }
    
    // Now safe to check for markers
    // ...
}
```

### Command Injection Prevention

When building commands from user input or detected values:

```go
// Bad: Vulnerable to command injection
command := fmt.Sprintf("deploy %s", userInput)

// Good: Use proper escaping or avoid shell execution
import "github.com/kballard/go-shellquote"

command := fmt.Sprintf("deploy %s", shellquote.Join(userInput))
```

### File Permissions

Check file permissions when dealing with sensitive operations:

```go
info, err := os.Stat(configFile)
if err != nil {
    return err
}

// Warn if permissions are too open
if info.Mode().Perm() & 0077 != 0 {
    log.Printf("Warning: %s has insecure permissions", configFile)
}
```

## Advanced Topics

### Sharing Data Between Commands

Use environment variables or temporary files:

```go
Commands: map[string]string{
    "init":  "echo 'initialized' > .my-state && export MY_VAR=value",
    "build": "[ -f .my-state ] && make build",
}
```

### Conditional Commands

Adjust commands based on environment:

```go
func (p *Plugin) Contexts() map[string]config.ContextConfig {
    buildCmd := "make build"
    if os.Getenv("CI") == "true" {
        buildCmd = "make build-ci"
    }
    
    return map[string]config.ContextConfig{
        "my-context": {
            Commands: map[string]string{
                "build": buildCmd,
            },
        },
    }
}
```

### Plugin Metadata

Add metadata for plugin discovery:

```go
type PluginMetadata struct {
    Author      string
    Description string
    Homepage    string
    License     string
}

func (p *MyPlugin) Metadata() PluginMetadata {
    return PluginMetadata{
        Author:      "Your Name",
        Description: "Plugin for XYZ workflow",
        Homepage:    "https://github.com/user/repo",
        License:     "MIT",
    }
}
```

## Next Steps

- Review the [Ubuntu Plugin](../internal/plugin/ubuntu.go) for a real-world example
- Check out the [Plugin Interface](../internal/plugin/plugin.go) source code
- See [Configuration Guide](configuration.md) for integrating plugins with config files
- Read [API Reference](api-reference.md) for complete plugin API documentation

## Get Help

- Open an issue on [GitHub Issues](https://github.com/bamf0/toolbox/issues)
- Join discussions on [GitHub Discussions](https://github.com/bamf0/toolbox/discussions)
