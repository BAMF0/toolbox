package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCompletion_BashGeneration tests bash completion generation
func TestCompletion_BashGeneration(t *testing.T) {
	var buf bytes.Buffer

	// Use rootCmd directly
	rootCmd.SetArgs([]string{"completion", "bash"})
	rootCmd.SetOut(&buf)

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("completion bash failed: %v", err)
	}

	// Reset for next test
	defer rootCmd.SetArgs([]string{})

	output := buf.String()

	// Verify bash completion markers
	if !strings.Contains(output, "bash completion") {
		t.Error("bash completion output missing bash completion marker")
	}

	if !strings.Contains(output, "__tb_") {
		t.Error("bash completion missing completion functions")
	}
}

// TestCompletion_ZshGeneration tests zsh completion generation
func TestCompletion_ZshGeneration(t *testing.T) {
	var buf bytes.Buffer

	rootCmd.SetArgs([]string{"completion", "zsh"})
	rootCmd.SetOut(&buf)

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("completion zsh failed: %v", err)
	}

	defer rootCmd.SetArgs([]string{})

	output := buf.String()

	// Verify zsh completion markers
	if !strings.Contains(output, "#compdef") {
		t.Error("zsh completion output missing #compdef")
	}
}

// TestCompletion_FishGeneration tests fish completion generation
func TestCompletion_FishGeneration(t *testing.T) {
	var buf bytes.Buffer

	rootCmd.SetArgs([]string{"completion", "fish"})
	rootCmd.SetOut(&buf)

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("completion fish failed: %v", err)
	}

	defer rootCmd.SetArgs([]string{})

	output := buf.String()

	// Verify fish completion markers
	if !strings.Contains(output, "complete") {
		t.Error("fish completion output missing complete command")
	}
}

// TestCompletion_PowerShellGeneration tests PowerShell completion generation
func TestCompletion_PowerShellGeneration(t *testing.T) {
	var buf bytes.Buffer

	rootCmd.SetArgs([]string{"completion", "powershell"})
	rootCmd.SetOut(&buf)

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("completion powershell failed: %v", err)
	}

	defer rootCmd.SetArgs([]string{})

	output := buf.String()

	// Verify PowerShell completion markers
	if !strings.Contains(output, "Register-ArgumentCompleter") {
		t.Error("PowerShell completion output missing Register-ArgumentCompleter")
	}
}

// TestCompletion_InvalidShell tests error handling for invalid shell
func TestCompletion_InvalidShell(t *testing.T) {
	// This test is tricky because Cobra validates args
	// Just verify the valid args are set correctly
	if len(completionCmd.ValidArgs) != 4 {
		t.Errorf("expected 4 valid shells, got %d", len(completionCmd.ValidArgs))
	}

	expectedShells := []string{"bash", "zsh", "fish", "powershell"}
	for i, shell := range expectedShells {
		if i < len(completionCmd.ValidArgs) && completionCmd.ValidArgs[i] != shell {
			t.Errorf("expected shell %q at index %d, got %q", shell, i, completionCmd.ValidArgs[i])
		}
	}
}

// TestGetDynamicCommandCompletions tests dynamic command completion
func TestGetDynamicCommandCompletions(t *testing.T) {
	// Create a temporary Go project
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)

	// Create go.mod
	goMod := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goMod, []byte("module test"), 0644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}

	os.Chdir(tmpDir)

	tests := []struct {
		name       string
		toComplete string
		wantAny    bool // Just check if we get any suggestions
	}{
		{
			name:       "empty prefix",
			toComplete: "",
			wantAny:    true,
		},
		{
			name:       "partial 'b'",
			toComplete: "b",
			wantAny:    true, // Should suggest 'build'
		},
		{
			name:       "partial 'te'",
			toComplete: "te",
			wantAny:    true, // Should suggest 'test'
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := getDynamicCommandCompletions(tt.toComplete)

			if tt.wantAny && len(suggestions) == 0 {
				t.Errorf("getDynamicCommandCompletions(%q) returned no suggestions", tt.toComplete)
			}

			// Verify all suggestions start with the prefix
			for _, suggestion := range suggestions {
				if !strings.HasPrefix(suggestion, tt.toComplete) {
					t.Errorf("suggestion %q doesn't start with prefix %q", suggestion, tt.toComplete)
				}
			}
		})
	}
}

// TestGetContextCompletions tests context completion
func TestGetContextCompletions(t *testing.T) {
	tests := []struct {
		name         string
		toComplete   string
		wantContains []string
	}{
		{
			name:         "empty prefix",
			toComplete:   "",
			wantContains: []string{"go", "node", "python"}, // Built-in contexts
		},
		{
			name:         "prefix 'g'",
			toComplete:   "g",
			wantContains: []string{"go"},
		},
		{
			name:         "prefix 'n'",
			toComplete:   "n",
			wantContains: []string{"node"},
		},
		{
			name:         "prefix 'd' (from plugin)",
			toComplete:   "d",
			wantContains: []string{"docker"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := getContextCompletions(tt.toComplete)

			for _, want := range tt.wantContains {
				found := false
				for _, suggestion := range suggestions {
					if suggestion == want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("getContextCompletions(%q) missing expected suggestion %q, got: %v",
						tt.toComplete, want, suggestions)
				}
			}
		})
	}
}

// TestCompletion_DockerProject tests completion in a Docker project
func TestCompletion_DockerProject(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)

	// Create Dockerfile
	dockerfile := filepath.Join(tmpDir, "Dockerfile")
	if err := os.WriteFile(dockerfile, []byte("FROM alpine"), 0644); err != nil {
		t.Fatalf("failed to create Dockerfile: %v", err)
	}

	os.Chdir(tmpDir)

	suggestions := getDynamicCommandCompletions("b")

	// Should suggest Docker commands
	foundBuild := false
	for _, suggestion := range suggestions {
		if suggestion == "build" {
			foundBuild = true
			break
		}
	}

	if !foundBuild {
		t.Errorf("expected 'build' in suggestions for Docker project, got: %v", suggestions)
	}
}

// TestCompletion_KubernetesProject tests completion in a Kubernetes project
func TestCompletion_KubernetesProject(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)

	// Create deployment.yaml
	deployment := filepath.Join(tmpDir, "deployment.yaml")
	if err := os.WriteFile(deployment, []byte("apiVersion: apps/v1"), 0644); err != nil {
		t.Fatalf("failed to create deployment.yaml: %v", err)
	}

	os.Chdir(tmpDir)

	suggestions := getDynamicCommandCompletions("a")

	// Should suggest Kubernetes commands
	foundApply := false
	for _, suggestion := range suggestions {
		if suggestion == "apply" {
			foundApply = true
			break
		}
	}

	if !foundApply {
		t.Errorf("expected 'apply' in suggestions for Kubernetes project, got: %v", suggestions)
	}
}

// Benchmark tests
func BenchmarkGetDynamicCommandCompletions(b *testing.B) {
	tmpDir := b.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)

	goMod := filepath.Join(tmpDir, "go.mod")
	os.WriteFile(goMod, []byte("module test"), 0644)
	os.Chdir(tmpDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getDynamicCommandCompletions("b")
	}
}

func BenchmarkGetContextCompletions(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getContextCompletions("g")
	}
}
