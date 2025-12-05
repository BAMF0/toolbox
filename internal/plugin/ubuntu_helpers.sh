#!/bin/bash
# Ubuntu packaging helper scripts for ToolBox
# These implement the dynamic PPA-aware commands
#
# NOTE: This is a copy of ../../scripts/ubuntu_helpers.sh that is embedded into the binary.
# See .ubuntu_helpers_note.md for details on keeping it in sync.

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Error handling
error() {
    echo -e "${RED}Error: $1${NC}" >&2
    exit 1
}

warn() {
    echo -e "${YELLOW}Warning: $1${NC}" >&2
}

info() {
    echo -e "${BLUE}Info: $1${NC}"
}

success() {
    echo -e "${GREEN}$1${NC}"
}

# Check if we're in a packaging directory
check_packaging_dir() {
    if [[ ! -f "debian/control" ]] && [[ ! -f "debian/changelog" ]]; then
        error "Not in a Debian/Ubuntu packaging directory (debian/control or debian/changelog not found)"
    fi
}

# Get current git branch
get_current_branch() {
    git rev-parse --abbrev-ref HEAD 2>/dev/null
}

# Verify branch contains a Launchpad bug ID
verify_branch_has_bug() {
    local branch="$1"
    
    if [[ -z "$branch" ]]; then
        error "Not on a git branch"
    fi
    
    # Check if branch contains lp<digits>
    if [[ ! "$branch" =~ lp[0-9]+ ]]; then
        error "Current branch '$branch' does not contain a Launchpad bug ID (lp<number>).
Use 'gbranch' to create a proper branch first:
  gbranch <project> <bug-id> [type] [description]"
    fi
}

# Detect project name from debian/control
detect_project() {
    check_packaging_dir
    
    local control="debian/control"
    local project=$(grep -m1 "^Source:" "$control" | awk '{print $2}' | tr -d '\r\n')
    
    if [[ -z "$project" ]]; then
        error "Could not detect project name from debian/control"
    fi
    
    echo "$project"
}

# Parse branch name into PPA components
# Branch formats:
#   - Merge: merge-lp<bug>
#   - SRU: sru-lp<bug>-<release>
#   - Bug: bug-lp<bug>-<release> or lp<bug>-<release>
parse_branch() {
    local branch="$1"
    
    if [[ -z "$branch" ]]; then
        error "Branch name required"
    fi
    
    # Verify branch has bug ID first
    verify_branch_has_bug "$branch"
    
    # Try to get PPA name from git config first (this preserves description)
    local stored_ppa_name=$(git config "branch.${branch}.ppaname" 2>/dev/null || echo "")
    if [[ -n "$stored_ppa_name" ]]; then
        # Parse the stored PPA name to get all components
        PPA_FULL_NAME="$stored_ppa_name"
        
        # Extract components from stored PPA name
        # Formats: 
        #   Merge: <release>-<project>-merge-lp<bug>
        #   SRU:   <release>-<project>-sru-lp<bug>[-<desc>]
        #   Bug:   <release>-<project>-lp<bug>[-<desc>]
        
        if [[ "$stored_ppa_name" =~ ^([a-z]+)-(.+)-merge-lp([0-9]+)$ ]]; then
            PPA_RELEASE="${BASH_REMATCH[1]}"
            PPA_PROJECT="${BASH_REMATCH[2]}"
            PPA_BUGID="${BASH_REMATCH[3]}"
            PPA_TYPE="merge"
            PPA_DESC=""
            return 0
        elif [[ "$stored_ppa_name" =~ ^([a-z]+)-(.+)-sru-lp([0-9]+)(-(.+))?$ ]]; then
            PPA_RELEASE="${BASH_REMATCH[1]}"
            PPA_PROJECT="${BASH_REMATCH[2]}"
            PPA_BUGID="${BASH_REMATCH[3]}"
            PPA_DESC="${BASH_REMATCH[5]}"
            PPA_TYPE="sru"
            return 0
        elif [[ "$stored_ppa_name" =~ ^([a-z]+)-(.+)-lp([0-9]+)(-(.+))?$ ]]; then
            PPA_RELEASE="${BASH_REMATCH[1]}"
            PPA_PROJECT="${BASH_REMATCH[2]}"
            PPA_BUGID="${BASH_REMATCH[3]}"
            PPA_DESC="${BASH_REMATCH[5]}"
            PPA_TYPE="bug"
            return 0
        fi
    fi
    
    # Fallback: detect release and project from packaging files
    local release=$(detect_release)
    local project=$(detect_project)
    
    # Check for merge branch: merge-lp2133493
    if [[ "$branch" =~ ^merge-lp([0-9]+)$ ]]; then
        PPA_BUGID="${BASH_REMATCH[1]}"
        PPA_RELEASE="$release"
        PPA_PROJECT="$project"
        PPA_TYPE="merge"
        PPA_DESC=""
        PPA_FULL_NAME="${release}-${project}-merge-lp${PPA_BUGID}"
        return 0
    fi
    
    # Check for SRU branch: sru-lp2127080-jammy
    if [[ "$branch" =~ ^sru-lp([0-9]+)-([a-z]+)$ ]]; then
        PPA_BUGID="${BASH_REMATCH[1]}"
        PPA_RELEASE="${BASH_REMATCH[2]}"
        PPA_PROJECT="$project"
        PPA_TYPE="sru"
        PPA_DESC=""
        PPA_FULL_NAME="${PPA_RELEASE}-${project}-sru-lp${PPA_BUGID}"
        return 0
    fi
    
    # Check for bug branch: bug-lp2127080-jammy or lp2127080-jammy
    if [[ "$branch" =~ ^(bug-)?lp([0-9]+)-([a-z]+)$ ]]; then
        PPA_BUGID="${BASH_REMATCH[2]}"
        PPA_RELEASE="${BASH_REMATCH[3]}"
        PPA_PROJECT="$project"
        PPA_TYPE="bug"
        PPA_DESC=""
        PPA_FULL_NAME="${PPA_RELEASE}-${project}-lp${PPA_BUGID}"
        return 0
    fi
    
    error "Invalid branch format: $branch"
}

# Parse from current branch (convenience wrapper)
parse_current_branch() {
    local branch=$(get_current_branch)
    parse_branch "$branch"
}

# Get PPA target using full PPA name
# e.g., ppa:username/jammy-sudo-rs-sru-lp2127080
get_ppa_target() {
    local username="${1:-$(whoami)}"
    
    # PPA_FULL_NAME is set by parse_branch or parse_current_branch
    if [[ -z "$PPA_FULL_NAME" ]]; then
        error "PPA_FULL_NAME not set - call parse_current_branch first"
    fi
    
    echo "ppa:$username/$PPA_FULL_NAME"
}

# Get git branch name
# For merge: merge-lp<bug>
# For sru/bug: <type>-lp<bug>-<release>
get_branch_name() {
    case "$PPA_TYPE" in
        merge)
            echo "merge-lp${PPA_BUGID}"
            ;;
        sru)
            echo "sru-lp${PPA_BUGID}-${PPA_RELEASE}"
            ;;
        bug)
            echo "bug-lp${PPA_BUGID}-${PPA_RELEASE}"
            ;;
        *)
            echo "lp${PPA_BUGID}-${PPA_RELEASE}"
            ;;
    esac
}

# Get changelog message
get_changelog_message() {
    local bug_ref="LP: #${PPA_BUGID}"
    
    if [[ -n "$PPA_DESC" ]]; then
        # Convert hyphens to spaces
        local desc="${PPA_DESC//-/ }"
        echo "* ${desc}. ${bug_ref}"
    else
        case "$PPA_TYPE" in
            merge)
                echo "* Merge from Debian. ${bug_ref}"
                ;;
            sru)
                echo "* SRU update. ${bug_ref}"
                ;;
            *)
                echo "* Bug fix. ${bug_ref}"
                ;;
        esac
    fi
}

# Get version suffix with auto-increment
get_version_suffix() {
    local current_version="$1"
    
    # Extract current number if present
    if [[ "$current_version" =~ ~${PPA_RELEASE}([0-9]+) ]]; then
        local current_num="${BASH_REMATCH[1]}"
        local next_num=$((current_num + 1))
        echo "~${PPA_RELEASE}${next_num}"
    else
        echo "~${PPA_RELEASE}1"
    fi
}

# Find latest .changes file
find_latest_changes() {
    local changes_file=$(ls -t ../*.changes 2>/dev/null | head -1)
    if [[ -z "$changes_file" ]]; then
        error "No .changes file found in parent directory"
    fi
    echo "$changes_file"
}

# Detect Ubuntu release from debian/changelog
detect_release() {
    check_packaging_dir
    
    # Parse first line: package (version) release; urgency=...
    local changelog="debian/changelog"
    local release=$(head -n1 "$changelog" | sed -n 's/.*) \([a-z]*\);.*/\1/p')
    
    if [[ -z "$release" ]]; then
        error "Could not detect Ubuntu release from debian/changelog"
    fi
    
    echo "$release"
}

# Command: gbranch - Create/checkout git branch (ONLY command that takes arguments)
cmd_gbranch() {
    check_packaging_dir
    
    local project="$1"
    local bug_id="$2"
    local ppa_type="${3:-bug}"  # Default to bug
    local description="$4"
    
    if [[ -z "$project" ]]; then
        error "Usage: gbranch <project> <bug-id> [type] [description]
  project:     Project name (e.g., sudo-rs, efibootmgr)
  bug-id:      Launchpad bug ID (e.g., 2127080 or lp2127080)
  type:        PPA type: merge|m, sru|s, bug|b (default: bug)
  description: Optional description for the PPA

Examples:
  gbranch sudo-rs 2127080 sru escape-equals
  gbranch efibootmgr 2133493 merge
  gbranch myproject 123456 bug test-fix"
    fi
    
    if [[ -z "$bug_id" ]]; then
        error "Bug ID is required"
    fi
    
    # Clean bug ID - remove lp prefix if present
    bug_id="${bug_id#lp}"
    
    # Validate bug ID is numeric
    if ! [[ "$bug_id" =~ ^[0-9]+$ ]]; then
        error "Invalid bug ID: $bug_id (must be numeric)"
    fi
    
    # Detect current release from debian/changelog
    local release=$(detect_release)
    
    # Normalize inputs
    project=$(echo "$project" | tr '[:upper:]' '[:lower:]' | tr ' ' '-')
    ppa_type=$(echo "$ppa_type" | tr '[:upper:]' '[:lower:]')
    description=$(echo "$description" | tr '[:upper:]' '[:lower:]' | tr ' ' '-')
    
    # Determine branch name based on type
    local branch_name=""
    local ppa_name=""
    
    case "$ppa_type" in
        merge|m)
            # Branch: merge-lp<bug>
            branch_name="merge-lp${bug_id}"
            ppa_name="${release}-${project}-merge-lp${bug_id}"
            ;;
        sru|s)
            # Branch: sru-lp<bug>-<release>
            branch_name="sru-lp${bug_id}-${release}"
            if [[ -n "$description" ]]; then
                ppa_name="${release}-${project}-sru-lp${bug_id}-${description}"
            else
                ppa_name="${release}-${project}-sru-lp${bug_id}"
            fi
            ;;
        bug|b|"")
            # Branch: bug-lp<bug>-<release>
            branch_name="bug-lp${bug_id}-${release}"
            if [[ -n "$description" ]]; then
                ppa_name="${release}-${project}-lp${bug_id}-${description}"
            else
                ppa_name="${release}-${project}-lp${bug_id}"
            fi
            ;;
        *)
            error "Invalid PPA type: $ppa_type (use 'merge', 'sru', or 'bug')"
            ;;
    esac
    
    info "Branch name: $branch_name"
    info "PPA name: $ppa_name"
    
    # Check if branch exists
    if git rev-parse --verify "$branch_name" >/dev/null 2>&1; then
        info "Branch exists, checking out..."
        git checkout "$branch_name"
    else
        info "Creating new branch..."
        git checkout -b "$branch_name"
        
        # Store PPA metadata in git config for this branch
        git config "branch.${branch_name}.ppaname" "$ppa_name"
        [[ -n "$description" ]] && git config "branch.${branch_name}.description" "$description"
    fi
    
    success "On branch: $branch_name"
    echo ""
    
    # Show PPA information
    parse_branch "$branch_name"
    cmd_ppa_status
}

# Command: ppa-status - Show PPA information from current branch
cmd_ppa_status() {
    check_packaging_dir
    parse_current_branch
    
    echo "PPA: $PPA_FULL_NAME"
    echo "  Release: $PPA_RELEASE"
    echo "  Project: $PPA_PROJECT"
    echo "  Type: $PPA_TYPE"
    echo "  Bug ID: LP#${PPA_BUGID}"
    [[ -n "$PPA_DESC" ]] && echo "  Description: $PPA_DESC"
    echo "  Branch: $(get_current_branch)"
    echo "  PPA Target: $(get_ppa_target)"
    echo ""
    
    # Show latest .changes if available
    if changes_file=$(ls -t ../*.changes 2>/dev/null | head -1); then
        echo "  Latest .changes: $(basename "$changes_file")"
    fi
}

# Command: dch-auto - Automatic changelog update from current branch
cmd_dch_auto() {
    check_packaging_dir
    parse_current_branch
    
    # Get current version
    local current_version=$(dpkg-parsechangelog -S Version 2>/dev/null)
    if [[ -z "$current_version" ]]; then
        error "Could not parse current version from debian/changelog"
    fi
    
    # Remove old suffix if present (strip ~<release><number>)
    local base_version="${current_version%~${PPA_RELEASE}*}"
    
    # Calculate new version with suffix
    local version_suffix=$(get_version_suffix "$current_version")
    local new_version="${base_version}${version_suffix}"
    local changelog_msg=$(get_changelog_message)
    
    info "Current version: $current_version"
    info "New version: $new_version"
    info "Release: $PPA_RELEASE"
    info "Message: $changelog_msg"
    
    # Update changelog
    dch -v "$new_version" -D "$PPA_RELEASE" "$changelog_msg"
    
    success "Changelog updated successfully"
}

# Command: sb-auto - Build with sbuild from current branch
cmd_sb_auto() {
    check_packaging_dir
    parse_current_branch
    
    info "Building for $PPA_RELEASE using sbuild"
    
    # Build source package first
    info "Building source package..."
    dpkg-buildpackage -S -us -uc
    
    # Find the .dsc file
    local dsc_file=$(ls -t ../*.dsc 2>/dev/null | head -1)
    if [[ -z "$dsc_file" ]]; then
        error "No .dsc file found"
    fi
    
    info "Found: $(basename "$dsc_file")"
    info "Building with sbuild -d $PPA_RELEASE..."
    
    sbuild -d "$PPA_RELEASE" "$dsc_file"
    
    success "Build completed"
}

# Command: dput-auto - Automatic upload to correct PPA from current branch
cmd_dput_auto() {
    check_packaging_dir
    parse_current_branch
    
    local ppa_target=$(get_ppa_target)
    local changes_file=$(find_latest_changes)
    
    info "Uploading to: $ppa_target"
    info "Changes file: $(basename "$changes_file")"
    
    # Confirm upload
    echo -n "Proceed with upload? [y/N] "
    read -r response
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        warn "Upload cancelled"
        exit 0
    fi
    
    dput "$ppa_target" "$changes_file"
    
    success "Upload completed"
}

# Command: ubuild - Complete build and upload workflow from current branch
cmd_ubuild() {
    check_packaging_dir
    parse_current_branch
    
    info "Starting complete build workflow"
    info "PPA: $PPA_FULL_NAME"
    info "Branch: $(get_current_branch)"
    echo ""
    
    # Step 1: Build
    cmd_sb_auto
    echo ""
    
    # Step 2: Upload
    cmd_dput_auto
    
    success "Build and upload workflow completed"
}

# Main command dispatcher
main() {
    local cmd="$1"
    shift
    
    case "$cmd" in
        gbranch)
            cmd_gbranch "$@"
            ;;
        ppa-status)
            cmd_ppa_status "$@"
            ;;
        dch-auto)
            cmd_dch_auto "$@"
            ;;
        sb-auto)
            cmd_sb_auto "$@"
            ;;
        dput-auto)
            cmd_dput_auto "$@"
            ;;
        ubuild)
            cmd_ubuild "$@"
            ;;
        *)
            error "Unknown command: $cmd
Available commands:
  gbranch <project> <bug-id> [type] [description]  - Create/checkout branch
  ppa-status                                        - Show PPA info from branch
  dch-auto                                          - Update changelog from branch
  sb-auto                                           - Build with sbuild from branch
  dput-auto                                         - Upload to PPA from branch
  ubuild                                            - Complete build+upload from branch"
            ;;
    esac
}

# Run if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
