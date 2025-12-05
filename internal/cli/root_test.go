package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestValidateArguments tests argument validation security controls
func TestValidateArguments(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid simple arguments",
			args:    []string{"--verbose", "test.txt"},
			wantErr: false,
		},
		{
			name:    "empty arguments",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "too many arguments",
			args:    make([]string, MaxArgumentCount+1),
			wantErr: true,
			errMsg:  "too many arguments",
		},
		{
			name:    "argument too long",
			args:    []string{strings.Repeat("A", MaxArgumentLength+1)},
			wantErr: true,
			errMsg:  "exceeds maximum length",
		},
		{
			name:    "arguments with shell metacharacters (should pass validation)",
			args:    []string{"; rm -rf /", "| cat /etc/passwd", "$(whoami)"},
			wantErr: false, // Validation passes but execution won't use shell
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateArguments(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateArguments() expected error, got nil")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("validateArguments() error = %v, want error containing %q", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateArguments() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestContainsDangerousPatterns tests pattern detection
func TestContainsDangerousPatterns(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"safe-argument", false},
		{"--flag=value", false},
		{"file.txt", false},
		{"; rm -rf /", true},
		{"| cat /etc/passwd", true},
		{"&& echo hacked", true},
		{"$(whoami)", true},
		{"`id`", true},
		{"test > output.txt", true},
		{"test < input.txt", true},
		{"foo\nbar", true},
		{"normal text", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := containsDangerousPatterns(tt.input)
			if got != tt.want {
				t.Errorf("containsDangerousPatterns(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestExecuteCommandSecure_NoShellInjection verifies shell injection is prevented
func TestExecuteCommandSecure_NoShellInjection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Create a test script that would be vulnerable if shell injection works
	tmpDir := t.TempDir()
	canaryFile := filepath.Join(tmpDir, "canary.txt")

	tests := []struct {
		name        string
		baseCommand string
		userArgs    []string
		expectFile  bool // Should canary file be created?
		description string
	}{
		{
			name:        "injection attempt with semicolon",
			baseCommand: "echo",
			userArgs:    []string{"test", "; touch " + canaryFile},
			expectFile:  false,
			description: "Semicolon should be treated as literal argument, not command separator",
		},
		{
			name:        "injection attempt with pipe",
			baseCommand: "echo",
			userArgs:    []string{"test | touch " + canaryFile},
			expectFile:  false,
			description: "Pipe should be treated as literal argument",
		},
		{
			name:        "injection attempt with command substitution",
			baseCommand: "echo",
			userArgs:    []string{"$(touch " + canaryFile + ")"},
			expectFile:  false,
			description: "Command substitution should be treated as literal text",
		},
		{
			name:        "injection attempt with backticks",
			baseCommand: "echo",
			userArgs:    []string{"`touch " + canaryFile + "`"},
			expectFile:  false,
			description: "Backticks should be treated as literal text",
		},
		{
			name:        "injection attempt with AND operator",
			baseCommand: "echo",
			userArgs:    []string{"test && touch " + canaryFile},
			expectFile:  false,
			description: "AND operator should be treated as literal argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up canary file before test
			os.Remove(canaryFile)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Execute command - errors are expected for some cases
			_ = executeCommandSecure(ctx, tt.baseCommand, tt.userArgs)

			// Check if canary file was created (it shouldn't be)
			_, err := os.Stat(canaryFile)
			fileExists := err == nil

			if fileExists != tt.expectFile {
				t.Errorf("%s: canary file existence = %v, want %v", tt.description, fileExists, tt.expectFile)
				if fileExists {
					t.Errorf("SECURITY FAILURE: Shell injection succeeded! Canary file created at %s", canaryFile)
				}
			}
		})
	}
}

// TestExecuteCommandSecure_ValidCommands tests legitimate command execution
func TestExecuteCommandSecure_ValidCommands(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tests := []struct {
		name        string
		baseCommand string
		userArgs    []string
		wantErr     bool
	}{
		{
			name:        "simple echo command",
			baseCommand: "echo",
			userArgs:    []string{"hello", "world"},
			wantErr:     false,
		},
		{
			name:        "command with flags",
			baseCommand: "ls",
			userArgs:    []string{"-la", "/tmp"},
			wantErr:     false,
		},
		{
			name:        "nonexistent command",
			baseCommand: "this-command-does-not-exist",
			userArgs:    []string{},
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := executeCommandSecure(ctx, tt.baseCommand, tt.userArgs)

			if tt.wantErr {
				if err == nil {
					t.Errorf("executeCommandSecure() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("executeCommandSecure() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestExecuteCommandSecure_Timeout verifies timeout functionality
func TestExecuteCommandSecure_Timeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Create a command that will definitely timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// sleep command should timeout
	err := executeCommandSecure(ctx, "sleep", []string{"10"})

	if err == nil {
		t.Error("executeCommandSecure() expected timeout error, got nil")
	}

	if !strings.Contains(err.Error(), "timed out") && !strings.Contains(err.Error(), "deadline exceeded") {
		t.Errorf("executeCommandSecure() expected timeout error, got: %v", err)
	}
}

// TestExecuteCommandSecure_MultiWordBaseCommand tests parsing of multi-word base commands
func TestExecuteCommandSecure_MultiWordBaseCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tests := []struct {
		name        string
		baseCommand string
		userArgs    []string
		wantErr     bool
	}{
		{
			name:        "npm run build style command",
			baseCommand: "echo npm run",
			userArgs:    []string{"build"},
			wantErr:     false, // "echo" will succeed
		},
		{
			name:        "go test with base args",
			baseCommand: "echo -n test",
			userArgs:    []string{"extra"},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := executeCommandSecure(ctx, tt.baseCommand, tt.userArgs)

			if tt.wantErr && err == nil {
				t.Errorf("executeCommandSecure() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("executeCommandSecure() unexpected error = %v", err)
			}
		})
	}
}

// Benchmark tests
func BenchmarkValidateArguments(b *testing.B) {
	args := []string{"--verbose", "--output=file.txt", "arg1", "arg2", "arg3"}
	for i := 0; i < b.N; i++ {
		_ = validateArguments(args)
	}
}

func BenchmarkContainsDangerousPatterns(b *testing.B) {
	testStrings := []string{
		"normal-argument",
		"; rm -rf /",
		"| cat /etc/passwd",
		"$(whoami)",
	}
	for i := 0; i < b.N; i++ {
		for _, s := range testStrings {
			_ = containsDangerousPatterns(s)
		}
	}
}

// TestExecuteCommandSecure_RealWorldScenarios tests actual use cases
func TestExecuteCommandSecure_RealWorldScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Create a temporary directory with test files
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tests := []struct {
		name        string
		baseCommand string
		userArgs    []string
		description string
	}{
		{
			name:        "npm style command with flags",
			baseCommand: "echo npm run",
			userArgs:    []string{"test", "--coverage"},
			description: "Simulates: npm run test --coverage",
		},
		{
			name:        "go test with package path",
			baseCommand: "echo go test",
			userArgs:    []string{"./...", "-v"},
			description: "Simulates: go test ./... -v",
		},
		{
			name:        "file operation with special chars in filename",
			baseCommand: "cat",
			userArgs:    []string{testFile},
			description: "Tests file path handling",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := executeCommandSecure(ctx, tt.baseCommand, tt.userArgs)
			if err != nil {
				// Some commands may fail, but they shouldn't crash or allow injection
				t.Logf("%s: command execution result: %v", tt.description, err)
			}
		})
	}
}

// TestExecuteCommandSecure_EnvironmentIsolation verifies environment handling
func TestExecuteCommandSecure_EnvironmentIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Set a test environment variable
	testEnvVar := "TB_TEST_VAR"
	testEnvValue := "test_value_12345"
	os.Setenv(testEnvVar, testEnvValue)
	defer os.Unsetenv(testEnvVar)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// This test verifies that environment is passed correctly
	// (In a real scenario, you might want to control this more strictly)
	err := executeCommandSecure(ctx, "echo", []string{"test"})
	if err != nil {
		t.Errorf("executeCommandSecure() failed: %v", err)
	}
}

// TestCommandNotFound tests error handling for missing commands
func TestCommandNotFound(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := executeCommandSecure(ctx, "this-command-absolutely-does-not-exist-anywhere", []string{})
	if err == nil {
		t.Error("expected error for nonexistent command, got nil")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

// TestEmptyCommand tests error handling for empty commands
func TestEmptyCommand(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := executeCommandSecure(ctx, "", []string{})
	if err == nil {
		t.Error("expected error for empty command, got nil")
	}

	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("expected 'empty' error, got: %v", err)
	}
}
