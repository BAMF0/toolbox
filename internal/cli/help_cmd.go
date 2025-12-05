package cli

import (
	"fmt"
	"sort"

	"github.com/bamf0/toolbox/internal/config"
	contextpkg "github.com/bamf0/toolbox/internal/context"
	"github.com/spf13/cobra"
)

var helpCmd = &cobra.Command{
	Use:   "help [command]",
	Short: "Show help for a command",
	Long: `Display help information for a specific command in the current or specified context.

Examples:
  tb help gbranch                    # Show help for gbranch in current context
  tb help --context ubuntu-packaging gbranch  # Show help for gbranch in ubuntu-packaging context
  tb help                            # Show general help`,
	RunE: showHelp,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		// Return all available commands from current/forced context
		return getDynamicCommandCompletions(toComplete), cobra.ShellCompDirectiveNoFileComp
	},
}

func init() {
	rootCmd.AddCommand(helpCmd)
}

func showHelp(cmd *cobra.Command, args []string) error {
	// If no command specified, show root help
	if len(args) == 0 {
		return rootCmd.Help()
	}

	commandName := args[0]

	// Load configuration
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Merge plugin contexts
	pm := getPluginManager()
	pluginContexts := pm.GetContexts()
	for ctxName, ctxConfig := range pluginContexts {
		if _, exists := cfg.Contexts[ctxName]; !exists {
			cfg.Contexts[ctxName] = ctxConfig
		}
	}

	// Detect or use forced context
	var detectedCtx string
	if forceCtx != "" {
		detectedCtx = forceCtx
	} else {
		pluginCtx, _, foundByPlugin := pm.DetectContext(".")
		if foundByPlugin {
			detectedCtx = pluginCtx
		} else {
			detector := contextpkg.NewDetector()
			detectedCtx, err = detector.Detect(".")
			if err != nil {
				// If we can't detect context, show all contexts where this command exists
				return showCommandInAllContexts(commandName, cfg)
			}
		}
	}

	// Find the command in the detected context
	ctxConfig, exists := cfg.Contexts[detectedCtx]
	if !exists {
		return fmt.Errorf("context '%s' not found", detectedCtx)
	}

	cmdString, cmdExists := ctxConfig.Commands[commandName]
	if !cmdExists {
		// Check if command exists in other contexts
		return showCommandInAllContexts(commandName, cfg)
	}

	// Display help
	fmt.Printf("Command: %s\n", commandName)
	fmt.Printf("Context: %s\n\n", detectedCtx)

	if description, hasDesc := ctxConfig.Descriptions[commandName]; hasDesc {
		fmt.Printf("Description:\n  %s\n\n", description)
	}

	fmt.Printf("Executes:\n  %s\n", cmdString)

	return nil
}

// showCommandInAllContexts shows where a command exists across all contexts
func showCommandInAllContexts(commandName string, cfg *config.Config) error {
	var foundContexts []string

	for ctxName, ctxConfig := range cfg.Contexts {
		if _, exists := ctxConfig.Commands[commandName]; exists {
			foundContexts = append(foundContexts, ctxName)
		}
	}

	if len(foundContexts) == 0 {
		return fmt.Errorf("command '%s' not found in any context", commandName)
	}

	fmt.Printf("Command '%s' is available in the following contexts:\n\n", commandName)

	sort.Strings(foundContexts)
	for _, ctxName := range foundContexts {
		ctxConfig := cfg.Contexts[ctxName]
		cmdString := ctxConfig.Commands[commandName]

		fmt.Printf("Context: %s\n", ctxName)
		if desc, hasDesc := ctxConfig.Descriptions[commandName]; hasDesc {
			fmt.Printf("  Description: %s\n", desc)
		}
		fmt.Printf("  Executes: %s\n\n", cmdString)
	}

	fmt.Printf("Use 'tb --context <context> %s' to run in a specific context.\n", commandName)

	return nil
}
