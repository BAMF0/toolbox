package cli

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/bamf0/toolbox/internal/plugin"
	"github.com/spf13/cobra"
)

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage ToolBox plugins",
	Long: `Manage plugins that extend ToolBox with additional contexts and commands.

Examples:
  tb plugin list              List all installed plugins
  tb plugin info docker       Show details about a specific plugin
  tb plugin contexts          List all contexts provided by plugins`,
}

var pluginListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all installed plugins",
	RunE:  runPluginList,
}

var pluginInfoCmd = &cobra.Command{
	Use:   "info [plugin-name]",
	Short: "Show detailed information about a plugin",
	Args:  cobra.ExactArgs(1),
	RunE:  runPluginInfo,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return getPluginNameCompletions(toComplete), cobra.ShellCompDirectiveNoFileComp
	},
}

var pluginContextsCmd = &cobra.Command{
	Use:   "contexts",
	Short: "List all contexts provided by plugins",
	RunE:  runPluginContexts,
}

func init() {
	rootCmd.AddCommand(pluginCmd)
	pluginCmd.AddCommand(pluginListCmd)
	pluginCmd.AddCommand(pluginInfoCmd)
	pluginCmd.AddCommand(pluginContextsCmd)
}

// getPluginManager returns a configured plugin manager with built-in plugins
func getPluginManager() *plugin.PluginManager {
	pm := plugin.NewPluginManager("")

	// Register built-in plugins
	pm.RegisterPlugin(plugin.NewDockerPlugin())
	pm.RegisterPlugin(plugin.NewKubernetesPlugin())
	pm.RegisterPlugin(plugin.NewUbuntuPlugin())

	return pm
}

func runPluginList(cmd *cobra.Command, args []string) error {
	pm := getPluginManager()
	metadata := pm.GetMetadata()

	if len(metadata) == 0 {
		fmt.Println("No plugins installed")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "NAME\tVERSION\tCONTEXTS\tSTATUS")
	fmt.Fprintln(w, "────\t───────\t────────\t──────")

	for _, meta := range metadata {
		status := "enabled"
		if !meta.Enabled {
			status = "disabled"
		}

		contextsStr := fmt.Sprintf("%d", meta.ContextCount)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			meta.Name,
			meta.Version,
			contextsStr,
			status,
		)
	}

	w.Flush()
	return nil
}

func runPluginInfo(cmd *cobra.Command, args []string) error {
	pm := getPluginManager()
	pluginName := args[0]

	metadata := pm.GetMetadata()
	meta, exists := metadata[pluginName]
	if !exists {
		return fmt.Errorf("plugin %q not found", pluginName)
	}

	fmt.Printf("Plugin: %s\n", meta.Name)
	fmt.Printf("Version: %s\n", meta.Version)
	fmt.Printf("Status: %s\n", boolToStatus(meta.Enabled))
	fmt.Printf("Contexts: %d\n\n", meta.ContextCount)

	if len(meta.Contexts) > 0 {
		fmt.Println("Provided Contexts:")
		for _, ctx := range meta.Contexts {
			fmt.Printf("  - %s\n", ctx)
		}
	}

	return nil
}

func runPluginContexts(cmd *cobra.Command, args []string) error {
	pm := getPluginManager()
	contexts := pm.GetContexts()

	if len(contexts) == 0 {
		fmt.Println("No plugin contexts available")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "CONTEXT\tCOMMANDS")
	fmt.Fprintln(w, "───────\t────────")

	for ctxName, ctxConfig := range contexts {
		commandCount := len(ctxConfig.Commands)
		fmt.Fprintf(w, "%s\t%d\n", ctxName, commandCount)
	}

	w.Flush()
	return nil
}

func boolToStatus(enabled bool) string {
	if enabled {
		return "enabled"
	}
	return "disabled"
}

// getPluginNameCompletions returns plugin names for autocomplete
func getPluginNameCompletions(toComplete string) []string {
	pm := getPluginManager()
	metadata := pm.GetMetadata()

	var suggestions []string
	for name := range metadata {
		if strings.HasPrefix(name, toComplete) {
			suggestions = append(suggestions, name)
		}
	}
	return suggestions
}
