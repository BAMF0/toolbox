package plugin

import (
	"os"
	"path/filepath"
	"testing"
)

// TestPluginManager_RegisterPlugin tests plugin registration
func TestPluginManager_RegisterPlugin(t *testing.T) {
	pm := NewPluginManager("/tmp/plugins")
	dockerPlugin := NewDockerPlugin()

	err := pm.RegisterPlugin(dockerPlugin)
	if err != nil {
		t.Fatalf("RegisterPlugin() failed: %v", err)
	}

	// Verify plugin is registered
	plugins := pm.GetPlugins()
	if len(plugins) != 1 {
		t.Errorf("expected 1 plugin, got %d", len(plugins))
	}

	// Verify metadata
	metadata := pm.GetMetadata()
	if meta, exists := metadata["docker"]; !exists {
		t.Error("expected metadata for docker plugin")
	} else {
		if meta.Name != "docker" {
			t.Errorf("expected name 'docker', got %q", meta.Name)
		}
		if meta.Version != "1.0.0" {
			t.Errorf("expected version '1.0.0', got %q", meta.Version)
		}
		if !meta.Enabled {
			t.Error("expected plugin to be enabled")
		}
	}
}

// TestPluginManager_DuplicateRegistration tests duplicate plugin detection
func TestPluginManager_DuplicateRegistration(t *testing.T) {
	pm := NewPluginManager("/tmp/plugins")
	dockerPlugin1 := NewDockerPlugin()
	dockerPlugin2 := NewDockerPlugin()

	// First registration should succeed
	err := pm.RegisterPlugin(dockerPlugin1)
	if err != nil {
		t.Fatalf("first RegisterPlugin() failed: %v", err)
	}

	// Second registration should fail
	err = pm.RegisterPlugin(dockerPlugin2)
	if err == nil {
		t.Error("expected error for duplicate plugin, got nil")
	}
}

// TestPluginManager_MultiplePlugins tests multiple plugin registration
func TestPluginManager_MultiplePlugins(t *testing.T) {
	pm := NewPluginManager("/tmp/plugins")

	plugins := []Plugin{
		NewDockerPlugin(),
		NewKubernetesPlugin(),
	}

	for _, plugin := range plugins {
		if err := pm.RegisterPlugin(plugin); err != nil {
			t.Fatalf("RegisterPlugin() failed for %s: %v", plugin.Name(), err)
		}
	}

	// Verify both plugins are registered
	registered := pm.GetPlugins()
	if len(registered) != 2 {
		t.Errorf("expected 2 plugins, got %d", len(registered))
	}
}

// TestPluginManager_GetContexts tests context merging from multiple plugins
func TestPluginManager_GetContexts(t *testing.T) {
	pm := NewPluginManager("/tmp/plugins")
	
	pm.RegisterPlugin(NewDockerPlugin())
	pm.RegisterPlugin(NewKubernetesPlugin())

	contexts := pm.GetContexts()

	// Should have both namespaced and non-namespaced contexts
	expectedContexts := []string{
		"docker",
		"docker-compose",
		"kubernetes",
		"helm",
	}

	for _, expected := range expectedContexts {
		if _, exists := contexts[expected]; !exists {
			t.Errorf("expected context %q not found", expected)
		}
	}

	// Verify namespaced contexts exist
	if _, exists := contexts["docker:docker"]; !exists {
		t.Error("expected namespaced context 'docker:docker'")
	}
}

// TestDockerPlugin_Detect tests Docker context detection
func TestDockerPlugin_Detect(t *testing.T) {
	plugin := NewDockerPlugin()

	tests := []struct {
		name           string
		setupFiles     []string
		expectedCtx    string
		expectedDetect bool
	}{
		{
			name:           "Dockerfile present",
			setupFiles:     []string{"Dockerfile"},
			expectedCtx:    "docker",
			expectedDetect: true,
		},
		{
			name:           "docker-compose.yml present",
			setupFiles:     []string{"docker-compose.yml"},
			expectedCtx:    "docker-compose",
			expectedDetect: true,
		},
		{
			name:           "docker-compose.yaml present",
			setupFiles:     []string{"docker-compose.yaml"},
			expectedCtx:    "docker-compose",
			expectedDetect: true,
		},
		{
			name:           "no Docker files",
			setupFiles:     []string{"main.go"},
			expectedCtx:    "",
			expectedDetect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create test files
			for _, file := range tt.setupFiles {
				filePath := filepath.Join(tmpDir, file)
				if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
			}

			ctx, detected := plugin.Detect(tmpDir)

			if detected != tt.expectedDetect {
				t.Errorf("Detect() detected = %v, want %v", detected, tt.expectedDetect)
			}

			if ctx != tt.expectedCtx {
				t.Errorf("Detect() context = %q, want %q", ctx, tt.expectedCtx)
			}
		})
	}
}

// TestKubernetesPlugin_Detect tests Kubernetes context detection
func TestKubernetesPlugin_Detect(t *testing.T) {
	plugin := NewKubernetesPlugin()

	tests := []struct {
		name           string
		setupFiles     []string
		expectedCtx    string
		expectedDetect bool
	}{
		{
			name:           "deployment.yaml present",
			setupFiles:     []string{"deployment.yaml"},
			expectedCtx:    "kubernetes",
			expectedDetect: true,
		},
		{
			name:           "Chart.yaml present",
			setupFiles:     []string{"Chart.yaml"},
			expectedCtx:    "helm",
			expectedDetect: true,
		},
		{
			name:           "k8s/deployment.yaml present",
			setupFiles:     []string{"k8s/deployment.yaml"},
			expectedCtx:    "kubernetes",
			expectedDetect: true,
		},
		{
			name:           "no Kubernetes files",
			setupFiles:     []string{"main.go"},
			expectedCtx:    "",
			expectedDetect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create test files
			for _, file := range tt.setupFiles {
				filePath := filepath.Join(tmpDir, file)
				fileDir := filepath.Dir(filePath)
				
				if err := os.MkdirAll(fileDir, 0755); err != nil {
					t.Fatalf("failed to create directory: %v", err)
				}
				
				if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
			}

			ctx, detected := plugin.Detect(tmpDir)

			if detected != tt.expectedDetect {
				t.Errorf("Detect() detected = %v, want %v", detected, tt.expectedDetect)
			}

			if ctx != tt.expectedCtx {
				t.Errorf("Detect() context = %q, want %q", ctx, tt.expectedCtx)
			}
		})
	}
}

// TestDockerPlugin_Contexts tests Docker plugin contexts
func TestDockerPlugin_Contexts(t *testing.T) {
	plugin := NewDockerPlugin()
	contexts := plugin.Contexts()

	// Verify docker context exists
	dockerCtx, exists := contexts["docker"]
	if !exists {
		t.Fatal("expected 'docker' context")
	}

	// Verify expected commands
	expectedCommands := []string{"build", "run", "push", "compose", "stop", "logs", "shell"}
	for _, cmd := range expectedCommands {
		if _, exists := dockerCtx.Commands[cmd]; !exists {
			t.Errorf("expected command %q in docker context", cmd)
		}
	}

	// Verify docker-compose context
	if _, exists := contexts["docker-compose"]; !exists {
		t.Error("expected 'docker-compose' context")
	}
}

// TestDockerPlugin_Validate tests plugin validation
func TestDockerPlugin_Validate(t *testing.T) {
	plugin := NewDockerPlugin()
	
	err := plugin.Validate()
	if err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}
}

// TestKubernetesPlugin_Validate tests Kubernetes plugin validation
func TestKubernetesPlugin_Validate(t *testing.T) {
	plugin := NewKubernetesPlugin()
	
	err := plugin.Validate()
	if err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}
}

// TestPluginManager_DetectContext tests context detection via plugin manager
func TestPluginManager_DetectContext(t *testing.T) {
	pm := NewPluginManager("/tmp/plugins")
	pm.RegisterPlugin(NewDockerPlugin())
	pm.RegisterPlugin(NewKubernetesPlugin())

	tmpDir := t.TempDir()

	// Create Dockerfile
	dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte("FROM alpine"), 0644); err != nil {
		t.Fatalf("failed to create Dockerfile: %v", err)
	}

	ctx, pluginName, found := pm.DetectContext(tmpDir)

	if !found {
		t.Error("expected context to be detected")
	}

	if ctx != "docker" {
		t.Errorf("expected context 'docker', got %q", ctx)
	}

	if pluginName != "docker" {
		t.Errorf("expected plugin 'docker', got %q", pluginName)
	}
}

// TestPluginManager_AddTrustedHash tests hash allowlist
func TestPluginManager_AddTrustedHash(t *testing.T) {
	pm := NewPluginManager("/tmp/plugins")

	testHash := "abcdef1234567890"
	pm.AddTrustedHash(testHash)

	if !pm.allowedHashes[testHash] {
		t.Error("hash not added to allowlist")
	}

	pm.RemoveTrustedHash(testHash)

	if pm.allowedHashes[testHash] {
		t.Error("hash not removed from allowlist")
	}
}

// TestValidatePluginPath tests plugin path validation
func TestValidatePluginPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		setup   func() (string, func())
		wantErr bool
		errMsg  string
	}{
		{
			name:    "directory traversal",
			path:    "../../../etc/passwd",
			wantErr: true,
			errMsg:  "directory traversal",
		},
		{
			name: "valid .so file",
			setup: func() (string, func()) {
				tmpDir := t.TempDir()
				soFile := filepath.Join(tmpDir, "plugin.so")
				os.WriteFile(soFile, []byte("fake plugin"), 0644)
				return soFile, func() {}
			},
			wantErr: false,
		},
		{
			name: "invalid extension",
			setup: func() (string, func()) {
				tmpDir := t.TempDir()
				txtFile := filepath.Join(tmpDir, "plugin.txt")
				os.WriteFile(txtFile, []byte("fake plugin"), 0644)
				return txtFile, func() {}
			},
			wantErr: true,
			errMsg:  "invalid plugin extension",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var path string
			var cleanup func()

			if tt.setup != nil {
				path, cleanup = tt.setup()
				if cleanup != nil {
					defer cleanup()
				}
			} else {
				path = tt.path
			}

			err := validatePluginPath(path)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("error = %v, want error containing %q", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestCalculateFileHash tests file hash calculation
func TestCalculateFileHash(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := []byte("test content")

	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	hash, err := calculateFileHash(testFile)
	if err != nil {
		t.Fatalf("calculateFileHash() failed: %v", err)
	}

	if hash == "" {
		t.Error("expected non-empty hash")
	}

	// Hash should be deterministic
	hash2, err := calculateFileHash(testFile)
	if err != nil {
		t.Fatalf("second calculateFileHash() failed: %v", err)
	}

	if hash != hash2 {
		t.Error("hash should be deterministic")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) && s[:len(substr)] == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Benchmark tests
func BenchmarkPluginManager_DetectContext(b *testing.B) {
	pm := NewPluginManager("/tmp/plugins")
	pm.RegisterPlugin(NewDockerPlugin())
	pm.RegisterPlugin(NewKubernetesPlugin())

	tmpDir := b.TempDir()
	dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
	os.WriteFile(dockerfilePath, []byte("FROM alpine"), 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pm.DetectContext(tmpDir)
	}
}
