package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestValidateConfigPath tests path validation security
func TestValidateConfigPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid relative yaml",
			path:    "config.yaml",
			wantErr: false,
		},
		{
			name:    "valid relative yml",
			path:    "myconfig.yml",
			wantErr: false,
		},
		{
			name:    "valid nested relative",
			path:    "configs/toolbox.yaml",
			wantErr: false,
		},
		{
			name:    "absolute path - should fail",
			path:    "/etc/passwd",
			wantErr: true,
			errMsg:  "absolute paths not allowed",
		},
		{
			name:    "directory traversal with ..",
			path:    "../../../etc/passwd",
			wantErr: true,
			errMsg:  "directory traversal",
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
			errMsg:  "empty path",
		},
		{
			name:    "non-yaml extension",
			path:    "config.txt",
			wantErr: true,
			errMsg:  "must have .yaml or .yml extension",
		},
		{
			name:    "no extension",
			path:    "config",
			wantErr: true,
			errMsg:  "must have .yaml or .yml extension",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfigPath(tt.path)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateConfigPath() expected error, got nil")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("validateConfigPath() error = %v, want error containing %q", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateConfigPath() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestLoadFromFile_SizeLimit tests file size enforcement
func TestLoadFromFile_SizeLimit(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		fileSize int64
		wantErr  bool
	}{
		{
			name:     "small valid file",
			fileSize: 1024,
			wantErr:  false,
		},
		{
			name:     "file at size limit",
			fileSize: MaxConfigFileSize,
			wantErr:  false,
		},
		{
			name:     "file exceeds size limit",
			fileSize: MaxConfigFileSize + 1,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file with specific size
			testFile := filepath.Join(tmpDir, "test.yaml")

			// Create valid YAML content
			validYAML := "contexts:\n  test:\n    commands:\n      build: echo test\n"

			// Pad to desired size
			padding := strings.Repeat("# comment\n", int(tt.fileSize/10))
			content := validYAML + padding

			// Truncate or extend to exact size
			if int64(len(content)) < tt.fileSize {
				content += strings.Repeat("x", int(tt.fileSize)-len(content))
			} else {
				content = content[:tt.fileSize]
			}

			if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			_, err := loadFromFile(testFile)

			if tt.wantErr {
				if err == nil {
					t.Errorf("loadFromFile() expected error for file size %d, got nil", tt.fileSize)
				} else if !strings.Contains(err.Error(), "exceeds maximum size") {
					t.Errorf("loadFromFile() expected size limit error, got: %v", err)
				}
			} else {
				// Note: May fail due to malformed YAML from padding, that's OK for this test
				// We're primarily testing size enforcement
				if err != nil && strings.Contains(err.Error(), "exceeds maximum size") {
					t.Errorf("loadFromFile() unexpected size limit error: %v", err)
				}
			}
		})
	}
}

// TestLoadFromFile_ValidConfig tests loading valid configurations
func TestLoadFromFile_ValidConfig(t *testing.T) {
	tmpDir := t.TempDir()

	validConfig := `contexts:
  custom:
    commands:
      build: make all
      test: make test
      deploy: ./deploy.sh
`

	testFile := filepath.Join(tmpDir, "valid.yaml")
	if err := os.WriteFile(testFile, []byte(validConfig), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cfg, err := loadFromFile(testFile)
	if err != nil {
		t.Fatalf("loadFromFile() unexpected error: %v", err)
	}

	// Verify custom context loaded
	customCtx, exists := cfg.Contexts["custom"]
	if !exists {
		t.Error("expected custom context, not found")
	}

	// Verify commands
	if customCtx.Commands["build"] != "make all" {
		t.Errorf("expected build command 'make all', got %q", customCtx.Commands["build"])
	}

	// Verify defaults were merged
	if _, exists := cfg.Contexts["node"]; !exists {
		t.Error("expected default 'node' context to be merged")
	}
}

// TestLoadFromFile_InvalidYAML tests YAML parsing security
func TestLoadFromFile_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		content string
		errMsg  string
	}{
		{
			name:    "completely invalid YAML",
			content: "{ invalid yaml content ][[ }",
			errMsg:  "invalid YAML format",
		},
		{
			name:    "malicious nested structures",
			content: strings.Repeat("a:\n  ", 1000) + "value: test",
			errMsg:  "", // Should be caught by unmarshal or validation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, tt.name+".yaml")
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			_, err := loadFromFile(testFile)
			if err == nil {
				t.Errorf("loadFromFile() expected error for invalid YAML, got nil")
			}
		})
	}
}

// TestValidateConfig tests configuration validation
func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &Config{
				Contexts: map[string]ContextConfig{
					"test": {
						Commands: map[string]string{
							"build": "make",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "nil contexts",
			config: &Config{
				Contexts: nil,
			},
			wantErr: true,
			errMsg:  "no contexts defined",
		},
		{
			name: "too many contexts",
			config: func() *Config {
				contexts := make(map[string]ContextConfig)
				for i := 0; i < MaxContexts+1; i++ {
					contexts[fmt.Sprintf("ctx%d", i)] = ContextConfig{
						Commands: map[string]string{"test": "echo"},
					}
				}
				return &Config{Contexts: contexts}
			}(),
			wantErr: true,
			errMsg:  "too many contexts",
		},
		{
			name: "too many commands in context",
			config: func() *Config {
				commands := make(map[string]string)
				for i := 0; i < MaxCommandsPerContext+1; i++ {
					commands[fmt.Sprintf("cmd%d", i)] = "echo test"
				}
				return &Config{
					Contexts: map[string]ContextConfig{
						"test": {Commands: commands},
					},
				}
			}(),
			wantErr: true,
			errMsg:  "too many commands",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateConfig() expected error, got nil")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("validateConfig() error = %v, want error containing %q", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateConfig() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestValidateContextName tests context name validation
func TestValidateContextName(t *testing.T) {
	tests := []struct {
		name    string
		ctxName string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid simple name",
			ctxName: "node",
			wantErr: false,
		},
		{
			name:    "valid with dash",
			ctxName: "my-context",
			wantErr: false,
		},
		{
			name:    "valid with underscore",
			ctxName: "my_context",
			wantErr: false,
		},
		{
			name:    "valid mixed",
			ctxName: "My_Context-2024",
			wantErr: false,
		},
		{
			name:    "empty name",
			ctxName: "",
			wantErr: true,
			errMsg:  "empty context name",
		},
		{
			name:    "too long",
			ctxName: strings.Repeat("a", 51),
			wantErr: true,
			errMsg:  "too long",
		},
		{
			name:    "invalid characters - slash",
			ctxName: "my/context",
			wantErr: true,
			errMsg:  "invalid character",
		},
		{
			name:    "invalid characters - space",
			ctxName: "my context",
			wantErr: true,
			errMsg:  "invalid character",
		},
		{
			name:    "invalid characters - special",
			ctxName: "ctx@123",
			wantErr: true,
			errMsg:  "invalid character",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateContextName(tt.ctxName)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateContextName() expected error, got nil")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("validateContextName() error = %v, want error containing %q", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateContextName() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestValidateCommand tests command validation
func TestValidateCommand(t *testing.T) {
	tests := []struct {
		name    string
		cmdName string
		cmdStr  string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid command",
			cmdName: "build",
			cmdStr:  "npm run build",
			wantErr: false,
		},
		{
			name:    "empty command name",
			cmdName: "",
			cmdStr:  "test",
			wantErr: true,
			errMsg:  "empty command name",
		},
		{
			name:    "empty command string",
			cmdName: "test",
			cmdStr:  "",
			wantErr: true,
			errMsg:  "empty command string",
		},
		{
			name:    "command name too long",
			cmdName: strings.Repeat("a", 51),
			cmdStr:  "test",
			wantErr: true,
			errMsg:  "too long",
		},
		{
			name:    "command string too long",
			cmdName: "build",
			cmdStr:  strings.Repeat("a", MaxCommandLength+1),
			wantErr: true,
			errMsg:  "exceeds maximum length",
		},
		{
			name:    "command with shell metacharacters - allowed in config",
			cmdName: "complex",
			cmdStr:  "npm run build && npm run test | tee output.log",
			wantErr: false, // Allowed in config files
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCommand(tt.cmdName, tt.cmdStr)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateCommand() expected error, got nil")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("validateCommand() error = %v, want error containing %q", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateCommand() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestLoad_PathTraversalPrevention tests that Load prevents path traversal
func TestLoad_PathTraversalPrevention(t *testing.T) {
	attacks := []string{
		"../../../etc/passwd",
		"../../.ssh/id_rsa",
		"/etc/shadow",
		"~/.ssh/id_rsa",
	}

	for _, attack := range attacks {
		t.Run(attack, func(t *testing.T) {
			_, err := Load(attack)
			if err == nil {
				t.Errorf("Load() should prevent path traversal for %q, but succeeded", attack)
			}
		})
	}
}

// TestLoad_DefaultConfig tests loading when no config file exists
func TestLoad_DefaultConfig(t *testing.T) {
	// Change to temp directory where no config exists
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	// Verify default contexts exist
	expectedContexts := []string{"node", "go", "python", "rust", "make"}
	for _, ctx := range expectedContexts {
		if _, exists := cfg.Contexts[ctx]; !exists {
			t.Errorf("expected default context %q, not found", ctx)
		}
	}
}

// TestFileExists tests the fileExists helper
func TestFileExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tmpDir, "exists.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create a directory
	testDir := filepath.Join(tmpDir, "dir")
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "existing file",
			path: testFile,
			want: true,
		},
		{
			name: "non-existent file",
			path: filepath.Join(tmpDir, "nonexistent.txt"),
			want: false,
		},
		{
			name: "directory - should return false",
			path: testDir,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fileExists(tt.path)
			if got != tt.want {
				t.Errorf("fileExists(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

// Benchmark tests
func BenchmarkValidateConfig(b *testing.B) {
	cfg := &Config{
		Contexts: map[string]ContextConfig{
			"test": {
				Commands: map[string]string{
					"build": "make",
					"test":  "make test",
				},
			},
		},
	}

	for i := 0; i < b.N; i++ {
		_ = validateConfig(cfg)
	}
}

func BenchmarkValidateContextName(b *testing.B) {
	name := "my-valid-context"
	for i := 0; i < b.N; i++ {
		_ = validateContextName(name)
	}
}
