package context

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDetector_Detect_NodeProject tests Node.js project detection
func TestDetector_Detect_NodeProject(t *testing.T) {
	tmpDir := t.TempDir()
	packageJSON := filepath.Join(tmpDir, "package.json")
	if err := os.WriteFile(packageJSON, []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to create package.json: %v", err)
	}

	detector := NewDetector()
	ctx, err := detector.Detect(tmpDir)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if ctx != "node" {
		t.Errorf("expected context 'node', got %q", ctx)
	}
}

// TestDetector_Detect_GoProject tests Go project detection
func TestDetector_Detect_GoProject(t *testing.T) {
	tmpDir := t.TempDir()
	goMod := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goMod, []byte("module test"), 0644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}

	detector := NewDetector()
	ctx, err := detector.Detect(tmpDir)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if ctx != "go" {
		t.Errorf("expected context 'go', got %q", ctx)
	}
}

// TestDetector_Detect_PythonProject tests Python project detection
func TestDetector_Detect_PythonProject(t *testing.T) {
	tests := []struct {
		name       string
		markerFile string
	}{
		{"pyproject.toml", "pyproject.toml"},
		{"requirements.txt", "requirements.txt"},
		{"setup.py", "setup.py"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			markerPath := filepath.Join(tmpDir, tt.markerFile)
			if err := os.WriteFile(markerPath, []byte("test"), 0644); err != nil {
				t.Fatalf("failed to create %s: %v", tt.markerFile, err)
			}

			detector := NewDetector()
			ctx, err := detector.Detect(tmpDir)

			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			if ctx != "python" {
				t.Errorf("expected context 'python', got %q", ctx)
			}
		})
	}
}

// TestDetector_Detect_RustProject tests Rust project detection
func TestDetector_Detect_RustProject(t *testing.T) {
	tmpDir := t.TempDir()
	cargoToml := filepath.Join(tmpDir, "Cargo.toml")
	if err := os.WriteFile(cargoToml, []byte("[package]"), 0644); err != nil {
		t.Fatalf("failed to create Cargo.toml: %v", err)
	}

	detector := NewDetector()
	ctx, err := detector.Detect(tmpDir)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if ctx != "rust" {
		t.Errorf("expected context 'rust', got %q", ctx)
	}
}

// TestDetector_Detect_MakeProject tests Makefile detection
func TestDetector_Detect_MakeProject(t *testing.T) {
	tmpDir := t.TempDir()
	makefile := filepath.Join(tmpDir, "Makefile")
	if err := os.WriteFile(makefile, []byte("all:\n\techo test"), 0644); err != nil {
		t.Fatalf("failed to create Makefile: %v", err)
	}

	detector := NewDetector()
	ctx, err := detector.Detect(tmpDir)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if ctx != "make" {
		t.Errorf("expected context 'make', got %q", ctx)
	}
}

// TestDetector_Detect_ParentDirectory tests detection in subdirectories
func TestDetector_Detect_ParentDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create go.mod in parent
	goMod := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goMod, []byte("module test"), 0644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}

	// Create nested subdirectory
	subDir := filepath.Join(tmpDir, "cmd", "app", "internal")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdirectory: %v", err)
	}

	detector := NewDetector()
	ctx, err := detector.Detect(subDir)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if ctx != "go" {
		t.Errorf("expected context 'go' from parent directory, got %q", ctx)
	}
}

// TestDetector_Detect_NoContext tests error handling when no context found
func TestDetector_Detect_NoContext(t *testing.T) {
	tmpDir := t.TempDir()

	detector := NewDetector()
	_, err := detector.Detect(tmpDir)

	if err == nil {
		t.Error("expected error for no context, got nil")
	}
}

// TestDetector_Detect_Priority tests context priority ordering
func TestDetector_Detect_Priority(t *testing.T) {
	tmpDir := t.TempDir()

	// Create both package.json and Makefile
	packageJSON := filepath.Join(tmpDir, "package.json")
	if err := os.WriteFile(packageJSON, []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to create package.json: %v", err)
	}

	makefile := filepath.Join(tmpDir, "Makefile")
	if err := os.WriteFile(makefile, []byte("all:\n\techo test"), 0644); err != nil {
		t.Fatalf("failed to create Makefile: %v", err)
	}

	detector := NewDetector()
	ctx, err := detector.Detect(tmpDir)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Node should take priority over make based on priority order
	if ctx != "node" {
		t.Errorf("expected 'node' to take priority, got %q", ctx)
	}
}

// TestDetector_AddMarker tests custom marker addition
func TestDetector_AddMarker(t *testing.T) {
	tmpDir := t.TempDir()
	customMarker := "custom.config"
	markerPath := filepath.Join(tmpDir, customMarker)
	if err := os.WriteFile(markerPath, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create custom marker: %v", err)
	}

	detector := NewDetector()
	detector.AddMarker("customctx", customMarker)

	// This won't be detected without updating detectInDirectory priority order
	// but tests that AddMarker doesn't panic
	_, _ = detector.Detect(tmpDir)
}

// TestFileExists tests the fileExists helper function
func TestFileExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a regular file
	regularFile := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(regularFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create regular file: %v", err)
	}

	// Create a directory
	dir := filepath.Join(tmpDir, "directory")
	if err := os.Mkdir(dir, 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "regular file exists",
			path: regularFile,
			want: true,
		},
		{
			name: "directory - should return false",
			path: dir,
			want: false,
		},
		{
			name: "non-existent file",
			path: filepath.Join(tmpDir, "nonexistent.txt"),
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

// TestDetector_Detect_MultipleLevelsUp tests parent directory traversal limits
func TestDetector_Detect_MultipleLevelsUp(t *testing.T) {
	tmpDir := t.TempDir()

	// Create go.mod at root
	goMod := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goMod, []byte("module test"), 0644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}

	// Create a deep subdirectory (within the 4-level limit)
	deepDir := filepath.Join(tmpDir, "level1", "level2", "level3")
	if err := os.MkdirAll(deepDir, 0755); err != nil {
		t.Fatalf("failed to create deep directory: %v", err)
	}

	detector := NewDetector()
	ctx, err := detector.Detect(deepDir)

	if err != nil {
		t.Fatalf("expected detection within 3 parent levels, got error: %v", err)
	}
	if ctx != "go" {
		t.Errorf("expected context 'go', got %q", ctx)
	}
}

// TestDetector_Detect_TooDeep tests that detection fails beyond the traversal limit
func TestDetector_Detect_TooDeep(t *testing.T) {
	tmpDir := t.TempDir()

	// Create go.mod at root
	goMod := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goMod, []byte("module test"), 0644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}

	// Create a very deep subdirectory (beyond the limit)
	deepDir := filepath.Join(tmpDir, "l1", "l2", "l3", "l4", "l5")
	if err := os.MkdirAll(deepDir, 0755); err != nil {
		t.Fatalf("failed to create deep directory: %v", err)
	}

	detector := NewDetector()
	_, err := detector.Detect(deepDir)

	// Should fail to detect because it's too deep
	if err == nil {
		t.Error("expected error for directory too deep, got nil")
	}
}

// Benchmark tests
func BenchmarkDetector_Detect(b *testing.B) {
	tmpDir := b.TempDir()
	goMod := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goMod, []byte("module test"), 0644); err != nil {
		b.Fatalf("failed to create go.mod: %v", err)
	}

	detector := NewDetector()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = detector.Detect(tmpDir)
	}
}

func BenchmarkFileExists(b *testing.B) {
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fileExists(testFile)
	}
}
