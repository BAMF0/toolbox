#!/bin/bash
# Install ToolBox man pages

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
info() {
    echo -e "${BLUE}Info: $1${NC}"
}

success() {
    echo -e "${GREEN}$1${NC}"
}

error() {
    echo -e "${RED}Error: $1${NC}" >&2
    exit 1
}

warn() {
    echo -e "${YELLOW}Warning: $1${NC}" >&2
}

# Parse arguments
USER_INSTALL=false
if [[ "$1" == "--user" ]]; then
    USER_INSTALL=true
fi

# Determine script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Check if man pages exist
if [[ ! -f "$SCRIPT_DIR/tb.1" ]]; then
    error "Man pages not found in $SCRIPT_DIR"
fi

# Installation
if $USER_INSTALL; then
    info "Installing man pages for current user..."
    
    # Create directories
    MAN1_DIR="$HOME/.local/share/man/man1"
    MAN5_DIR="$HOME/.local/share/man/man5"
    
    mkdir -p "$MAN1_DIR"
    mkdir -p "$MAN5_DIR"
    
    # Copy man pages
    cp "$SCRIPT_DIR"/*.1 "$MAN1_DIR/" 2>/dev/null || true
    cp "$SCRIPT_DIR"/*.5 "$MAN5_DIR/" 2>/dev/null || true
    
    success "Man pages installed to ~/.local/share/man/"
    
    # Check MANPATH
    if [[ ":$MANPATH:" != *":$HOME/.local/share/man:"* ]]; then
        warn "Add to your ~/.bashrc or ~/.zshrc:"
        echo ""
        echo "  export MANPATH=\"\$HOME/.local/share/man:\$MANPATH\""
        echo ""
    fi
else
    # System-wide installation
    if [[ $EUID -ne 0 ]]; then
        error "System-wide installation requires root. Run with sudo or use --user flag."
    fi
    
    info "Installing man pages system-wide..."
    
    # Determine system man directories
    if [[ -d "/usr/share/man" ]]; then
        MAN1_DIR="/usr/share/man/man1"
        MAN5_DIR="/usr/share/man/man5"
    elif [[ -d "/usr/local/share/man" ]]; then
        MAN1_DIR="/usr/local/share/man/man1"
        MAN5_DIR="/usr/local/share/man/man5"
    else
        error "Could not find system man directory"
    fi
    
    # Copy man pages
    cp "$SCRIPT_DIR"/*.1 "$MAN1_DIR/" 2>/dev/null || true
    cp "$SCRIPT_DIR"/*.5 "$MAN5_DIR/" 2>/dev/null || true
    
    success "Man pages installed to $MAN1_DIR/ and $MAN5_DIR/"
    
    # Update man database
    info "Updating man database..."
    if command -v mandb &> /dev/null; then
        mandb -q 2>/dev/null || warn "Could not update man database"
    elif command -v makewhatis &> /dev/null; then
        makewhatis 2>/dev/null || warn "Could not update man database"
    else
        warn "Could not find mandb or makewhatis to update man database"
    fi
    
    success "Installation complete"
fi

echo ""
info "You can now use: man tb, man tb-plugin, man tb-completion, man tb-help, man tb-config"
echo ""

# Test installation
if man -w tb &> /dev/null; then
    success "✓ Man pages are accessible"
    echo ""
    echo "Try: man tb"
else
    warn "Man pages may not be in your MANPATH yet. You may need to:"
    echo "  - Restart your shell"
    echo "  - Add ~/.local/share/man to MANPATH (for user install)"
fi
