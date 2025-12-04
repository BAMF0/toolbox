package registry

import (
	"fmt"

	"github.com/bamf0/toolbox/internal/config"
)

// Registry manages command lookups across contexts
type Registry struct {
	config *config.Config
}

// New creates a new command registry
func New(cfg *config.Config) *Registry {
	return &Registry{
		config: cfg,
	}
}

// GetCommand retrieves the full command for a given context and command name
func (r *Registry) GetCommand(context, commandName string) (string, error) {
	// Check if context exists
	ctxConfig, exists := r.config.Contexts[context]
	if !exists {
		return "", fmt.Errorf("unknown context '%s'", context)
	}

	// Check if command exists in context
	fullCommand, exists := ctxConfig.Commands[commandName]
	if !exists {
		return "", fmt.Errorf("command '%s' not defined in context '%s'", commandName, context)
	}

	return fullCommand, nil
}

// ListCommands returns all available commands for a context
func (r *Registry) ListCommands(context string) ([]string, error) {
	ctxConfig, exists := r.config.Contexts[context]
	if !exists {
		return nil, fmt.Errorf("unknown context '%s'", context)
	}

	commands := make([]string, 0, len(ctxConfig.Commands))
	for cmd := range ctxConfig.Commands {
		commands = append(commands, cmd)
	}

	return commands, nil
}

// ListContexts returns all available contexts
func (r *Registry) ListContexts() []string {
	contexts := make([]string, 0, len(r.config.Contexts))
	for ctx := range r.config.Contexts {
		contexts = append(contexts, ctx)
	}
	return contexts
}
