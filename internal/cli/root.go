package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/bamf0/toolbox/internal/config"
	contextpkg "github.com/bamf0/toolbox/internal/context"
	"github.com/bamf0/toolbox/internal/registry"
	"github.com/spf13/cobra"
)

const (
	// DefaultCommandTimeout is the maximum time a command can run
	DefaultCommandTimeout = 10 * time.Minute

	// MaxArgumentLength limits individual argument size to prevent memory exhaustion
	MaxArgumentLength = 8192

	// MaxArgumentCount limits total number of arguments
	MaxArgumentCount = 100
)

var (
	// Version is the current version of ToolBox
	// This can be overridden at build time with -ldflags
	Version = "0.1.0"
	// GitCommit is set via ldflags during build
	GitCommit = "unknown"
	// BuildTime is set via ldflags during build
	BuildTime = "unknown"
)

var (
	cfgFile        string
	forceCtx       string
	dryRun         bool
	verbose        bool
	versionFlag    bool
	commandTimeout time.Duration
)

var rootCmd = &cobra.Command{
	Use:   "tb",
	Short: "ToolBox - Context-aware command aliasing",
	Long: `ToolBox (tb) provides intelligent command shortcuts based on your project context.
	
Define simple commands like 'tb build' or 'tb test' that automatically expand
to the correct commands for your current project type (Node.js, Go, Python, etc.).`,
	SilenceUsage:  true,
	SilenceErrors: true,
	// Disable Cobra's default help command to avoid duplication
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: false,
	},
	// Disable flag parsing - we'll handle it manually
	DisableFlagParsing: true,
}

// customHelp provides enhanced help output with context-specific commands
func customHelp(cmd *cobra.Command, args []string) {
	fmt.Println(cmd.Long)
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  tb [flags]")
	fmt.Println("  tb [command]")
	fmt.Println()
	
	// Show built-in commands
	fmt.Println("Available Commands:")
	for _, subCmd := range cmd.Commands() {
		if !subCmd.Hidden {
			fmt.Printf("  %-12s %s\n", subCmd.Name(), subCmd.Short)
		}
	}
	fmt.Println()
	
	// Try to detect context and show context-specific commands
	showContextCommands()
	
	// Show flags
	fmt.Println("Flags:")
	fmt.Println("      --config string      config file (default: .toolbox.yaml or ~/.toolbox/config.yaml)")
	fmt.Println("      --context string     force a specific context (node, go, python, etc.)")
	fmt.Println("      --dry-run            print command without executing")
	fmt.Println("  -h, --help               help for tb")
	fmt.Println("      --timeout duration   command execution timeout (default 10m0s)")
	fmt.Println("      --verbose            verbose output")
	fmt.Println("      --version            show version information")
	fmt.Println()
	fmt.Println("Use \"tb [command] --help\" for more information about a command.")
	fmt.Println("Use \"tb status\" to see current context and available commands.")
}

// showContextCommands displays commands available in the current context
func showContextCommands() {
	// Load configuration
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return // Silently skip if config can't be loaded
	}

	// Merge plugin contexts
	pm := getPluginManager()
	pluginContexts := pm.GetContexts()
	for ctxName, ctxConfig := range pluginContexts {
		if _, exists := cfg.Contexts[ctxName]; !exists {
			cfg.Contexts[ctxName] = ctxConfig
		}
	}

	// Detect context
	var activeContext string
	if forceCtx != "" {
		activeContext = forceCtx
	} else {
		pluginCtx, _, foundByPlugin := pm.DetectContext(".")
		if foundByPlugin {
			activeContext = pluginCtx
		} else {
			detector := contextpkg.NewDetector()
			activeContext, err = detector.Detect(".")
			if err != nil {
				return // No context detected
			}
		}
	}

	// Get commands for the active context
	reg := registry.New(cfg)
	commands, err := reg.ListCommands(activeContext)
	if err != nil {
		return // Silently skip if error
	}
	
	if len(commands) == 0 {
		return
	}

	fmt.Printf("Context-Specific Commands (%s):\n", activeContext)
	
	// Sort commands
	sort.Strings(commands)
	
	// Get descriptions if available
	contextConfig, exists := cfg.Contexts[activeContext]
	if exists {
		for _, cmdName := range commands {
			desc := contextConfig.Descriptions[cmdName]
			if desc != "" {
				fmt.Printf("  %-12s %s\n", cmdName, desc)
			} else {
				fmt.Printf("  %-12s\n", cmdName)
			}
		}
	} else {
		for _, cmdName := range commands {
			fmt.Printf("  %-12s\n", cmdName)
		}
	}
	fmt.Println()
}

// Execute runs the root command and returns any error encountered.
// This is the main entry point for the CLI application.
func Execute() error {
	// Pre-process args to handle special cases
	args := os.Args[1:]
	
	// Handle -- separator (everything after -- is passed to the command)
	separatorIdx := -1
	for i, arg := range args {
		if arg == "--" {
			separatorIdx = i
			break
		}
	}
	
	if separatorIdx >= 0 {
		// Split args at --
		// Before: tb flags and command
		// After: arguments for the underlying command
		beforeSep := args[:separatorIdx]
		afterSep := args[separatorIdx+1:]
		
		// Find the command name (first non-flag arg before --)
		var cmdName string
		tbFlags := []string{}
		for _, arg := range beforeSep {
			if !strings.HasPrefix(arg, "-") && cmdName == "" {
				cmdName = arg
			} else {
				tbFlags = append(tbFlags, arg)
			}
		}
		
		if cmdName != "" {
			// Reconstruct: tb [flags] command -- [command args]
			// becomes: tb [flags] command [command args]
			os.Args = append([]string{os.Args[0]}, tbFlags...)
			os.Args = append(os.Args, cmdName)
			os.Args = append(os.Args, afterSep...)
		}
	}
	
	// Pre-process args to handle --help on dynamic commands
	args = os.Args[1:]
	if len(args) >= 2 {
		// Check if this looks like a dynamic command with --help
		// (not a known subcommand like "plugin", "completion", "help", "status")
		potentialCmd := args[0]
		knownCommands := map[string]bool{
			"plugin":     true,
			"completion": true,
			"help":       true,
			"status":     true,
		}

		if !knownCommands[potentialCmd] {
			// This might be a dynamic command
			for _, arg := range args[1:] {
				if arg == "--help" || arg == "-h" {
					// Redirect to: tb help <command>
					os.Args = []string{os.Args[0], "help", potentialCmd}
					break
				}
			}
		}
	}

	return rootCmd.Execute()
}

// GetVersion returns the current version of ToolBox
func GetVersion() string {
	return Version
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: .toolbox.yaml or ~/.toolbox/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&forceCtx, "context", "", "force a specific context (node, go, python, etc.)")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "print command without executing")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().DurationVar(&commandTimeout, "timeout", DefaultCommandTimeout, "command execution timeout")
	rootCmd.Flags().BoolVar(&versionFlag, "version", false, "show version information")

	// Set custom help function
	rootCmd.SetHelpFunc(customHelp)
	
	// Disable Cobra's auto-generated help command (we have our own)
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	// Allow unknown commands to be handled dynamically
	rootCmd.Args = cobra.ArbitraryArgs

	// Dynamic command handler - intercepts unknown commands
	rootCmd.RunE = handleDynamicCommand
}

// handleDynamicCommand processes commands not explicitly defined (build, test, etc.)
func handleDynamicCommand(cmd *cobra.Command, args []string) error {
	// Manually parse flags since we disabled automatic parsing
	// Strategy: Parse tb flags until we hit a non-flag argument (the command),
	// then everything after is passed to the underlying command
	var commandName string
	var commandArgs []string
	foundCommand := false
	 
	for i := 0; i < len(args); i++ {
		arg := args[i]
		
		// Once we've found the command, everything else is a command argument
		if foundCommand {
			commandArgs = append(commandArgs, arg)
			continue
		}
		
		// Handle --version
		if arg == "--version" {
			fmt.Printf("ToolBox (tb) version %s\n", Version)
			if GitCommit != "unknown" {
				fmt.Printf("Git commit: %s\n", GitCommit)
			}
			if BuildTime != "unknown" {
				fmt.Printf("Built: %s\n", BuildTime)
			}
			return nil
		}
		
		// Handle --help or -h
		if arg == "--help" || arg == "-h" {
			return cmd.Help()
		}
		
		// Handle --config
		if arg == "--config" && i+1 < len(args) {
			cfgFile = args[i+1]
			i++ // skip next arg
			continue
		}
		
		// Handle --context
		if arg == "--context" && i+1 < len(args) {
			forceCtx = args[i+1]
			i++ // skip next arg
			continue
		}
		
		// Handle --dry-run
		if arg == "--dry-run" {
			dryRun = true
			continue
		}
		
		// Handle --verbose (but not -v, to avoid conflicts)
		if arg == "--verbose" {
			verbose = true
			continue
		}
		
		// Handle --timeout
		if arg == "--timeout" && i+1 < len(args) {
			var err error
			commandTimeout, err = time.ParseDuration(args[i+1])
			if err != nil {
				return fmt.Errorf("invalid timeout duration: %w", err)
			}
			i++ // skip next arg
			continue
		}
		
		// If it doesn't start with -, it's the command name
		if !strings.HasPrefix(arg, "-") {
			commandName = arg
			foundCommand = true
			continue
		}
		
		// Unknown flag - could be for the command, so treat as command name
		commandName = arg
		foundCommand = true
	}

	if commandName == "" {
		return cmd.Help()
	}

	// Check if user wants help for this command (anywhere in args)
	for _, arg := range commandArgs {
		if arg == "--help" || arg == "-h" {
			return showHelp(cmd, []string{commandName})
		}
	}

	// Validate arguments early
	if err := validateArguments(commandArgs); err != nil {
		return fmt.Errorf("invalid arguments: %w", err)
	}

	// Load configuration
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Merge plugin contexts into config
	pm := getPluginManager()
	pluginContexts := pm.GetContexts()
	for ctxName, ctxConfig := range pluginContexts {
		// Only add if not already in config (config takes precedence)
		if _, exists := cfg.Contexts[ctxName]; !exists {
			cfg.Contexts[ctxName] = ctxConfig
		}
	}

	// Detect context (or use forced context)
	var detectedCtx string
	if forceCtx != "" {
		detectedCtx = forceCtx
		if verbose {
			fmt.Printf("Using forced context: %s\n", detectedCtx)
		}
	} else {
		// Try plugin-based detection first
		pluginCtx, pluginName, foundByPlugin := pm.DetectContext(".")

		if foundByPlugin {
			detectedCtx = pluginCtx
			if verbose {
				fmt.Printf("Detected context: %s (via plugin: %s)\n", detectedCtx, pluginName)
			}
		} else {
			// Fall back to built-in detection
			detector := contextpkg.NewDetector()
			detectedCtx, err = detector.Detect(".")
			if err != nil {
				return fmt.Errorf("failed to detect context: %w", err)
			}
			if verbose {
				fmt.Printf("Detected context: %s\n", detectedCtx)
			}
		}
	}

	// Get command from registry
	reg := registry.New(cfg)
	baseCommand, err := reg.GetCommand(detectedCtx, commandName)
	if err != nil {
		return fmt.Errorf("command '%s' not found in context '%s': %w", commandName, detectedCtx, err)
	}

	if dryRun || verbose {
		fmt.Printf("Context: %s\n", detectedCtx)
		fmt.Printf("Base command: %s\n", baseCommand)
		if len(commandArgs) > 0 {
			fmt.Printf("Additional arguments: %s\n", strings.Join(commandArgs, " "))
		}
		if dryRun {
			return nil
		}
	}

	// Execute the command securely
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	return executeCommandSecure(ctx, baseCommand, commandArgs)
}

// validateArguments performs security validation on user-supplied arguments
func validateArguments(args []string) error {
	if len(args) > MaxArgumentCount {
		return fmt.Errorf("too many arguments (max: %d, got: %d)", MaxArgumentCount, len(args))
	}

	for i, arg := range args {
		if len(arg) > MaxArgumentLength {
			return fmt.Errorf("argument %d exceeds maximum length of %d bytes", i, MaxArgumentLength)
		}

		// Warn about potentially dangerous characters (informational only in this version)
		if containsDangerousPatterns(arg) && verbose {
			fmt.Fprintf(os.Stderr, "Warning: argument %d contains shell metacharacters: %q\n", i, arg)
		}
	}

	return nil
}

// containsDangerousPatterns checks for common shell injection characters
// This is informational; actual protection comes from not using a shell
func containsDangerousPatterns(s string) bool {
	dangerous := []string{";", "|", "&", "$", "`", "(", ")", "<", ">", "\n", "\r"}
	for _, pattern := range dangerous {
		if strings.Contains(s, pattern) {
			return true
		}
	}
	return false
}

// executeCommandSecure runs the command WITHOUT shell interpretation
// This is the primary defense against command injection
func executeCommandSecure(ctx context.Context, baseCommand string, userArgs []string) error {
	// Parse the base command into program and arguments
	// We split on whitespace, which handles simple cases like "npm run build"
	// For complex commands with pipes/redirects, those should be in shell scripts
	parts := strings.Fields(baseCommand)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	program := parts[0]
	baseArgs := parts[1:]

	// Combine base arguments with user-supplied arguments
	allArgs := append(baseArgs, userArgs...)

	// Validate that the program exists and is executable
	programPath, err := exec.LookPath(program)
	if err != nil {
		return fmt.Errorf("command not found: %s: %w", program, err)
	}

	if verbose {
		fmt.Printf("Executing: %s %s\n", programPath, strings.Join(allArgs, " "))
	}

	// Create command with explicit arguments (NO SHELL)
	cmd := exec.CommandContext(ctx, programPath, allArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ() // Explicitly set environment

	// Execute and handle errors with context
	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("command timed out after %v", commandTimeout)
		}
		// Preserve original error for debugging
		return fmt.Errorf("command failed: %w", err)
	}

	return nil
}

// executeCommandShellFallback is for commands that genuinely need shell features
// This should ONLY be used for trusted commands from config files, NEVER user input
// DEPRECATED: Use shell scripts in config instead
func executeCommandShellFallback(ctx context.Context, command string) error {
	// Determine shell based on OS
	shell := "/bin/sh"
	shellArg := "-c"

	// Try to use bash if available (better error handling)
	if bashPath, err := exec.LookPath("bash"); err == nil {
		shell = bashPath
	}

	// On Windows, use cmd
	if os.PathSeparator == '\\' {
		shell = "cmd"
		shellArg = "/C"
	}

	if verbose {
		fmt.Printf("Warning: Using shell execution: %s %s %q\n", shell, shellArg, command)
	}

	cmd := exec.CommandContext(ctx, shell, shellArg, command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("command timed out after %v", commandTimeout)
		}
		return fmt.Errorf("shell command failed: %w", err)
	}

	return nil
}
