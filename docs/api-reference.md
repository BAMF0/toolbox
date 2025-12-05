# API Reference

Complete API documentation for ToolBox plugin development.

## Table of Contents

- [Plugin Interface](#plugin-interface)
- [Config Types](#config-types)
- [Plugin Manager](#plugin-manager)
- [Helper Functions](#helper-functions)
- [Ubuntu Plugin API](#ubuntu-plugin-api)

## Plugin Interface

### Plugin

Main interface that all plugins must implement.

```go
type Plugin interface {
    Name() string
    Version() string
    Contexts() map[string]config.ContextConfig
    Detect(dir string) (string, bool)
    Validate() error
}
```

#### Name

```go
Name() string
```

Returns the unique identifier for this plugin.

**Requirements**:
- Must be unique across all plugins
- Should be lowercase with hyphens
- Max 50 characters
- Alphanumeric, dash, and underscore only

**Example**:
```go
func (p *MyPlugin) Name() string {
    return "my-plugin"
}
```

#### Version

```go
Version() string
```

Returns the plugin version.

**Recommendations**:
- Use semantic versioning (e.g., "1.0.0")
- Increment on breaking changes, new features, or bug fixes

**Example**:
```go
func (p *MyPlugin) Version() string {
    return "1.2.3"
}
```

#### Contexts

```go
Contexts() map[string]config.ContextConfig
```

Returns all contexts provided by this plugin.

**Return Value**:
- Map key: context name
- Map value: ContextConfig with commands and descriptions

**Example**:
```go
func (p *MyPlugin) Contexts() map[string]config.ContextConfig {
    return map[string]config.ContextConfig{
        "my-context": {
            Commands: map[string]string{
                "build": "make build",
                "test":  "make test",
            },
            Descriptions: map[string]string{
                "build": "Build the project",
                "test":  "Run tests",
            },
        },
    }
}
```

#### Detect

```go
Detect(dir string) (string, bool)
```

Attempts to detect if this plugin's context applies to the given directory.

**Parameters**:
- `dir`: Directory path to check (usually current working directory)

**Returns**:
- `string`: Context name if detected
- `bool`: `true` if detected, `false` otherwise

**Example**:
```go
func (p *MyPlugin) Detect(dir string) (string, bool) {
    markerFile := filepath.Join(dir, "my-config.yml")
    if _, err := os.Stat(markerFile); err == nil {
        return "my-context", true
    }
    return "", false
}
```

**Best Practices**:
- Keep detection fast (avoid expensive operations)
- Check for specific marker files or directories
- Use multiple indicators for reliability
- Return early when detected

#### Validate

```go
Validate() error
```

Performs plugin self-validation.

**Returns**:
- `nil` if valid
- `error` describing validation failure

**Example**:
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

## Config Types

### ContextConfig

Configuration for a single context.

```go
type ContextConfig struct {
    Commands     map[string]string
    Descriptions map[string]string
}
```

**Fields**:
- `Commands`: Map of command name to shell command string
- `Descriptions`: Map of command name to human-readable description (optional)

**Example**:
```go
cfg := config.ContextConfig{
    Commands: map[string]string{
        "build": "npm run build",
        "test":  "npm test",
    },
    Descriptions: map[string]string{
        "build": "Build the project",
        "test":  "Run unit tests",
    },
}
```

### Config

Top-level configuration structure.

```go
type Config struct {
    Contexts map[string]ContextConfig
}
```

**Fields**:
- `Contexts`: Map of context name to ContextConfig

**Loading**:
```go
cfg, err := config.Load("") // Load default config
cfg, err := config.Load(".toolbox.yaml") // Load specific file
```

## Plugin Manager

### PluginManager

Manages loading and lifecycle of plugins.

```go
type PluginManager struct {
    // Private fields
}
```

#### NewPluginManager

```go
func NewPluginManager(pluginDir string) *PluginManager
```

Creates a new plugin manager.

**Parameters**:
- `pluginDir`: Directory for plugins (currently unused, for future expansion)

**Example**:
```go
pm := plugin.NewPluginManager("")
```

#### RegisterPlugin

```go
func (pm *PluginManager) RegisterPlugin(plugin Plugin) error
```

Registers a pre-compiled plugin.

**Parameters**:
- `plugin`: Plugin instance to register

**Returns**:
- `nil` on success
- `error` if validation fails or plugin name conflicts

**Example**:
```go
pm := plugin.NewPluginManager("")
if err := pm.RegisterPlugin(plugin.NewMyPlugin()); err != nil {
    log.Fatal(err)
}
```

#### GetPlugins

```go
func (pm *PluginManager) GetPlugins() []Plugin
```

Returns all registered plugins.

**Example**:
```go
plugins := pm.GetPlugins()
for _, p := range plugins {
    fmt.Printf("Plugin: %s v%s\n", p.Name(), p.Version())
}
```

#### GetMetadata

```go
func (pm *PluginManager) GetMetadata() map[string]*PluginMetadata
```

Returns metadata for all plugins.

**Return Value**:
- Map of plugin name to PluginMetadata

**Example**:
```go
metadata := pm.GetMetadata()
for name, meta := range metadata {
    fmt.Printf("%s: %d contexts\n", name, meta.ContextCount)
}
```

#### DetectContext

```go
func (pm *PluginManager) DetectContext(dir string) (context string, pluginName string, found bool)
```

Attempts to detect context using all loaded plugins.

**Parameters**:
- `dir`: Directory to check

**Returns**:
- `context`: Detected context name
- `pluginName`: Name of plugin that detected it
- `found`: Whether any plugin detected a context

**Example**:
```go
ctx, pluginName, found := pm.DetectContext(".")
if found {
    fmt.Printf("Detected %s context (via %s plugin)\n", ctx, pluginName)
}
```

#### GetContexts

```go
func (pm *PluginManager) GetContexts() map[string]config.ContextConfig
```

Returns all contexts from all plugins.

**Note**: Contexts are namespaced as `plugin-name:context-name` to avoid conflicts.

**Example**:
```go
contexts := pm.GetContexts()
for ctxName := range contexts {
    fmt.Println(ctxName)
}
// Output:
// ubuntu:ubuntu-packaging
// ubuntu-packaging
// my-plugin:my-context
// my-context
```

### PluginMetadata

Metadata about a loaded plugin.

```go
type PluginMetadata struct {
    Name         string
    Version      string
    Path         string
    Hash         string
    Enabled      bool
    ContextCount int
    Contexts     []string
}
```

**Fields**:
- `Name`: Plugin name
- `Version`: Plugin version
- `Path`: File path (for dynamically loaded plugins)
- `Hash`: SHA256 hash (for verification)
- `Enabled`: Whether plugin is active
- `ContextCount`: Number of contexts provided
- `Contexts`: List of context names

## Helper Functions

### File Operations

#### fileExists

```go
func fileExists(path string) bool
```

Checks if a file exists and is a regular file.

**Example**:
```go
if fileExists("package.json") {
    // Node.js project detected
}
```

### Path Validation

#### validatePluginPath

```go
func validatePluginPath(path string) error
```

Validates plugin path for security.

**Checks**:
- No empty paths
- No directory traversal
- Must be regular file
- Must have valid extension (.so, .dll, .dylib)

### Hash Calculation

#### calculateFileHash

```go
func calculateFileHash(path string) (string, error)
```

Calculates SHA256 hash of a file.

**Returns**:
- Hex-encoded hash string
- Error if file can't be read

## Ubuntu Plugin API

### UbuntuPlugin

Plugin for Ubuntu/Debian packaging workflows.

```go
type UbuntuPlugin struct {
    // Private fields
}
```

#### NewUbuntuPlugin

```go
func NewUbuntuPlugin() *UbuntuPlugin
```

Creates a new Ubuntu plugin instance.

### PPAInfo

Contains parsed PPA metadata.

```go
type PPAInfo struct {
    Release     string  // Ubuntu release (e.g., "noble", "jammy")
    Project     string  // Project name
    Type        string  // "merge", "sru", or "bug"
    BugID       string  // Bug ID (e.g., "2133493")
    Description string  // Optional description
    FullName    string  // Original PPA name
}
```

#### ParsePPAName

```go
func ParsePPAName(ppaName string) (*PPAInfo, error)
```

Parses a PPA name into its components.

**Supported Formats**:
- Merge: `<release>-<project>-merge-lp<bug>`
- SRU: `<release>-<project>-sru-lp<bug>-<desc>`
- Bug: `<release>-<project>-lp<bug>-<desc>`

**Example**:
```go
info, err := ParsePPAName("noble-efibootmgr-merge-lp2133493")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Release: %s, Bug: %s\n", info.Release, info.BugID)
```

#### CreatePPAName

```go
func CreatePPAName(project, bugID, ppaType, description string) (string, error)
```

Generates a PPA name from components.

**Parameters**:
- `project`: Project name
- `bugID`: Launchpad bug ID (with or without "lp" prefix)
- `ppaType`: "merge", "sru", or "bug"
- `description`: Optional description

**Example**:
```go
ppaName, err := CreatePPAName("myproject", "2133493", "merge", "")
// Returns: "noble-myproject-merge-lp2133493"
```

#### ParseBranchName

```go
func ParseBranchName(branchName string) (*PPAInfo, error)
```

Extracts PPA information from a git branch name.

**Branch Formats**:
- Merge: `merge-lp<bug>`
- SRU: `sru-lp<bug>-<release>`
- Bug: `bug-lp<bug>-<release>` or `lp<bug>-<release>`

**Example**:
```go
info, err := ParseBranchName("merge-lp2133493")
if err != nil {
    log.Fatal(err)
}
fmt.Println(info.GetBranchName())
```

#### PPAInfo Methods

##### GetPPATarget

```go
func (info *PPAInfo) GetPPATarget(username string) string
```

Returns the PPA target for dput.

**Example**:
```go
target := info.GetPPATarget("myuser")
// Returns: "ppa:myuser/noble-myproject-merge-lp2133493"
```

##### GetBranchName

```go
func (info *PPAInfo) GetBranchName() string
```

Returns the appropriate git branch name.

**Example**:
```go
branch := info.GetBranchName()
// For merge: "merge-lp2133493"
// For SRU: "sru-lp2133493-noble"
```

##### GetChangelogMessage

```go
func (info *PPAInfo) GetChangelogMessage() string
```

Returns a changelog entry message.

**Example**:
```go
msg := info.GetChangelogMessage()
// Returns: "* Merge from Debian. LP: #2133493"
```

##### GetVersionSuffix

```go
func (info *PPAInfo) GetVersionSuffix(currentVersion string) string
```

Returns the version suffix for the release.

**Format**: `~<release><n>` where n starts at 1

**Example**:
```go
suffix := info.GetVersionSuffix("1.0-1ubuntu1")
// Returns: "~noble1"
```

##### String

```go
func (info *PPAInfo) String() string
```

Returns a formatted summary of PPA info.

### Helper Functions

#### DetectProjectName

```go
func DetectProjectName() (string, error)
```

Reads project name from `debian/control`.

#### DetectUbuntuRelease

```go
func DetectUbuntuRelease() (string, error)
```

Reads current Ubuntu release from `debian/changelog`.

#### IsInPackagingDir

```go
func IsInPackagingDir() bool
```

Checks if we're in a Debian/Ubuntu packaging directory.

## Constants

### Plugin Types

```go
const (
    PPATypeMerge = "merge"
    PPATypeSRU   = "sru"
    PPATypeBug   = "bug"
)
```

### Configuration Limits

```go
const (
    MaxConfigFileSize     = 1024 * 1024  // 1MB
    MaxCommandLength      = 4096
    MaxContexts           = 100
    MaxCommandsPerContext = 50
)
```

## Error Handling

### Common Errors

**Plugin Registration**:
```go
// Name conflict
err := pm.RegisterPlugin(plugin)
// Returns: "plugin with name \"my-plugin\" already registered"

// Validation failure
err := pm.RegisterPlugin(plugin)
// Returns: "plugin validation failed: plugin must provide at least one context"
```

**Context Detection**:
```go
info, err := ParsePPAName("invalid-name")
// Returns: "invalid PPA name format: invalid-name"

info, err := ParseBranchName("main")
// Returns: "branch name does not contain a valid Launchpad bug ID: main"
```

**Configuration**:
```go
cfg, err := config.Load("huge.yaml")
// Returns: "config file exceeds maximum size of 1048576 bytes"

cfg, err := config.Load("/etc/passwd")
// Returns: "invalid config path: absolute paths not allowed"
```

## Next Steps

- See [Plugin Development Guide](plugin-development.md) for usage examples
- Check [Ubuntu Plugin source](../internal/plugin/ubuntu.go) for real implementation
- Review [Config source](../internal/config/config.go) for details
