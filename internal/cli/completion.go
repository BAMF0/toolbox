package cli

import (
	"fmt"
	"strings"

	"github.com/bamf0/toolbox/internal/config"
	contextpkg "github.com/bamf0/toolbox/internal/context"
	"github.com/spf13/cobra"
)

// setupCompletion configures custom completion for the root command
func setupCompletion() {
	// Add custom completion for the root command to suggest dynamic commands
	rootCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		// Get all available commands from current context
		suggestions := getDynamicCommandCompletions(toComplete)
		return suggestions, cobra.ShellCompDirectiveNoFileComp
	}

	// Add completion for --context flag
	rootCmd.RegisterFlagCompletionFunc("context", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return getContextCompletions(toComplete), cobra.ShellCompDirectiveNoFileComp
	})

	// Add completion for --config flag
	rootCmd.RegisterFlagCompletionFunc("config", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveDefault // Show files
	})
}

// getDynamicCommandCompletions returns command suggestions based on current context
func getDynamicCommandCompletions(toComplete string) []string {
	var suggestions []string

	// Try to detect context
	var detectedCtx string

	// Try plugin-based detection first
	pm := getPluginManager()
	pluginCtx, _, foundByPlugin := pm.DetectContext(".")

	if foundByPlugin {
		detectedCtx = pluginCtx
	} else {
		// Fall back to built-in detection
		detector := contextpkg.NewDetector()
		ctx, err := detector.Detect(".")
		if err == nil {
			detectedCtx = ctx
		}
	}

	if detectedCtx != "" {
		// Load config and get commands for detected context
		cfg, err := config.Load("")
		if err == nil {
			// Merge plugin contexts
			pluginContexts := pm.GetContexts()
			for ctxName, ctxConfig := range pluginContexts {
				if _, exists := cfg.Contexts[ctxName]; !exists {
					cfg.Contexts[ctxName] = ctxConfig
				}
			}

			// Get commands from detected context
			if ctxConfig, exists := cfg.Contexts[detectedCtx]; exists {
				for cmdName := range ctxConfig.Commands {
					if strings.HasPrefix(cmdName, toComplete) {
						suggestions = append(suggestions, cmdName)
					}
				}
			}
		}
	}

	// If no context-specific suggestions, add common commands
	if len(suggestions) == 0 {
		commonCommands := []string{"build", "test", "run", "deploy", "lint", "clean"}
		for _, cmd := range commonCommands {
			if strings.HasPrefix(cmd, toComplete) {
				suggestions = append(suggestions, cmd)
			}
		}
	}

	return suggestions
}

// getContextCompletions returns all available contexts for completion
func getContextCompletions(toComplete string) []string {
	var suggestions []string

	// Load config
	cfg, err := config.Load("")
	if err != nil {
		return suggestions
	}

	// Merge plugin contexts
	pm := getPluginManager()
	pluginContexts := pm.GetContexts()
	for ctxName, ctxConfig := range pluginContexts {
		if _, exists := cfg.Contexts[ctxName]; !exists {
			cfg.Contexts[ctxName] = ctxConfig
		}
	}

	// Get all context names
	for ctxName := range cfg.Contexts {
		if strings.HasPrefix(ctxName, toComplete) {
			suggestions = append(suggestions, ctxName)
		}
	}

	return suggestions
}

// Add enhanced completion command with instructions
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: `Generate shell completion script for ToolBox.

To load completions:

Bash:
  # Linux
  $ tb completion bash > /etc/bash_completion.d/tb
  
  # macOS
  $ tb completion bash > $(brew --prefix)/etc/bash_completion.d/tb

Zsh:
  # If shell completion is not already enabled:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc
  
  # Add completion script:
  $ tb completion zsh > "${fpath[1]}/_tb"
  
  # Or add to a custom completion directory:
  $ mkdir -p ~/.zsh/completion
  $ tb completion zsh > ~/.zsh/completion/_tb
  $ echo 'fpath=(~/.zsh/completion $fpath)' >> ~/.zshrc

Fish:
  $ tb completion fish > ~/.config/fish/completions/tb.fish

PowerShell:
  $ tb completion powershell | Out-String | Invoke-Expression
  
  # To persist, add to your PowerShell profile:
  $ tb completion powershell >> $PROFILE

After installing, you may need to restart your shell or run:
  $ source ~/.bashrc  # or ~/.zshrc
`,
	ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
	Args:      cobra.ExactArgs(1),
	RunE:      runCompletion,
}

func init() {
	rootCmd.AddCommand(completionCmd)
	setupCompletion()
}

func runCompletion(cmd *cobra.Command, args []string) error {
	shell := args[0]

	switch shell {
	case "bash":
		return rootCmd.GenBashCompletionV2(cmd.OutOrStdout(), true)
	case "zsh":
		return rootCmd.GenZshCompletion(cmd.OutOrStdout())
	case "fish":
		return rootCmd.GenFishCompletion(cmd.OutOrStdout(), true)
	case "powershell":
		return rootCmd.GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
	default:
		return fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish, powershell)", shell)
	}
}
