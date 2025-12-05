# ToolBox Plugin System

The ToolBox plugin system allows you to extend the tool with custom contexts and commands without modifying the core codebase.

## Overview

Plugins provide:
- **Custom Contexts**: Add support for new project types (Docker, Kubernetes, etc.)
- **Context Detection**: Automatically detect project types based on marker files
- **Commands**: Define commands specific to each context
- **Namespacing**: Avoid conflicts between plugins

## Built-in Plugins

ToolBox ships with two built-in plugins:

### 1. Docker Plugin
**Contexts**: `docker`, `docker-compose`

**Detection**:
- `docker`: Dockerfile present
- `docker-compose`: docker-compose.yml or docker-compose.yaml present

**Commands**:
```bash
tb build     # Build Docker image
tb run       # Run container
tb push      # Push to registry
tb compose   # Start docker-compose
tb stop      # Stop services
tb logs      # View logs
tb shell     # Enter container shell
```

### 2. Kubernetes Plugin
**Contexts**: `kubernetes`, `helm`

**Detection**:
- `kubernetes`: deployment.yaml, k8s/deployment.yaml, etc.
- `helm`: Chart.yaml present

**Commands**:
```bash
# Kubernetes
tb apply         # Apply manifests
tb delete        # Delete resources
tb get           # Get resources
tb logs          # View pod logs
tb describe      # Describe resources
tb exec          # Execute in pod
tb port-forward  # Port forward

# Helm
tb install   # Install chart
tb upgrade   # Upgrade release
tb rollback  # Rollback release
tb list      # List releases
tb delete    # Delete release
```

## Using Plugins

### List Installed Plugins
```bash
$ tb plugin list
NAME         VERSION   CONTEXTS   STATUS
────         ───────   ────────   ──────
docker       1.0.0     2          enabled
kubernetes   1.0.0     2          enabled
```

### View Plugin Details
```bash
$ tb plugin info docker
Plugin: docker
Version: 1.0.0
Status: enabled
Contexts: 2

Provided Contexts:
  - docker
  - docker-compose
```

### List Plugin Contexts
```bash
$ tb plugin contexts
CONTEXT                 COMMANDS
───────                 ────────
docker                  7
docker-compose          5
kubernetes              7
helm                    5
```

## Creating Custom Plugins

### Plugin Interface

A plugin must implement the `Plugin` interface:

```go
type Plugin interface {
    // Name returns the plugin's unique identifier
    Name() string
    
    // Version returns the plugin version
    Version() string
    
    // Contexts returns the contexts this plugin provides
    Contexts() map[string]config.ContextConfig
    
    // Detect attempts to detect if this plugin's context applies
    Detect(dir string) (string, bool)
    
    // Validate performs plugin self-validation
    Validate() error
}
```

### Example: Terraform Plugin

```go
package plugin

import (
    "fmt"
    "os"
    "path/filepath"
    
    "github.com/bamf0/toolbox/internal/config"
)

type TerraformPlugin struct {
    name    string
    version string
}

func NewTerraformPlugin() *TerraformPlugin {
    return &TerraformPlugin{
        name:    "terraform",
        version: "1.0.0",
    }
}

func (p *TerraformPlugin) Name() string {
    return p.name
}

func (p *TerraformPlugin) Version() string {
    return p.version
}

func (p *TerraformPlugin) Contexts() map[string]config.ContextConfig {
    return map[string]config.ContextConfig{
        "terraform": {
            Commands: map[string]string{
                "init":    "terraform init",
                "plan":    "terraform plan",
                "apply":   "terraform apply",
                "destroy": "terraform destroy",
                "fmt":     "terraform fmt",
                "validate": "terraform validate",
            },
        },
    }
}

func (p *TerraformPlugin) Detect(dir string) (string, bool) {
    // Check for *.tf files
    tfFiles, err := filepath.Glob(filepath.Join(dir, "*.tf"))
    if err == nil && len(tfFiles) > 0 {
        return "terraform", true
    }
    
    return "", false
}

func (p *TerraformPlugin) Validate() error {
    if p.name == "" {
        return fmt.Errorf("plugin name cannot be empty")
    }
    if p.version == "" {
        return fmt.Errorf("plugin version cannot be empty")
    }
    return nil
}
```

### Registering a Plugin

To register your custom plugin, add it to `plugin_cmd.go`:

```go
func getPluginManager() *plugin.PluginManager {
    pm := plugin.NewPluginManager("")
    
    // Register built-in plugins
    pm.RegisterPlugin(plugin.NewDockerPlugin())
    pm.RegisterPlugin(plugin.NewKubernetesPlugin())
    
    // Register your custom plugin
    pm.RegisterPlugin(NewTerraformPlugin())
    
    return pm
}
```

## Plugin Architecture

### Context Priority

When multiple plugins can detect the same directory:
1. Plugins are checked in registration order
2. First matching plugin wins
3. User can override with `--context` flag

### Namespacing

Contexts are available in two forms:
- **Non-namespaced**: `docker` (simpler, potential conflicts)
- **Namespaced**: `docker:docker` (explicit, no conflicts)

Example:
```bash
# Both work
tb build                # Uses non-namespaced context
tb --context docker:docker build  # Uses namespaced context
```

### Config Precedence

Configuration is merged with this priority:
1. User config file (highest priority)
2. Plugin contexts
3. Built-in defaults (lowest priority)

## Security Considerations

### Plugin Validation

All plugins are validated before registration:
- ✅ Name must be non-empty
- ✅ Version must be non-empty
- ✅ Contexts must be non-empty
- ✅ Commands must be non-empty

### Hash Allowlisting (Future)

For dynamically loaded plugins (.so/.dll files):
```go
pm := plugin.NewPluginManager("/path/to/plugins")

// Add trusted plugin hashes
pm.AddTrustedHash("sha256:abc123...")

// Load plugin (rejected if hash doesn't match)
err := pm.LoadPlugin("plugin.so")
```

### Command Execution

Plugin commands are executed with the same security controls as built-in commands:
- ✅ No shell interpretation (direct execution)
- ✅ Argument validation (length/count limits)
- ✅ Timeout enforcement (10-minute default)
- ✅ Path validation for executables

## Testing Plugins

### Unit Tests

```go
func TestTerraformPlugin_Detect(t *testing.T) {
    plugin := NewTerraformPlugin()
    tmpDir := t.TempDir()
    
    // Create test file
    tfFile := filepath.Join(tmpDir, "main.tf")
    os.WriteFile(tfFile, []byte("# terraform"), 0644)
    
    // Test detection
    ctx, detected := plugin.Detect(tmpDir)
    
    if !detected {
        t.Error("expected detection to succeed")
    }
    if ctx != "terraform" {
        t.Errorf("expected context 'terraform', got %q", ctx)
    }
}
```

### Integration Tests

```bash
# Create test project
mkdir test-terraform
cd test-terraform
echo '# test' > main.tf

# Test detection
tb build --dry-run
# Expected: Context: terraform
#           Command: terraform init
```

## Best Practices

### 1. Specific Detection
```go
// ✅ Good: Specific marker files
func (p *Plugin) Detect(dir string) (string, bool) {
    if fileExists(filepath.Join(dir, "terraform.tfvars")) {
        return "terraform", true
    }
    return "", false
}

// ❌ Bad: Generic files
func (p *Plugin) Detect(dir string) (string, bool) {
    if fileExists(filepath.Join(dir, "README.md")) {
        return "mycontext", true  // Too generic!
    }
    return "", false
}
```

### 2. Clear Command Names
```go
// ✅ Good: Clear, intuitive names
Commands: map[string]string{
    "build":  "docker build",
    "run":    "docker run",
    "deploy": "docker push",
}

// ❌ Bad: Cryptic abbreviations
Commands: map[string]string{
    "bld": "docker build",
    "r":   "docker run",
    "dpy": "docker push",
}
```

### 3. Validation
```go
// ✅ Good: Comprehensive validation
func (p *Plugin) Validate() error {
    if p.name == "" {
        return fmt.Errorf("name required")
    }
    for ctx, cfg := range p.Contexts() {
        if len(cfg.Commands) == 0 {
            return fmt.Errorf("context %q has no commands", ctx)
        }
    }
    return nil
}
```

## Future Enhancements

### Dynamic Plugin Loading
```yaml
# ~/.toolbox/config.yaml
plugins:
  - ~/.toolbox/plugins/terraform.so
  - ~/.toolbox/plugins/ansible.so
```

### Plugin Repository
```bash
# Install from registry
tb plugin install terraform

# Update plugins
tb plugin update

# Uninstall plugins
tb plugin remove terraform
```

### Plugin Dependencies
```go
type Plugin interface {
    // ... existing methods ...
    
    Dependencies() []string  // Required plugins
    Conflicts() []string     // Incompatible plugins
}
```

## Troubleshooting

### Plugin Not Detected
```bash
# Check if plugin is registered
tb plugin list

# Check plugin detection
tb --context <plugin-context> build --dry-run

# Enable verbose output
tb build --dry-run -v
```

### Command Not Found
```bash
# List available commands for context
tb plugin info <plugin-name>

# Check context commands
tb plugin contexts
```

### Conflicts
If two plugins provide the same context:
```bash
# Use namespaced context
tb --context plugin-name:context build
```

## Examples

### Example 1: Ansible Plugin
```go
func NewAnsiblePlugin() *AnsiblePlugin {
    return &AnsiblePlugin{
        name: "ansible",
        version: "1.0.0",
    }
}

func (p *AnsiblePlugin) Contexts() map[string]config.ContextConfig {
    return map[string]config.ContextConfig{
        "ansible": {
            Commands: map[string]string{
                "playbook": "ansible-playbook playbook.yml",
                "galaxy":   "ansible-galaxy install",
                "lint":     "ansible-lint",
                "vault":    "ansible-vault",
            },
        },
    }
}

func (p *AnsiblePlugin) Detect(dir string) (string, bool) {
    return fileExists(filepath.Join(dir, "ansible.cfg")), true
}
```

### Example 2: CDK Plugin
```go
func NewCDKPlugin() *CDKPlugin {
    return &CDKPlugin{name: "cdk", version: "1.0.0"}
}

func (p *CDKPlugin) Contexts() map[string]config.ContextConfig {
    return map[string]config.ContextConfig{
        "cdk": {
            Commands: map[string]string{
                "synth":   "cdk synth",
                "deploy":  "cdk deploy",
                "destroy": "cdk destroy",
                "diff":    "cdk diff",
            },
        },
    }
}

func (p *CDKPlugin) Detect(dir string) (string, bool) {
    return fileExists(filepath.Join(dir, "cdk.json")), true
}
```

---

## Summary

The plugin system provides:
✅ Easy extensibility without code changes  
✅ Type-safe plugin interface  
✅ Automatic context detection  
✅ Security validation  
✅ Namespace support for conflicts  
✅ Comprehensive testing support  

**Get started**: Create your plugin by implementing the `Plugin` interface!
