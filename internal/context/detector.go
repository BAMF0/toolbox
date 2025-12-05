package context

import (
	"fmt"
	"os"
	"path/filepath"
)

// Detector identifies the project context based on marker files
type Detector struct {
	// Map of context name to marker files that identify it
	markers map[string][]string
}

// NewDetector creates a new context detector with default markers
func NewDetector() *Detector {
	return &Detector{
		markers: map[string][]string{
			"node":   {"package.json", "package-lock.json", "yarn.lock", "pnpm-lock.yaml"},
			"go":     {"go.mod", "go.sum"},
			"python": {"pyproject.toml", "setup.py", "requirements.txt", "Pipfile"},
			"rust":   {"Cargo.toml", "Cargo.lock"},
			"make":   {"Makefile", "makefile"},
			"ruby":   {"Gemfile", "Gemfile.lock"},
			"java":   {"pom.xml", "build.gradle", "build.gradle.kts"},
			"php":    {"composer.json", "composer.lock"},
		},
	}
}

// Detect identifies the project context by searching for marker files
// Returns the first matching context or an error if none found
func (d *Detector) Detect(dir string) (string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Search current directory and up to 3 levels of parents
	// This allows detection even when in subdirectories
	searchDir := absDir
	for i := 0; i < 4; i++ {
		context, found := d.detectInDirectory(searchDir)
		if found {
			return context, nil
		}

		// Move up one directory
		parent := filepath.Dir(searchDir)
		if parent == searchDir {
			// Reached root
			break
		}
		searchDir = parent
	}

	return "", fmt.Errorf("no recognized project context found in %s or parent directories", absDir)
}

// detectInDirectory checks for marker files in a specific directory
func (d *Detector) detectInDirectory(dir string) (string, bool) {
	// Check each context's markers
	// Priority order matters - checked in map iteration order
	// For deterministic results, we check in a specific order
	priorityOrder := []string{"node", "go", "python", "rust", "java", "ruby", "php", "make"}

	for _, ctx := range priorityOrder {
		markers, exists := d.markers[ctx]
		if !exists {
			continue
		}

		for _, marker := range markers {
			markerPath := filepath.Join(dir, marker)
			if fileExists(markerPath) {
				return ctx, true
			}
		}
	}

	return "", false
}

// AddMarker adds a custom marker file for a context
func (d *Detector) AddMarker(context, markerFile string) {
	if _, exists := d.markers[context]; !exists {
		d.markers[context] = []string{}
	}
	d.markers[context] = append(d.markers[context], markerFile)
}

// FileExists checks if a file exists in the current directory
func (d *Detector) FileExists(filename string) bool {
	return fileExists(filename)
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
