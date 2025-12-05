package cli

import (
	"fmt"
	"sort"

	"github.com/bamf0/toolbox/internal/config"
	contextpkg "github.com/bamf0/toolbox/internal/context"
	"github.com/bamf0/toolbox/internal/plugin"
	"github.com/bamf0/toolbox/internal/registry"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current context and available commands",
	Long:  "Display the detected project context, available commands for that context, and other detected contexts.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := showStatus(); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func showStatus() error {
	// Load configuration
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Merge plugin contexts into config
	pm := getPluginManager()
	pluginContexts := pm.GetContexts()
	for ctxName, ctxConfig := range pluginContexts {
		if _, exists := cfg.Contexts[ctxName]; !exists {
			cfg.Contexts[ctxName] = ctxConfig
		}
	}

	// Detect all contexts
	var detectedContexts []string
	var activeContext string

	if forceCtx != "" {
		activeContext = forceCtx
		fmt.Printf("Context: %s (forced)\n", activeContext)
	} else {
		// Try plugin-based detection first
		pluginCtx, pluginName, foundByPlugin := pm.DetectContext(".")

		if foundByPlugin {
			activeContext = pluginCtx
			fmt.Printf("Context: %s (detected via plugin: %s)\n", activeContext, pluginName)
		} else {
			// Fall back to built-in detection
			detector := contextpkg.NewDetector()
			activeContext, err = detector.Detect(".")
			if err != nil {
				fmt.Println("Context: none detected")
			} else {
				fmt.Printf("Context: %s (detected)\n", activeContext)
			}
		}

		// Detect all possible contexts for "Other detected contexts"
		detectedContexts = detectAllContexts(pm)
	}

	fmt.Println()

	// Show available commands for the active context
	if activeContext != "" {
		reg := registry.New(cfg)
		commands, err := reg.ListCommands(activeContext)
		if err != nil {
			fmt.Printf("Error listing commands: %v\n", err)
		} else if len(commands) > 0 {
			fmt.Printf("Available commands in '%s' context:\n", activeContext)
			
			// Sort commands alphabetically
			sort.Strings(commands)
			
			// Get descriptions if available
			contextConfig, exists := cfg.Contexts[activeContext]
			if exists {
				for _, cmdName := range commands {
					desc := contextConfig.Descriptions[cmdName]
					cmd := contextConfig.Commands[cmdName]
					
					if desc != "" {
						fmt.Printf("  %-15s %s\n", cmdName, desc)
					} else {
						// Show the actual command if no description
						fmt.Printf("  %-15s → %s\n", cmdName, cmd)
					}
				}
			} else {
				// No config found, just list commands
				for _, cmdName := range commands {
					fmt.Printf("  %s\n", cmdName)
				}
			}
		} else {
			fmt.Printf("No commands available in '%s' context\n", activeContext)
		}
	}

	// Show other detected contexts
	if len(detectedContexts) > 1 {
		fmt.Println()
		fmt.Println("Other detected contexts:")
		for _, ctx := range detectedContexts {
			if ctx != activeContext {
				fmt.Printf("  %s\n", ctx)
			}
		}
	}

	// Show configuration file being used
	if cfgFile != "" {
		fmt.Println()
		fmt.Printf("Config file: %s\n", cfgFile)
	}

	return nil
}

// detectAllContexts returns all contexts that could be detected in the current directory
func detectAllContexts(pm *plugin.PluginManager) []string {
	var contexts []string
	seen := make(map[string]bool)

	// Check plugin contexts
	pluginContexts := pm.DetectAllContexts(".")
	for _, ctx := range pluginContexts {
		if !seen[ctx] {
			contexts = append(contexts, ctx)
			seen[ctx] = true
		}
	}

	// Check built-in contexts
	detector := contextpkg.NewDetector()
	builtinMarkers := map[string][]string{
		"node":   {"package.json"},
		"go":     {"go.mod"},
		"python": {"pyproject.toml", "requirements.txt", "setup.py"},
		"rust":   {"Cargo.toml"},
		"make":   {"Makefile"},
	}

	for ctx, markers := range builtinMarkers {
		if seen[ctx] {
			continue
		}
		for _, marker := range markers {
			if detector.FileExists(marker) {
				contexts = append(contexts, ctx)
				seen[ctx] = true
				break
			}
		}
	}

	sort.Strings(contexts)
	return contexts
}
