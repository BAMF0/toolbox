package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/bamf0/toolbox/internal/config"
	"github.com/bamf0/toolbox/internal/context"
	"github.com/bamf0/toolbox/internal/registry"
	"github.com/spf13/cobra"
)

var (
	cfgFile     string
	forceCtx    string
	dryRun      bool
	verbose     bool
)

var rootCmd = &cobra.Command{
	Use:   "tb",
	Short: "ToolBox - Context-aware command aliasing",
	Long: `ToolBox (tb) provides intelligent command shortcuts based on your project context.
	
Define simple commands like 'tb build' or 'tb test' that automatically expand
to the correct commands for your current project type (Node.js, Go, Python, etc.).`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: .toolbox.yaml or ~/.toolbox/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&forceCtx, "context", "", "force a specific context (node, go, python, etc.)")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "print command without executing")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Dynamic command handler - intercepts unknown commands
	rootCmd.RunE = handleDynamicCommand
}

// handleDynamicCommand processes commands not explicitly defined (build, test, etc.)
func handleDynamicCommand(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}

	commandName := args[0]
	commandArgs := args[1:]

	// Load configuration
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Detect context (or use forced context)
	var detectedCtx string
	if forceCtx != "" {
		detectedCtx = forceCtx
		if verbose {
			fmt.Printf("Using forced context: %s\n", detectedCtx)
		}
	} else {
		detector := context.NewDetector()
		detectedCtx, err = detector.Detect(".")
		if err != nil {
			return fmt.Errorf("failed to detect context: %w", err)
		}
		if verbose {
			fmt.Printf("Detected context: %s\n", detectedCtx)
		}
	}

	// Get command from registry
	reg := registry.New(cfg)
	fullCommand, err := reg.GetCommand(detectedCtx, commandName)
	if err != nil {
		return fmt.Errorf("command '%s' not found in context '%s': %w", commandName, detectedCtx, err)
	}

	// Append any additional arguments
	if len(commandArgs) > 0 {
		fullCommand = fullCommand + " " + strings.Join(commandArgs, " ")
	}

	if dryRun || verbose {
		fmt.Printf("Context: %s\n", detectedCtx)
		fmt.Printf("Command: %s\n", fullCommand)
		if dryRun {
			return nil
		}
	}

	// Execute the command
	return executeCommand(fullCommand)
}

// executeCommand runs the shell command
func executeCommand(command string) error {
	// Determine shell based on OS
	shell := "sh"
	shellArg := "-c"
	if _, err := exec.LookPath("bash"); err == nil {
		shell = "bash"
	}

	// On Windows, use cmd
	if os.PathSeparator == '\\' {
		shell = "cmd"
		shellArg = "/C"
	}

	cmd := exec.Command(shell, shellArg, command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
