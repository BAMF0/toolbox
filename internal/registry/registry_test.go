package registry

import (
	"testing"

	"github.com/bamf0/toolbox/internal/config"
)

// TestRegistry_GetCommand tests command retrieval
func TestRegistry_GetCommand(t *testing.T) {
	cfg := &config.Config{
		Contexts: map[string]config.ContextConfig{
			"test": {
				Commands: map[string]string{
					"build": "make all",
					"test":  "make test",
					"run":   "./app",
				},
			},
		},
	}

	reg := New(cfg)

	tests := []struct {
		name        string
		context     string
		commandName string
		wantCommand string
		wantErr     bool
	}{
		{
			name:        "valid command",
			context:     "test",
			commandName: "build",
			wantCommand: "make all",
			wantErr:     false,
		},
		{
			name:        "another valid command",
			context:     "test",
			commandName: "test",
			wantCommand: "make test",
			wantErr:     false,
		},
		{
			name:        "unknown command in valid context",
			context:     "test",
			commandName: "deploy",
			wantCommand: "",
			wantErr:     true,
		},
		{
			name:        "unknown context",
			context:     "nonexistent",
			commandName: "build",
			wantCommand: "",
			wantErr:     true,
		},
		{
			name:        "empty command name",
			context:     "test",
			commandName: "",
			wantCommand: "",
			wantErr:     true,
		},
		{
			name:        "empty context",
			context:     "",
			commandName: "build",
			wantCommand: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := reg.GetCommand(tt.context, tt.commandName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetCommand() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("GetCommand() unexpected error: %v", err)
				}
				if cmd != tt.wantCommand {
					t.Errorf("GetCommand() = %q, want %q", cmd, tt.wantCommand)
				}
			}
		})
	}
}

// TestRegistry_ListCommands tests listing all commands in a context
func TestRegistry_ListCommands(t *testing.T) {
	cfg := &config.Config{
		Contexts: map[string]config.ContextConfig{
			"test": {
				Commands: map[string]string{
					"build": "make all",
					"test":  "make test",
					"run":   "./app",
				},
			},
			"empty": {
				Commands: map[string]string{},
			},
		},
	}

	reg := New(cfg)

	tests := []struct {
		name         string
		context      string
		wantCommands int
		wantErr      bool
	}{
		{
			name:         "context with commands",
			context:      "test",
			wantCommands: 3,
			wantErr:      false,
		},
		{
			name:         "empty context",
			context:      "empty",
			wantCommands: 0,
			wantErr:      false,
		},
		{
			name:         "unknown context",
			context:      "nonexistent",
			wantCommands: 0,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commands, err := reg.ListCommands(tt.context)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ListCommands() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("ListCommands() unexpected error: %v", err)
				}
				if len(commands) != tt.wantCommands {
					t.Errorf("ListCommands() returned %d commands, want %d", len(commands), tt.wantCommands)
				}
			}
		})
	}
}

// TestRegistry_ListContexts tests listing all available contexts
func TestRegistry_ListContexts(t *testing.T) {
	cfg := &config.Config{
		Contexts: map[string]config.ContextConfig{
			"node":   {Commands: map[string]string{"build": "npm run build"}},
			"go":     {Commands: map[string]string{"build": "go build"}},
			"python": {Commands: map[string]string{"test": "pytest"}},
		},
	}

	reg := New(cfg)
	contexts := reg.ListContexts()

	if len(contexts) != 3 {
		t.Errorf("ListContexts() returned %d contexts, want 3", len(contexts))
	}

	// Verify all contexts are present (order doesn't matter)
	contextMap := make(map[string]bool)
	for _, ctx := range contexts {
		contextMap[ctx] = true
	}

	expectedContexts := []string{"node", "go", "python"}
	for _, expected := range expectedContexts {
		if !contextMap[expected] {
			t.Errorf("ListContexts() missing expected context %q", expected)
		}
	}
}

// TestRegistry_EmptyConfig tests registry with empty configuration
func TestRegistry_EmptyConfig(t *testing.T) {
	cfg := &config.Config{
		Contexts: map[string]config.ContextConfig{},
	}

	reg := New(cfg)

	// Should return empty list, not error
	contexts := reg.ListContexts()
	if len(contexts) != 0 {
		t.Errorf("ListContexts() on empty config returned %d contexts, want 0", len(contexts))
	}

	// Should error on GetCommand
	_, err := reg.GetCommand("any", "build")
	if err == nil {
		t.Error("GetCommand() on empty config expected error, got nil")
	}
}

// TestRegistry_NilConfig tests registry with nil configuration (defensive programming)
func TestRegistry_NilConfig(t *testing.T) {
	// This tests defensive programming - in practice this shouldn't happen
	// but we want to ensure it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("New() with nil config panicked: %v", r)
		}
	}()

	reg := New(nil)

	// Should not panic, may return error or empty results
	_ = reg.ListContexts()
}

// TestRegistry_GetCommand_WithDefaults tests using default configuration
func TestRegistry_GetCommand_WithDefaults(t *testing.T) {
	// Use actual default config
	cfg := &config.Config{
		Contexts: map[string]config.ContextConfig{
			"node": {
				Commands: map[string]string{
					"build": "npm run build",
					"test":  "npm test",
				},
			},
		},
	}

	reg := New(cfg)

	cmd, err := reg.GetCommand("node", "build")
	if err != nil {
		t.Fatalf("GetCommand() unexpected error: %v", err)
	}

	if cmd != "npm run build" {
		t.Errorf("GetCommand() = %q, want %q", cmd, "npm run build")
	}
}

// TestRegistry_SpecialCharactersInCommands tests handling of commands with special characters
func TestRegistry_SpecialCharactersInCommands(t *testing.T) {
	cfg := &config.Config{
		Contexts: map[string]config.ContextConfig{
			"test": {
				Commands: map[string]string{
					"complex": "npm run build && npm run test | tee output.log",
					"quoted":  "echo 'hello world'",
					"vars":    "GOOS=linux go build",
				},
			},
		},
	}

	reg := New(cfg)

	tests := []struct {
		name        string
		commandName string
		wantCommand string
	}{
		{
			name:        "command with pipes and redirects",
			commandName: "complex",
			wantCommand: "npm run build && npm run test | tee output.log",
		},
		{
			name:        "command with quotes",
			commandName: "quoted",
			wantCommand: "echo 'hello world'",
		},
		{
			name:        "command with environment variables",
			commandName: "vars",
			wantCommand: "GOOS=linux go build",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := reg.GetCommand("test", tt.commandName)
			if err != nil {
				t.Fatalf("GetCommand() unexpected error: %v", err)
			}
			if cmd != tt.wantCommand {
				t.Errorf("GetCommand() = %q, want %q", cmd, tt.wantCommand)
			}
		})
	}
}

// Benchmark tests
func BenchmarkRegistry_GetCommand(b *testing.B) {
	cfg := &config.Config{
		Contexts: map[string]config.ContextConfig{
			"test": {
				Commands: map[string]string{
					"build": "make all",
					"test":  "make test",
				},
			},
		},
	}

	reg := New(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = reg.GetCommand("test", "build")
	}
}

func BenchmarkRegistry_ListCommands(b *testing.B) {
	cfg := &config.Config{
		Contexts: map[string]config.ContextConfig{
			"test": {
				Commands: map[string]string{
					"build":   "make all",
					"test":    "make test",
					"deploy":  "./deploy.sh",
					"clean":   "make clean",
					"install": "make install",
				},
			},
		},
	}

	reg := New(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = reg.ListCommands("test")
	}
}

func BenchmarkRegistry_ListContexts(b *testing.B) {
	cfg := &config.Config{
		Contexts: map[string]config.ContextConfig{
			"node":   {Commands: map[string]string{"build": "npm run build"}},
			"go":     {Commands: map[string]string{"build": "go build"}},
			"python": {Commands: map[string]string{"test": "pytest"}},
			"rust":   {Commands: map[string]string{"build": "cargo build"}},
			"java":   {Commands: map[string]string{"build": "mvn package"}},
		},
	}

	reg := New(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = reg.ListContexts()
	}
}
