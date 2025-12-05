package plugin

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/bamf0/toolbox/internal/config"
)

//go:embed ubuntu_helpers.sh
var ubuntuHelpersScript string

// UbuntuPlugin provides Ubuntu/Debian packaging support with PPA workflows
type UbuntuPlugin struct {
	name    string
	version string
}

// PPAInfo contains parsed PPA metadata
type PPAInfo struct {
	Release     string // Ubuntu release (e.g., "noble", "jammy")
	Project     string // Project name
	Type        string // "merge", "sru", or "bug"
	BugID       string // Bug ID (e.g., "2133493")
	Description string // Optional description
	FullName    string // Original PPA name
}

// PPAType constants
const (
	PPATypeMerge = "merge"
	PPATypeSRU   = "sru"
	PPATypeBug   = "bug"
)

// NewUbuntuPlugin creates a new Ubuntu packaging plugin
func NewUbuntuPlugin() *UbuntuPlugin {
	return &UbuntuPlugin{
		name:    "ubuntu",
		version: "1.0.0",
	}
}

// getEmbeddedScriptPath writes the embedded script to a temporary location and returns its path
func getEmbeddedScriptPath() string {
	// Create cache directory in user's home
	cacheDir := filepath.Join(os.Getenv("HOME"), ".cache", "toolbox")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		// Fallback to temp directory if home cache is not available
		cacheDir = os.TempDir()
	}

	scriptPath := filepath.Join(cacheDir, "ubuntu_helpers.sh")

	// Write the embedded script to the cache location
	// Only write if file doesn't exist or is outdated
	if err := os.WriteFile(scriptPath, []byte(ubuntuHelpersScript), 0755); err != nil {
		// If we can't write, return a path that will fail gracefully
		return "ubuntu_helpers.sh"
	}

	return scriptPath
}

func (p *UbuntuPlugin) Name() string {
	return p.name
}

func (p *UbuntuPlugin) Version() string {
	return p.version
}

func (p *UbuntuPlugin) Contexts() map[string]config.ContextConfig {
	// Write embedded script to a temporary file
	scriptPath := getEmbeddedScriptPath()

	return map[string]config.ContextConfig{
		"ubuntu-packaging": {
			Commands: map[string]string{
				// Branch creation (takes arguments: project, bug-id, type, description)
				"gbranch": fmt.Sprintf("bash %s gbranch", scriptPath),

				// PPA-aware commands (infer from current branch)
				"ppa-status":  fmt.Sprintf("bash %s ppa-status", scriptPath),
				"ppa-migrate": fmt.Sprintf("bash %s ppa-migrate", scriptPath),
				"dch-auto":    fmt.Sprintf("bash %s dch-auto", scriptPath),
				"ubuild":      fmt.Sprintf("bash %s ubuild", scriptPath),
				"sb-auto":     fmt.Sprintf("bash %s sb-auto", scriptPath),
				"dput-auto":   fmt.Sprintf("bash %s dput-auto", scriptPath),

				// Standard changelog commands
				"dch":         "dch -i",
				"dch-release": "dch -r",

				// Build commands
				"build":        "dpkg-buildpackage -us -uc",
				"build-source": "dpkg-buildpackage -S -us -uc",

				// Status and info
				"changelog": "dpkg-parsechangelog",
				"version":   "dpkg-parsechangelog -S Version",

				// Clean commands
				"clean":     "debian/rules clean",
				"distclean": "fakeroot debian/rules clean",

				// Linting
				"lint":         "lintian",
				"lint-source":  "lintian --pedantic *.dsc",
				"lint-changes": "lintian --pedantic *.changes",
			},
			Descriptions: map[string]string{
				// Branch and PPA management
				"gbranch":     "Create/checkout git branch: gbranch <project> <bug-id> [merge|sru|bug] [description] [release]",
				"ppa-status":  "Show PPA information from current branch",
				"ppa-migrate": "Migrate stored PPA names from old format to new format",

				// Changelog commands
				"dch-auto":    "Auto-update changelog with version suffix from current branch",
				"dch":         "Add new changelog entry manually",
				"dch-release": "Mark changelog entry as released",

				// Build and upload
				"ubuild":    "Complete build and upload workflow (sb-auto + dput-auto)",
				"sb-auto":   "Build source package with sbuild for detected release",
				"dput-auto": "Upload to PPA inferred from current branch",

				// Standard builds
				"build":        "Build binary package (dpkg-buildpackage)",
				"build-source": "Build source package only",

				// Info and status
				"changelog": "Display full changelog",
				"version":   "Show current package version",

				// Cleanup
				"clean":     "Clean build artifacts",
				"distclean": "Deep clean (using fakeroot)",

				// Quality checks
				"lint":         "Run lintian on built packages",
				"lint-source":  "Run lintian on source package",
				"lint-changes": "Run lintian on .changes file",
			},
		},
	}
}

func (p *UbuntuPlugin) Detect(dir string) (string, bool) {
	// Check for debian/control - the primary indicator
	controlFile := filepath.Join(dir, "debian", "control")
	if _, err := os.Stat(controlFile); err == nil {
		return "ubuntu-packaging", true
	}

	// Check for debian/changelog
	changelogFile := filepath.Join(dir, "debian", "changelog")
	if _, err := os.Stat(changelogFile); err == nil {
		return "ubuntu-packaging", true
	}

	return "", false
}

func (p *UbuntuPlugin) Validate() error {
	if p.name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}
	if p.version == "" {
		return fmt.Errorf("plugin version cannot be empty")
	}

	contexts := p.Contexts()
	if len(contexts) == 0 {
		return fmt.Errorf("plugin must provide at least one context")
	}

	for ctxName, ctxConfig := range contexts {
		if len(ctxConfig.Commands) == 0 {
			return fmt.Errorf("context %q has no commands", ctxName)
		}
	}

	return nil
}

// ParsePPAName parses a PPA name into its components
// Formats:
//   - Merge: <release>-<project>-merge-lp<bug>
//   - SRU: <release>-<project>-sru-lp<bug>-<desc>
//   - Bug: <release>-<project>-lp<bug>-<desc>
func ParsePPAName(ppaName string) (*PPAInfo, error) {
	if ppaName == "" {
		return nil, fmt.Errorf("PPA name cannot be empty")
	}

	// Regex patterns for different PPA types
	// New format: <project>-<type>-lp<bug>-<release>
	// Merge: efibootmgr-merge-lp2133493-noble
	mergePattern := regexp.MustCompile(`^([a-z0-9\-]+)-merge-lp(\d+)-([a-z]+)$`)

	// SRU: sudo-rs-sru-lp2127080-jammy or sudo-rs-sru-lp2127080-escape-equals-jammy
	sruPattern := regexp.MustCompile(`^([a-z0-9\-]+)-sru-lp(\d+)-(.+)-([a-z]+)$|^([a-z0-9\-]+)-sru-lp(\d+)-([a-z]+)$`)

	// Normal bug: sudo-rs-lp2127080-noble or sudo-rs-lp2127080-description-noble
	bugPattern := regexp.MustCompile(`^([a-z0-9\-]+)-lp(\d+)-(.+)-([a-z]+)$|^([a-z0-9\-]+)-lp(\d+)-([a-z]+)$`)

	ppaName = strings.TrimSpace(ppaName)

	// Try merge pattern first: <project>-merge-lp<bug>-<release>
	if matches := mergePattern.FindStringSubmatch(ppaName); matches != nil {
		return &PPAInfo{
			Project:     matches[1],
			BugID:       matches[2],
			Release:     matches[3],
			Type:        PPATypeMerge,
			Description: "",
			FullName:    ppaName,
		}, nil
	}

	// Try SRU pattern: <project>-sru-lp<bug>-<release> or <project>-sru-lp<bug>-<desc>-<release>
	if matches := sruPattern.FindStringSubmatch(ppaName); matches != nil {
		// Check which pattern matched (with or without description)
		if matches[1] != "" {
			// Pattern with description: matches[1]=project, [2]=bug, [3]=desc, [4]=release
			return &PPAInfo{
				Project:     matches[1],
				BugID:       matches[2],
				Description: matches[3],
				Release:     matches[4],
				Type:        PPATypeSRU,
				FullName:    ppaName,
			}, nil
		} else {
			// Pattern without description: matches[5]=project, [6]=bug, [7]=release
			return &PPAInfo{
				Project:     matches[5],
				BugID:       matches[6],
				Release:     matches[7],
				Type:        PPATypeSRU,
				Description: "",
				FullName:    ppaName,
			}, nil
		}
	}

	// Try normal bug pattern: <project>-lp<bug>-<release> or <project>-lp<bug>-<desc>-<release>
	if matches := bugPattern.FindStringSubmatch(ppaName); matches != nil {
		// Check which pattern matched
		if matches[1] != "" {
			// Pattern with description: matches[1]=project, [2]=bug, [3]=desc, [4]=release
			return &PPAInfo{
				Project:     matches[1],
				BugID:       matches[2],
				Description: matches[3],
				Release:     matches[4],
				Type:        PPATypeBug,
				FullName:    ppaName,
			}, nil
		} else {
			// Pattern without description: matches[5]=project, [6]=bug, [7]=release
			return &PPAInfo{
				Project:     matches[5],
				BugID:       matches[6],
				Release:     matches[7],
				Type:        PPATypeBug,
				Description: "",
				FullName:    ppaName,
			}, nil
		}
	}

	return nil, fmt.Errorf("invalid PPA name format: %s", ppaName)
}

// GetPPATarget returns the correct PPA target using the full PPA name
func (info *PPAInfo) GetPPATarget(username string) string {
	if username == "" {
		username = "$(whoami)"
	}

	// The PPA name IS the full descriptive name
	// e.g., ppa:username/jammy-sudo-rs-sru-lp2127080-escape-equals
	return fmt.Sprintf("ppa:%s/%s", username, info.FullName)
}

// GetBranchName returns the appropriate git branch name
// For SRU and Bug: <type>-lp<bug>-<release>
// For Merge: merge-lp<bug> (no release suffix)
func (info *PPAInfo) GetBranchName() string {
	switch info.Type {
	case PPATypeMerge:
		return fmt.Sprintf("merge-lp%s", info.BugID)
	case PPATypeSRU:
		return fmt.Sprintf("sru-lp%s-%s", info.BugID, info.Release)
	case PPATypeBug:
		return fmt.Sprintf("bug-lp%s-%s", info.BugID, info.Release)
	default:
		return fmt.Sprintf("lp%s-%s", info.BugID, info.Release)
	}
}

// CreatePPAName generates a PPA name from components
// For merge type, optionalRelease should be provided (merges target future releases)
// For SRU/bug types, optionalRelease can be empty (will use debian/changelog)
func CreatePPAName(project, bugID, ppaType, description, optionalRelease string) (string, error) {
	if project == "" {
		return "", fmt.Errorf("project name is required")
	}
	if bugID == "" {
		return "", fmt.Errorf("bug ID is required")
	}

	// Clean bug ID - strip "lp" prefix if present
	bugID = strings.TrimPrefix(strings.TrimSpace(bugID), "lp")
	if _, err := strconv.Atoi(bugID); err != nil {
		return "", fmt.Errorf("invalid bug ID format: %s", bugID)
	}

	// Normalize inputs
	project = strings.ToLower(strings.TrimSpace(project))
	ppaType = strings.ToLower(strings.TrimSpace(ppaType))
	description = strings.ToLower(strings.TrimSpace(description))
	optionalRelease = strings.ToLower(strings.TrimSpace(optionalRelease))

	// Replace spaces with hyphens in description
	description = strings.ReplaceAll(description, " ", "-")

	// Determine which release to use
	var release string
	if optionalRelease != "" {
		// Use provided release (for merges, or when overriding)
		release = optionalRelease
	} else {
		// Detect release from debian/changelog (for SRU/bug)
		detectedRelease, err := DetectUbuntuRelease()
		if err != nil {
			return "", fmt.Errorf("could not detect Ubuntu release: %w (are you in a debian packaging directory?)", err)
		}
		release = detectedRelease
	}

	// For merge type, release is required
	if (ppaType == PPATypeMerge || ppaType == "m") && optionalRelease == "" {
		return "", fmt.Errorf("merge branches require a release parameter (the target release)")
	}

	// For merge type, description is not allowed
	if (ppaType == PPATypeMerge || ppaType == "m") && description != "" {
		return "", fmt.Errorf("merge branches cannot have a description")
	}

	// Build PPA name based on type
	var ppaName string
	switch ppaType {
	case PPATypeMerge, "m":
		// Format: <project>-merge-lp<bug>-<release>
		ppaName = fmt.Sprintf("%s-merge-lp%s-%s", project, bugID, release)

	case PPATypeSRU, "s":
		// Format: <project>-sru-lp<bug>-<release> or <project>-sru-lp<bug>-<desc>-<release>
		if description != "" {
			ppaName = fmt.Sprintf("%s-sru-lp%s-%s-%s", project, bugID, description, release)
		} else {
			ppaName = fmt.Sprintf("%s-sru-lp%s-%s", project, bugID, release)
		}

	case PPATypeBug, "b", "":
		// Format: <project>-lp<bug>-<release> or <project>-lp<bug>-<desc>-<release>
		if description != "" {
			ppaName = fmt.Sprintf("%s-lp%s-%s-%s", project, bugID, description, release)
		} else {
			ppaName = fmt.Sprintf("%s-lp%s-%s", project, bugID, release)
		}

	default:
		return "", fmt.Errorf("invalid PPA type: %s (use 'merge', 'sru', or 'bug')", ppaType)
	}

	return ppaName, nil
}

// ParseBranchName extracts PPA information from a git branch name
// Branch formats:
//   - Merge: merge-lp<bug>-<release> (release required, no description allowed)
//   - SRU: sru-lp<bug>-<release>
//   - Bug: bug-lp<bug>-<release> or lp<bug>-<release>
func ParseBranchName(branchName string) (*PPAInfo, error) {
	if branchName == "" {
		return nil, fmt.Errorf("branch name cannot be empty")
	}

	branchName = strings.TrimSpace(branchName)

	// Check for merge branch: merge-lp2133493-noble (requires release)
	mergePattern := regexp.MustCompile(`^merge-lp(\d+)-([a-z]+)$`)
	if matches := mergePattern.FindStringSubmatch(branchName); matches != nil {
		// For merge branches, use the release from branch name
		bugID := matches[1]
		release := matches[2]

		// Get project name from debian/control
		project, err := DetectProjectName()
		if err != nil {
			return nil, fmt.Errorf("could not detect project name: %w", err)
		}

		// Construct PPA name: <project>-merge-lp<bug>-<release>
		ppaName := fmt.Sprintf("%s-merge-lp%s-%s", project, bugID, release)

		return &PPAInfo{
			Release:     release,
			Project:     project,
			Type:        PPATypeMerge,
			BugID:       bugID,
			Description: "",
			FullName:    ppaName,
		}, nil
	}

	// Check for SRU branch: sru-lp2127080-jammy
	sruPattern := regexp.MustCompile(`^sru-lp(\d+)-([a-z]+)$`)
	if matches := sruPattern.FindStringSubmatch(branchName); matches != nil {
		bugID := matches[1]
		release := matches[2]
		
		project, err := DetectProjectName()
		if err != nil {
			return nil, fmt.Errorf("could not detect project name: %w", err)
		}

		// Construct PPA name: <project>-sru-lp<bug>-<release>
		ppaName := fmt.Sprintf("%s-sru-lp%s-%s", project, bugID, release)

		return &PPAInfo{
			Release:     release,
			Project:     project,
			Type:        PPATypeSRU,
			BugID:       bugID,
			Description: "",
			FullName:    ppaName,
		}, nil
	}

	// Check for bug branch: bug-lp2127080-jammy or lp2127080-jammy
	bugPattern := regexp.MustCompile(`^(?:bug-)?lp(\d+)-([a-z]+)$`)
	if matches := bugPattern.FindStringSubmatch(branchName); matches != nil {
		bugID := matches[1]
		release := matches[2]
		
		project, err := DetectProjectName()
		if err != nil {
			return nil, fmt.Errorf("could not detect project name: %w", err)
		}

		// Construct PPA name: <project>-lp<bug>-<release>
		ppaName := fmt.Sprintf("%s-lp%s-%s", project, bugID, release)

		return &PPAInfo{
			Release:     release,
			Project:     project,
			Type:        PPATypeBug,
			BugID:       bugID,
			Description: "",
			FullName:    ppaName,
		}, nil
	}

	return nil, fmt.Errorf("branch name does not contain a valid Launchpad bug ID: %s", branchName)
}

// GetCurrentBranch returns the current git branch name
func GetCurrentBranch() (string, error) {
	// This is a placeholder - in practice, this would be called from the shell script
	// which has access to git commands
	return "", fmt.Errorf("not implemented in Go - use shell helper")
}

// DetectProjectName reads the project name from debian/control
func DetectProjectName() (string, error) {
	controlPath := "debian/control"

	data, err := os.ReadFile(controlPath)
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %w", controlPath, err)
	}

	// Parse Source: line
	pattern := regexp.MustCompile(`(?m)^Source:\s+(.+)$`)
	matches := pattern.FindSubmatch(data)

	if len(matches) < 2 {
		return "", fmt.Errorf("could not parse Source from debian/control")
	}

	project := strings.TrimSpace(string(matches[1]))
	return project, nil
}

// DetectUbuntuRelease reads the current Ubuntu release from debian/changelog
func DetectUbuntuRelease() (string, error) {
	changelogPath := "debian/changelog"

	data, err := os.ReadFile(changelogPath)
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %w", changelogPath, err)
	}

	// Parse first line: package (version) release; urgency=...
	// Example: sudo-rs (0.2.3-1ubuntu1) noble; urgency=medium
	pattern := regexp.MustCompile(`^\S+\s+\([^)]+\)\s+([a-z]+);`)
	matches := pattern.FindSubmatch(data)

	if len(matches) < 2 {
		return "", fmt.Errorf("could not parse release from debian/changelog")
	}

	release := string(matches[1])
	return release, nil
}

// GetChangelogMessage returns a changelog entry message
func (info *PPAInfo) GetChangelogMessage() string {
	bugRef := fmt.Sprintf("LP: #%s", info.BugID)

	if info.Description != "" {
		// Convert hyphens to spaces for description
		desc := strings.ReplaceAll(info.Description, "-", " ")
		return fmt.Sprintf("* %s: %s", desc, bugRef)
	}

	switch info.Type {
	case PPATypeMerge:
		return fmt.Sprintf("* Merge from Debian. %s", bugRef)
	case PPATypeSRU:
		return fmt.Sprintf("* SRU update. %s", bugRef)
	default:
		return fmt.Sprintf("* Bug fix. %s", bugRef)
	}
}

// GetVersionSuffix returns the version suffix for this release
// Format: ~<release><n> where n starts at 1 and increments
func (info *PPAInfo) GetVersionSuffix(currentVersion string) string {
	// Extract current suffix number if present
	pattern := regexp.MustCompile(`~` + regexp.QuoteMeta(info.Release) + `(\d+)`)
	matches := pattern.FindStringSubmatch(currentVersion)

	n := 1
	if len(matches) > 1 {
		if num, err := strconv.Atoi(matches[1]); err == nil {
			n = num + 1
		}
	}

	return fmt.Sprintf("~%s%d", info.Release, n)
}

// String returns a formatted summary of PPA info
func (info *PPAInfo) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("PPA: %s\n", info.FullName))
	sb.WriteString(fmt.Sprintf("  Release: %s\n", info.Release))
	sb.WriteString(fmt.Sprintf("  Project: %s\n", info.Project))
	sb.WriteString(fmt.Sprintf("  Type: %s\n", info.Type))
	sb.WriteString(fmt.Sprintf("  Bug ID: LP#%s\n", info.BugID))

	if info.Description != "" {
		sb.WriteString(fmt.Sprintf("  Description: %s\n", info.Description))
	}

	sb.WriteString(fmt.Sprintf("  Branch: %s\n", info.GetBranchName()))
	sb.WriteString(fmt.Sprintf("  PPA Target: %s\n", info.GetPPATarget("")))

	return sb.String()
}

// IsInPackagingDir checks if we're in a Debian/Ubuntu packaging directory
func IsInPackagingDir() bool {
	if _, err := os.Stat("debian/control"); err == nil {
		return true
	}
	if _, err := os.Stat("debian/changelog"); err == nil {
		return true
	}
	return false
}
