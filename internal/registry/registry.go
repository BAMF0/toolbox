// Package registry manages command lookups and provides a safe interface
// for accessing context-specific command mappings.
package registry

import (
	"fmt"

	"github.com/bamf0/toolbox/internal/config"
)

// Registry manages command lookups across contexts
type Registry struct {
	config *config.Config
}

// New creates a new command registry.
// If cfg is nil, operations will return appropriate errors rather than panicking.
func New(cfg *config.Config) *Registry {
	return &Registry{
		config: cfg,
	}
}

// GetCommand retrieves the full command for a given context and command name.
// Returns an error if the config is nil, context doesn't exist, or command is not found.
func (r *Registry) GetCommand(context, commandName string) (string, error) {
	if r.config == nil || r.config.Contexts == nil {
		return "", fmt.Errorf("registry not properly initialized")
	}

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

// ListCommands returns all available commands for a context.
// Returns an error if the config is nil or context doesn't exist.
func (r *Registry) ListCommands(context string) ([]string, error) {
	if r.config == nil || r.config.Contexts == nil {
		return nil, fmt.Errorf("registry not properly initialized")
	}

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

// ListContexts returns all available contexts.
// Returns an empty slice if the config is nil.
func (r *Registry) ListContexts() []string {
	if r.config == nil || r.config.Contexts == nil {
		return []string{}
	}

	contexts := make([]string, 0, len(r.config.Contexts))
	for ctx := range r.config.Contexts {
		contexts = append(contexts, ctx)
	}
	return contexts
}
