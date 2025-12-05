package plugin

import (
	"os"
	"path/filepath"
	"testing"
)

// TestUbuntuPlugin_Basic tests basic plugin functionality
func TestUbuntuPlugin_Basic(t *testing.T) {
	plugin := NewUbuntuPlugin()

	if plugin.Name() != "ubuntu" {
		t.Errorf("expected name 'ubuntu', got %q", plugin.Name())
	}

	if plugin.Version() != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %q", plugin.Version())
	}

	if err := plugin.Validate(); err != nil {
		t.Errorf("validation failed: %v", err)
	}
}

// TestUbuntuPlugin_Contexts tests context provision
func TestUbuntuPlugin_Contexts(t *testing.T) {
	plugin := NewUbuntuPlugin()
	contexts := plugin.Contexts()

	if len(contexts) == 0 {
		t.Fatal("expected at least one context")
	}

	ctx, exists := contexts["ubuntu-packaging"]
	if !exists {
		t.Fatal("expected 'ubuntu-packaging' context")
	}

	expectedCommands := []string{
		"gbranch", "ppa-status", "dch-auto", "ubuild",
		"sb-auto", "dput-auto", "build", "lint",
	}

	for _, cmd := range expectedCommands {
		if _, exists := ctx.Commands[cmd]; !exists {
			t.Errorf("expected command %q not found", cmd)
		}
	}
}

// TestUbuntuPlugin_Detect tests project detection
func TestUbuntuPlugin_Detect(t *testing.T) {
	plugin := NewUbuntuPlugin()

	tests := []struct {
		name     string
		setup    func(string)
		expected bool
	}{
		{
			name: "debian/control present",
			setup: func(dir string) {
				os.MkdirAll(filepath.Join(dir, "debian"), 0755)
				os.WriteFile(filepath.Join(dir, "debian", "control"), []byte("Source: test"), 0644)
			},
			expected: true,
		},
		{
			name: "debian/changelog present",
			setup: func(dir string) {
				os.MkdirAll(filepath.Join(dir, "debian"), 0755)
				os.WriteFile(filepath.Join(dir, "debian", "changelog"), []byte("test (1.0) unstable; urgency=low"), 0644)
			},
			expected: true,
		},
		{
			name: "no debian directory",
			setup: func(dir string) {
				// No setup
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tt.setup(tmpDir)

			ctx, detected := plugin.Detect(tmpDir)

			if detected != tt.expected {
				t.Errorf("expected detection=%v, got %v", tt.expected, detected)
			}

			if detected && ctx != "ubuntu-packaging" {
				t.Errorf("expected context 'ubuntu-packaging', got %q", ctx)
			}
		})
	}
}

// TestParsePPAName_Merge tests parsing merge PPA names
func TestParsePPAName_Merge(t *testing.T) {
	tests := []struct {
		name        string
		ppaName     string
		expectError bool
		expected    *PPAInfo
	}{
		{
			name:        "valid merge PPA",
			ppaName:     "noble-efibootmgr-merge-lp2133493",
			expectError: false,
			expected: &PPAInfo{
				Release:     "noble",
				Project:     "efibootmgr",
				Type:        PPATypeMerge,
				BugID:       "2133493",
				Description: "",
				FullName:    "noble-efibootmgr-merge-lp2133493",
			},
		},
		{
			name:        "valid merge with hyphenated project",
			ppaName:     "jammy-sudo-rs-merge-lp2127080",
			expectError: false,
			expected: &PPAInfo{
				Release:     "jammy",
				Project:     "sudo-rs",
				Type:        PPATypeMerge,
				BugID:       "2127080",
				Description: "",
				FullName:    "jammy-sudo-rs-merge-lp2127080",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := ParsePPAName(tt.ppaName)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.expectError && info != nil {
				if info.Release != tt.expected.Release {
					t.Errorf("release: expected %q, got %q", tt.expected.Release, info.Release)
				}
				if info.Project != tt.expected.Project {
					t.Errorf("project: expected %q, got %q", tt.expected.Project, info.Project)
				}
				if info.Type != tt.expected.Type {
					t.Errorf("type: expected %q, got %q", tt.expected.Type, info.Type)
				}
				if info.BugID != tt.expected.BugID {
					t.Errorf("bugID: expected %q, got %q", tt.expected.BugID, info.BugID)
				}
			}
		})
	}
}

// TestParsePPAName_SRU tests parsing SRU PPA names
func TestParsePPAName_SRU(t *testing.T) {
	tests := []struct {
		name        string
		ppaName     string
		expectError bool
		expected    *PPAInfo
	}{
		{
			name:        "valid SRU with description",
			ppaName:     "jammy-sudo-rs-sru-lp2127080-escape-equals-question",
			expectError: false,
			expected: &PPAInfo{
				Release:     "jammy",
				Project:     "sudo-rs",
				Type:        PPATypeSRU,
				BugID:       "2127080",
				Description: "escape-equals-question",
				FullName:    "jammy-sudo-rs-sru-lp2127080-escape-equals-question",
			},
		},
		{
			name:        "valid SRU without description",
			ppaName:     "noble-systemd-sru-lp1234567",
			expectError: false,
			expected: &PPAInfo{
				Release:     "noble",
				Project:     "systemd",
				Type:        PPATypeSRU,
				BugID:       "1234567",
				Description: "",
				FullName:    "noble-systemd-sru-lp1234567",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := ParsePPAName(tt.ppaName)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.expectError && info != nil {
				if info.Release != tt.expected.Release {
					t.Errorf("release: expected %q, got %q", tt.expected.Release, info.Release)
				}
				if info.Type != tt.expected.Type {
					t.Errorf("type: expected %q, got %q", tt.expected.Type, info.Type)
				}
				if info.Description != tt.expected.Description {
					t.Errorf("description: expected %q, got %q", tt.expected.Description, info.Description)
				}
			}
		})
	}
}

// TestParsePPAName_Bug tests parsing normal bug PPA names
func TestParsePPAName_Bug(t *testing.T) {
	tests := []struct {
		name        string
		ppaName     string
		expectError bool
		expected    *PPAInfo
	}{
		{
			name:        "valid bug PPA with description",
			ppaName:     "noble-sudo-rs-lp2127080-fix-crash",
			expectError: false,
			expected: &PPAInfo{
				Release:     "noble",
				Project:     "sudo-rs",
				Type:        PPATypeBug,
				BugID:       "2127080",
				Description: "fix-crash",
				FullName:    "noble-sudo-rs-lp2127080-fix-crash",
			},
		},
		{
			name:        "valid bug PPA without description",
			ppaName:     "jammy-vim-lp9876543",
			expectError: false,
			expected: &PPAInfo{
				Release:     "jammy",
				Project:     "vim",
				Type:        PPATypeBug,
				BugID:       "9876543",
				Description: "",
				FullName:    "jammy-vim-lp9876543",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := ParsePPAName(tt.ppaName)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.expectError && info != nil {
				if info.Type != tt.expected.Type {
					t.Errorf("type: expected %q, got %q", tt.expected.Type, info.Type)
				}
				if info.BugID != tt.expected.BugID {
					t.Errorf("bugID: expected %q, got %q", tt.expected.BugID, info.BugID)
				}
			}
		})
	}
}

// TestParsePPAName_Invalid tests invalid PPA names
func TestParsePPAName_Invalid(t *testing.T) {
	invalid := []string{
		"",
		"invalid",
		"no-bug-id",
		"UPPERCASE-lp123",
		"noble-lp123",         // missing project
		"lp123-noble",         // wrong order
		"noble-project-",      // incomplete
		"noble-project-lp",    // no bug number
		"noble-project-lpXYZ", // non-numeric bug ID
	}

	for _, ppaName := range invalid {
		t.Run(ppaName, func(t *testing.T) {
			_, err := ParsePPAName(ppaName)
			if err == nil {
				t.Errorf("expected error for invalid PPA name %q", ppaName)
			}
		})
	}
}

// TestPPAInfo_GetPPATarget tests PPA target generation
func TestPPAInfo_GetPPATarget(t *testing.T) {
	tests := []struct {
		name     string
		fullName string
		username string
		expected string
	}{
		{
			name:     "merge PPA",
			fullName: "noble-efibootmgr-merge-lp2133493",
			username: "testuser",
			expected: "ppa:testuser/noble-efibootmgr-merge-lp2133493",
		},
		{
			name:     "SRU PPA with description",
			fullName: "jammy-sudo-rs-sru-lp2127080-escape-equals",
			username: "testuser",
			expected: "ppa:testuser/jammy-sudo-rs-sru-lp2127080-escape-equals",
		},
		{
			name:     "bug PPA",
			fullName: "noble-vim-lp1234567-fix-crash",
			username: "testuser",
			expected: "ppa:testuser/noble-vim-lp1234567-fix-crash",
		},
		{
			name:     "no username provided",
			fullName: "noble-test-lp123",
			username: "",
			expected: "ppa:$(whoami)/noble-test-lp123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &PPAInfo{
				FullName: tt.fullName,
			}
			target := info.GetPPATarget(tt.username)

			if target != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, target)
			}
		})
	}
}

// TestPPAInfo_GetBranchName tests git branch name generation
func TestPPAInfo_GetBranchName(t *testing.T) {
	tests := []struct {
		name     string
		ppaType  string
		bugID    string
		release  string
		expected string
	}{
		{"merge", PPATypeMerge, "2133493", "noble", "merge-lp2133493"},
		{"sru with release", PPATypeSRU, "2127080", "jammy", "sru-lp2127080-jammy"},
		{"bug with release", PPATypeBug, "1234567", "noble", "bug-lp1234567-noble"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &PPAInfo{
				Type:    tt.ppaType,
				BugID:   tt.bugID,
				Release: tt.release,
			}
			branch := info.GetBranchName()

			if branch != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, branch)
			}
		})
	}
}

// TestPPAInfo_GetChangelogMessage tests changelog message generation
func TestPPAInfo_GetChangelogMessage(t *testing.T) {
	tests := []struct {
		name        string
		info        *PPAInfo
		expectsText string
	}{
		{
			name: "merge without description",
			info: &PPAInfo{
				Type:  PPATypeMerge,
				BugID: "2133493",
			},
			expectsText: "LP: #2133493",
		},
		{
			name: "SRU with description",
			info: &PPAInfo{
				Type:        PPATypeSRU,
				BugID:       "2127080",
				Description: "escape-equals-question",
			},
			expectsText: "escape equals question",
		},
		{
			name: "bug fix",
			info: &PPAInfo{
				Type:  PPATypeBug,
				BugID: "1234567",
			},
			expectsText: "Bug fix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.info.GetChangelogMessage()

			if !contains(msg, tt.expectsText) {
				t.Errorf("expected message to contain %q, got %q", tt.expectsText, msg)
			}

			// All messages should reference the bug
			bugRef := "LP: #" + tt.info.BugID
			if !contains(msg, bugRef) {
				t.Errorf("expected message to contain %q, got %q", bugRef, msg)
			}
		})
	}
}

// TestPPAInfo_GetVersionSuffix tests version suffix generation
func TestPPAInfo_GetVersionSuffix(t *testing.T) {
	tests := []struct {
		name           string
		release        string
		currentVersion string
		expected       string
	}{
		{
			name:           "first upload",
			release:        "noble",
			currentVersion: "1.0-1",
			expected:       "~noble1",
		},
		{
			name:           "increment from noble1",
			release:        "noble",
			currentVersion: "1.0-1~noble1",
			expected:       "~noble2",
		},
		{
			name:           "increment from noble5",
			release:        "noble",
			currentVersion: "1.0-1~noble5",
			expected:       "~noble6",
		},
		{
			name:           "different release",
			release:        "jammy",
			currentVersion: "1.0-1",
			expected:       "~jammy1",
		},
		{
			name:           "increment jammy",
			release:        "jammy",
			currentVersion: "1.0-1~jammy3",
			expected:       "~jammy4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &PPAInfo{Release: tt.release}
			suffix := info.GetVersionSuffix(tt.currentVersion)

			if suffix != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, suffix)
			}
		})
	}
}

// TestPPAInfo_String tests string representation
func TestPPAInfo_String(t *testing.T) {
	info := &PPAInfo{
		Release:     "noble",
		Project:     "sudo-rs",
		Type:        PPATypeSRU,
		BugID:       "2127080",
		Description: "fix-crash",
		FullName:    "noble-sudo-rs-sru-lp2127080-fix-crash",
	}

	output := info.String()

	expectedParts := []string{
		"noble",
		"sudo-rs",
		"sru",
		"LP#2127080",
		"fix-crash",
		"sru-lp2127080",
	}

	for _, part := range expectedParts {
		if !contains(output, part) {
			t.Errorf("expected output to contain %q\nGot: %s", part, output)
		}
	}
}

// TestIsInPackagingDir tests packaging directory detection
func TestIsInPackagingDir(t *testing.T) {
	// Save current directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)

	tests := []struct {
		name     string
		setup    func(string)
		expected bool
	}{
		{
			name: "with debian/control",
			setup: func(dir string) {
				os.MkdirAll(filepath.Join(dir, "debian"), 0755)
				os.WriteFile(filepath.Join(dir, "debian", "control"), []byte("test"), 0644)
			},
			expected: true,
		},
		{
			name: "with debian/changelog",
			setup: func(dir string) {
				os.MkdirAll(filepath.Join(dir, "debian"), 0755)
				os.WriteFile(filepath.Join(dir, "debian", "changelog"), []byte("test"), 0644)
			},
			expected: true,
		},
		{
			name: "not a packaging dir",
			setup: func(dir string) {
				// Empty directory
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tt.setup(tmpDir)
			os.Chdir(tmpDir)

			result := IsInPackagingDir()

			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestDetectUbuntuRelease tests Ubuntu release detection from changelog
func TestDetectUbuntuRelease(t *testing.T) {
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)

	tests := []struct {
		name        string
		changelog   string
		expected    string
		expectError bool
	}{
		{
			name:        "noble release",
			changelog:   "sudo-rs (0.2.3-1ubuntu1) noble; urgency=medium\n\n  * Test\n",
			expected:    "noble",
			expectError: false,
		},
		{
			name:        "jammy release",
			changelog:   "vim (2:9.0.1000-1) jammy; urgency=low\n\n  * Update\n",
			expected:    "jammy",
			expectError: false,
		},
		{
			name:        "focal release",
			changelog:   "package (1.0-1) focal; urgency=high\n\n  * Fix\n",
			expected:    "focal",
			expectError: false,
		},
		{
			name:        "malformed changelog",
			changelog:   "invalid format",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			os.Chdir(tmpDir)

			os.MkdirAll("debian", 0755)
			os.WriteFile("debian/changelog", []byte(tt.changelog), 0644)

			release, err := DetectUbuntuRelease()

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.expectError && release != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, release)
			}
		})
	}
}

// TestCreatePPAName tests PPA name generation
func TestCreatePPAName(t *testing.T) {
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)

	// Setup test environment with debian/changelog
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	os.MkdirAll("debian", 0755)
	changelog := "sudo-rs (0.2.3-1) noble; urgency=medium\n\n  * Test\n"
	os.WriteFile("debian/changelog", []byte(changelog), 0644)

	tests := []struct {
		name        string
		project     string
		bugID       string
		ppaType     string
		description string
		expected    string
		expectError bool
	}{
		{
			name:        "merge PPA",
			project:     "efibootmgr",
			bugID:       "2133493",
			ppaType:     "merge",
			description: "",
			expected:    "noble-efibootmgr-merge-lp2133493",
			expectError: false,
		},
		{
			name:        "SRU with description",
			project:     "sudo-rs",
			bugID:       "2127080",
			ppaType:     "sru",
			description: "escape equals question",
			expected:    "noble-sudo-rs-sru-lp2127080-escape-equals-question",
			expectError: false,
		},
		{
			name:        "bug PPA with description",
			project:     "vim",
			bugID:       "1234567",
			ppaType:     "bug",
			description: "fix crash",
			expected:    "noble-vim-lp1234567-fix-crash",
			expectError: false,
		},
		{
			name:        "bug PPA without description",
			project:     "systemd",
			bugID:       "lp9999999",
			ppaType:     "bug",
			description: "",
			expected:    "noble-systemd-lp9999999",
			expectError: false,
		},
		{
			name:        "short type aliases - merge",
			project:     "test",
			bugID:       "123",
			ppaType:     "m",
			description: "",
			expected:    "noble-test-merge-lp123",
			expectError: false,
		},
		{
			name:        "short type aliases - sru",
			project:     "test",
			bugID:       "456",
			ppaType:     "s",
			description: "desc",
			expected:    "noble-test-sru-lp456-desc",
			expectError: false,
		},
		{
			name:        "missing project",
			project:     "",
			bugID:       "123",
			ppaType:     "bug",
			description: "",
			expected:    "",
			expectError: true,
		},
		{
			name:        "missing bug ID",
			project:     "test",
			bugID:       "",
			ppaType:     "bug",
			description: "",
			expected:    "",
			expectError: true,
		},
		{
			name:        "invalid bug ID",
			project:     "test",
			bugID:       "notanumber",
			ppaType:     "bug",
			description: "",
			expected:    "",
			expectError: true,
		},
		{
			name:        "invalid type",
			project:     "test",
			bugID:       "123",
			ppaType:     "invalid",
			description: "",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ppaName, err := CreatePPAName(tt.project, tt.bugID, tt.ppaType, tt.description)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.expectError && ppaName != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, ppaName)
			}
		})
	}
}
