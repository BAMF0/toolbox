// Package plugin provides a secure plugin system for ToolBox.
// It allows third-party extensions to add custom contexts and commands
// while maintaining security through validation and sandboxing.
package plugin

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bamf0/toolbox/internal/config"
)

// Plugin represents a ToolBox plugin that can provide custom contexts and commands.
type Plugin interface {
	// Name returns the plugin's unique identifier
	Name() string

	// Version returns the plugin version (semantic versioning recommended)
	Version() string

	// Contexts returns the contexts this plugin provides
	Contexts() map[string]config.ContextConfig

	// Detect attempts to detect if this plugin's context applies to the given directory
	// Returns the context name and true if detected, empty string and false otherwise
	Detect(dir string) (string, bool)

	// Validate performs plugin self-validation
	// Returns error if plugin is in an invalid state
	Validate() error
}

// PluginMetadata contains information about a loaded plugin
type PluginMetadata struct {
	Name         string
	Version      string
	Path         string
	Hash         string // SHA256 hash for verification
	Enabled      bool
	ContextCount int
	Contexts     []string
}

// PluginManager manages loading and lifecycle of plugins
type PluginManager struct {
	plugins       []Plugin
	metadata      map[string]*PluginMetadata
	pluginDir     string
	allowedHashes map[string]bool // Allowlist of trusted plugin hashes
}

// NewPluginManager creates a new plugin manager
func NewPluginManager(pluginDir string) *PluginManager {
	return &PluginManager{
		plugins:       make([]Plugin, 0),
		metadata:      make(map[string]*PluginMetadata),
		pluginDir:     pluginDir,
		allowedHashes: make(map[string]bool),
	}
}

// LoadPluginsFromConfig loads plugins specified in the config
func (pm *PluginManager) LoadPluginsFromConfig(pluginPaths []string) error {
	for _, path := range pluginPaths {
		if err := pm.LoadPlugin(path); err != nil {
			return fmt.Errorf("failed to load plugin %s: %w", path, err)
		}
	}
	return nil
}

// LoadPlugin loads a single plugin from the given path
// Note: In this implementation, we use Go-based plugins (*.so on Unix, *.dll on Windows)
// For security, we validate plugins before loading
func (pm *PluginManager) LoadPlugin(path string) error {
	// Validate path
	if err := validatePluginPath(path); err != nil {
		return fmt.Errorf("invalid plugin path: %w", err)
	}

	// Calculate and verify hash
	hash, err := calculateFileHash(path)
	if err != nil {
		return fmt.Errorf("failed to calculate plugin hash: %w", err)
	}

	// Check if plugin is in allowlist (if allowlist is configured)
	if len(pm.allowedHashes) > 0 && !pm.allowedHashes[hash] {
		return fmt.Errorf("plugin not in allowlist (hash: %s)", hash)
	}

	// For now, we'll create a registry-based plugin system
	// instead of native .so/.dll plugins for better security and portability
	// This avoids the security risks of loading arbitrary native code

	return fmt.Errorf("native plugin loading not yet implemented - use registry-based plugins")
}

// RegisterPlugin registers a pre-compiled plugin (safer than dynamic loading)
func (pm *PluginManager) RegisterPlugin(plugin Plugin) error {
	// Validate plugin
	if err := plugin.Validate(); err != nil {
		return fmt.Errorf("plugin validation failed: %w", err)
	}

	// Check for name conflicts
	name := plugin.Name()
	if _, exists := pm.metadata[name]; exists {
		return fmt.Errorf("plugin with name %q already registered", name)
	}

	// Add to plugins list
	pm.plugins = append(pm.plugins, plugin)

	// Create metadata
	contexts := plugin.Contexts()
	contextNames := make([]string, 0, len(contexts))
	for ctx := range contexts {
		contextNames = append(contextNames, ctx)
	}

	pm.metadata[name] = &PluginMetadata{
		Name:         name,
		Version:      plugin.Version(),
		Enabled:      true,
		ContextCount: len(contexts),
		Contexts:     contextNames,
	}

	return nil
}

// GetPlugins returns all registered plugins
func (pm *PluginManager) GetPlugins() []Plugin {
	return pm.plugins
}

// GetMetadata returns metadata for all plugins
func (pm *PluginManager) GetMetadata() map[string]*PluginMetadata {
	return pm.metadata
}

// DetectContext attempts to detect context using all loaded plugins
// Returns the first matching context and the plugin that detected it
func (pm *PluginManager) DetectContext(dir string) (context string, pluginName string, found bool) {
	for _, plugin := range pm.plugins {
		if ctx, detected := plugin.Detect(dir); detected {
			return ctx, plugin.Name(), true
		}
	}
	return "", "", false
}

// GetContexts returns all contexts from all plugins
func (pm *PluginManager) GetContexts() map[string]config.ContextConfig {
	allContexts := make(map[string]config.ContextConfig)

	for _, plugin := range pm.plugins {
		for ctxName, ctxConfig := range plugin.Contexts() {
			// Namespace context names with plugin name to avoid conflicts
			namespacedName := plugin.Name() + ":" + ctxName
			allContexts[namespacedName] = ctxConfig

			// Also add without namespace if no conflict
			if _, exists := allContexts[ctxName]; !exists {
				allContexts[ctxName] = ctxConfig
			}
		}
	}

	return allContexts
}

// AddTrustedHash adds a plugin hash to the allowlist
func (pm *PluginManager) AddTrustedHash(hash string) {
	pm.allowedHashes[hash] = true
}

// RemoveTrustedHash removes a plugin hash from the allowlist
func (pm *PluginManager) RemoveTrustedHash(hash string) {
	delete(pm.allowedHashes, hash)
}

// validatePluginPath validates the plugin path for security
func validatePluginPath(path string) error {
	// Clean path
	cleanPath := filepath.Clean(path)

	// Prevent directory traversal
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("directory traversal not allowed")
	}

	// Check if file exists
	info, err := os.Stat(cleanPath)
	if err != nil {
		return fmt.Errorf("plugin file not accessible: %w", err)
	}

	// Must be a regular file
	if !info.Mode().IsRegular() {
		return fmt.Errorf("plugin must be a regular file")
	}

	// Check file extension (platform-specific)
	ext := filepath.Ext(cleanPath)
	validExts := []string{".so", ".dll", ".dylib"}
	valid := false
	for _, validExt := range validExts {
		if ext == validExt {
			valid = true
			break
		}
	}

	if !valid {
		return fmt.Errorf("invalid plugin extension %q (expected: .so, .dll, or .dylib)", ext)
	}

	return nil
}

// calculateFileHash calculates SHA256 hash of a file
func calculateFileHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
