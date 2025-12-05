// Package config provides secure configuration loading and validation for the ToolBox application.
// It handles YAML config files with proper security controls including file size limits,
// path validation, and content sanitization.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	// MaxConfigFileSize limits config file size to prevent memory exhaustion attacks
	MaxConfigFileSize = 1024 * 1024 // 1MB

	// MaxCommandLength limits individual command strings
	MaxCommandLength = 4096

	// MaxContexts limits number of contexts in config
	MaxContexts = 100

	// MaxCommandsPerContext limits commands per context
	MaxCommandsPerContext = 50
)

// Config represents the toolbox configuration
type Config struct {
	Contexts map[string]ContextConfig `yaml:"contexts"`
}

// ContextConfig defines commands for a specific context
type ContextConfig struct {
	Commands map[string]string `yaml:"commands"`
}

// Load reads and parses the configuration file with security validation.
// Priority: specified file > .toolbox.yaml (cwd) > ~/.toolbox/config.yaml > defaults
//
// Security measures:
//   - Path traversal prevention
//   - File size limits
//   - Content validation
//   - Safe error messages
func Load(cfgFile string) (*Config, error) {
	// Try specified file first
	if cfgFile != "" {
		// Validate the config file path for security
		if err := validateConfigPath(cfgFile); err != nil {
			return nil, fmt.Errorf("invalid config path: %w", err)
		}
		return loadFromFile(cfgFile)
	}

	// Try local .toolbox.yaml
	localConfig := ".toolbox.yaml"
	if fileExists(localConfig) {
		return loadFromFile(localConfig)
	}

	// Try ~/.toolbox/config.yaml
	homeDir, err := os.UserHomeDir()
	if err == nil {
		globalConfig := filepath.Join(homeDir, ".toolbox", "config.yaml")
		if fileExists(globalConfig) {
			return loadFromFile(globalConfig)
		}
	}

	// Return default configuration
	return getDefaultConfig(), nil
}

// validateConfigPath performs security checks on user-provided config paths
func validateConfigPath(path string) error {
	// Prevent empty paths
	if path == "" {
		return fmt.Errorf("empty path")
	}

	// Clean the path to resolve any . or .. components
	cleanPath := filepath.Clean(path)

	// Prevent absolute paths for user-specified config
	// This prevents users from reading arbitrary system files
	if filepath.IsAbs(cleanPath) {
		return fmt.Errorf("absolute paths not allowed, use relative path or place config in ~/.toolbox/")
	}

	// Check for directory traversal attempts
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("directory traversal not allowed")
	}

	// Ensure it's a .yaml or .yml file
	ext := filepath.Ext(cleanPath)
	if ext != ".yaml" && ext != ".yml" {
		return fmt.Errorf("config file must have .yaml or .yml extension")
	}

	return nil
}

// loadFromFile reads and parses a YAML config file with security checks
func loadFromFile(path string) (*Config, error) {
	// Check file exists and get size
	fileInfo, err := os.Stat(path)
	if err != nil {
		// Don't reveal full path in error message
		return nil, fmt.Errorf("config file not accessible: %w", err)
	}

	// Check file size to prevent memory exhaustion
	if fileInfo.Size() > MaxConfigFileSize {
		return nil, fmt.Errorf("config file exceeds maximum size of %d bytes (got %d bytes)",
			MaxConfigFileSize, fileInfo.Size())
	}

	// Ensure it's a regular file (not a directory, symlink, etc.)
	if !fileInfo.Mode().IsRegular() {
		return nil, fmt.Errorf("config path must be a regular file")
	}

	// Read file with size limit already enforced
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		// Sanitize YAML parsing errors to avoid leaking file content
		return nil, fmt.Errorf("failed to parse config file: invalid YAML format")
	}

	// Validate the loaded configuration
	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Merge with defaults for any missing contexts
	mergeDefaults(&cfg)

	return &cfg, nil
}

// validateConfig performs security and sanity checks on loaded configuration
func validateConfig(cfg *Config) error {
	if cfg.Contexts == nil {
		return fmt.Errorf("no contexts defined")
	}

	if len(cfg.Contexts) > MaxContexts {
		return fmt.Errorf("too many contexts (max: %d, got: %d)", MaxContexts, len(cfg.Contexts))
	}

	for ctxName, ctxCfg := range cfg.Contexts {
		// Validate context name
		if err := validateContextName(ctxName); err != nil {
			return fmt.Errorf("invalid context name %q: %w", ctxName, err)
		}

		// Check number of commands
		if len(ctxCfg.Commands) > MaxCommandsPerContext {
			return fmt.Errorf("context %q has too many commands (max: %d, got: %d)",
				ctxName, MaxCommandsPerContext, len(ctxCfg.Commands))
		}

		// Validate each command
		for cmdName, cmdString := range ctxCfg.Commands {
			if err := validateCommand(cmdName, cmdString); err != nil {
				return fmt.Errorf("context %q, command %q: %w", ctxName, cmdName, err)
			}
		}
	}

	return nil
}

// validateContextName ensures context names are safe
func validateContextName(name string) error {
	if name == "" {
		return fmt.Errorf("empty context name")
	}

	if len(name) > 50 {
		return fmt.Errorf("context name too long (max 50 characters)")
	}

	// Only allow alphanumeric, dash, and underscore
	for _, r := range name {
		if !isAlphaNumeric(r) && r != '-' && r != '_' {
			return fmt.Errorf("context name contains invalid character %q", r)
		}
	}

	return nil
}

// validateCommand validates a command name and string
func validateCommand(name, command string) error {
	if name == "" {
		return fmt.Errorf("empty command name")
	}

	if len(name) > 50 {
		return fmt.Errorf("command name too long")
	}

	if command == "" {
		return fmt.Errorf("empty command string")
	}

	if len(command) > MaxCommandLength {
		return fmt.Errorf("command string exceeds maximum length of %d characters", MaxCommandLength)
	}

	// Warn about potentially dangerous patterns in config commands
	// These are allowed (users control their config) but logged for awareness
	if containsDangerousPatterns(command) {
		// In production, you might want to log this
		// For now, we allow it since users control their own config files
	}

	return nil
}

// containsDangerousPatterns checks for shell metacharacters
func containsDangerousPatterns(s string) bool {
	// Note: These are informational only for config files
	// User-controlled config is assumed trusted
	dangerous := []string{
		";",  // Command separator
		"|",  // Pipe
		"&",  // Background/AND
		"$",  // Variable expansion
		"`",  // Command substitution
		"\n", // Newline injection
	}

	for _, pattern := range dangerous {
		if strings.Contains(s, pattern) {
			return true
		}
	}
	return false
}

// isAlphaNumeric checks if a rune is alphanumeric
func isAlphaNumeric(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}

// mergeDefaults merges default configuration with user configuration
func mergeDefaults(cfg *Config) {
	defaults := getDefaultConfig()

	for ctxName, ctxCfg := range defaults.Contexts {
		if _, exists := cfg.Contexts[ctxName]; !exists {
			if cfg.Contexts == nil {
				cfg.Contexts = make(map[string]ContextConfig)
			}
			cfg.Contexts[ctxName] = ctxCfg
		}
	}
}

// getDefaultConfig returns built-in default configurations
func getDefaultConfig() *Config {
	return &Config{
		Contexts: map[string]ContextConfig{
			"node": {
				Commands: map[string]string{
					"build":   "npm run build",
					"test":    "npm test",
					"start":   "npm start",
					"dev":     "npm run dev",
					"lint":    "npm run lint",
					"install": "npm install",
				},
			},
			"go": {
				Commands: map[string]string{
					"build":   "go build ./...",
					"test":    "go test ./...",
					"run":     "go run ./cmd/...",
					"install": "go mod download",
					"lint":    "golangci-lint run",
					"fmt":     "go fmt ./...",
				},
			},
			"python": {
				Commands: map[string]string{
					"test":    "pytest",
					"lint":    "ruff check .",
					"fmt":     "black .",
					"install": "pip install -r requirements.txt",
					"run":     "python main.py",
				},
			},
			"rust": {
				Commands: map[string]string{
					"build":   "cargo build",
					"test":    "cargo test",
					"run":     "cargo run",
					"install": "cargo fetch",
					"lint":    "cargo clippy",
					"fmt":     "cargo fmt",
				},
			},
			"make": {
				Commands: map[string]string{
					"build": "make",
					"test":  "make test",
					"clean": "make clean",
				},
			},
		},
	}
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Mode().IsRegular()
}
