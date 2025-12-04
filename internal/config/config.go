package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the toolbox configuration
type Config struct {
	Contexts map[string]ContextConfig `yaml:"contexts"`
}

// ContextConfig defines commands for a specific context
type ContextConfig struct {
	Commands map[string]string `yaml:"commands"`
}

// Load reads and parses the configuration file
// Priority: specified file > .toolbox.yaml (cwd) > ~/.toolbox/config.yaml > defaults
func Load(cfgFile string) (*Config, error) {
	var cfg *Config

	// Try specified file first
	if cfgFile != "" {
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
	cfg = getDefaultConfig()
	return cfg, nil
}

// loadFromFile reads and parses a YAML config file
func loadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	// Merge with defaults for any missing contexts
	defaults := getDefaultConfig()
	for ctxName, ctxCfg := range defaults.Contexts {
		if _, exists := cfg.Contexts[ctxName]; !exists {
			if cfg.Contexts == nil {
				cfg.Contexts = make(map[string]ContextConfig)
			}
			cfg.Contexts[ctxName] = ctxCfg
		}
	}

	return &cfg, nil
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
	_, err := os.Stat(path)
	return err == nil
}
