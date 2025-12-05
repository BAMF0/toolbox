#!/bin/bash
# ToolBox Autocompletion Installation Script
# Supports: Bash, Zsh, Fish

set -e

echo "========================================"
echo "ToolBox Autocompletion Installer"
echo "========================================"
echo ""

# Detect shell
SHELL_NAME=$(basename "$SHELL")

# Check if tb is installed
if ! command -v tb &> /dev/null; then
    echo "Error: 'tb' command not found in PATH"
    echo "Please install ToolBox first: go install github.com/bamf0/toolbox/cmd/tb@latest"
    exit 1
fi

echo "Detected shell: $SHELL_NAME"
echo ""

# Install based on shell type
case "$SHELL_NAME" in
    bash)
        echo "Installing Bash completion..."
        
        # Try system-wide installation first (requires sudo)
        if [ -w /etc/bash_completion.d ]; then
            tb completion bash > /etc/bash_completion.d/tb
            echo "✓ Installed to: /etc/bash_completion.d/tb"
        elif command -v brew &> /dev/null; then
            # macOS with Homebrew
            BREW_PREFIX=$(brew --prefix)
            mkdir -p "$BREW_PREFIX/etc/bash_completion.d"
            tb completion bash > "$BREW_PREFIX/etc/bash_completion.d/tb"
            echo "✓ Installed to: $BREW_PREFIX/etc/bash_completion.d/tb"
        else
            # User-local installation
            tb completion bash > ~/.tb-completion.bash
            
            if ! grep -q '.tb-completion.bash' ~/.bashrc; then
                echo 'source ~/.tb-completion.bash' >> ~/.bashrc
            fi
            
            echo "✓ Installed to: ~/.tb-completion.bash"
            echo "✓ Added source command to ~/.bashrc"
        fi
        
        echo ""
        echo "Reload your shell to activate:"
        echo "  source ~/.bashrc"
        ;;
        
    zsh)
        echo "Installing Zsh completion..."
        
        # Get first fpath directory
        if [ -n "$fpath" ]; then
            # Get writable fpath directory
            COMPLETION_DIR=""
            for dir in $fpath; do
                if [ -w "$dir" ]; then
                    COMPLETION_DIR="$dir"
                    break
                fi
            done
            
            if [ -n "$COMPLETION_DIR" ]; then
                tb completion zsh > "$COMPLETION_DIR/_tb"
                echo "✓ Installed to: $COMPLETION_DIR/_tb"
            else
                # Use custom directory
                mkdir -p ~/.zsh/completion
                tb completion zsh > ~/.zsh/completion/_tb
                
                if ! grep -q 'fpath=(~/.zsh/completion $fpath)' ~/.zshrc; then
                    echo 'fpath=(~/.zsh/completion $fpath)' >> ~/.zshrc
                fi
                
                if ! grep -q 'autoload -U compinit' ~/.zshrc; then
                    echo 'autoload -U compinit; compinit' >> ~/.zshrc
                fi
                
                echo "✓ Installed to: ~/.zsh/completion/_tb"
                echo "✓ Updated ~/.zshrc"
            fi
        else
            # Fallback
            mkdir -p ~/.zsh/completion
            tb completion zsh > ~/.zsh/completion/_tb
            
            if ! grep -q 'fpath=(~/.zsh/completion $fpath)' ~/.zshrc; then
                echo 'fpath=(~/.zsh/completion $fpath)' >> ~/.zshrc
                echo 'autoload -U compinit; compinit' >> ~/.zshrc
            fi
            
            echo "✓ Installed to: ~/.zsh/completion/_tb"
        fi
        
        echo ""
        echo "Reload your shell to activate:"
        echo "  source ~/.zshrc"
        ;;
        
    fish)
        echo "Installing Fish completion..."
        
        mkdir -p ~/.config/fish/completions
        tb completion fish > ~/.config/fish/completions/tb.fish
        
        echo "✓ Installed to: ~/.config/fish/completions/tb.fish"
        echo ""
        echo "Fish will automatically load the completion!"
        ;;
        
    *)
        echo "Unsupported shell: $SHELL_NAME"
        echo ""
        echo "Supported shells:"
        echo "  - bash"
        echo "  - zsh"
        echo "  - fish"
        echo ""
        echo "Manual installation:"
        echo "  tb completion --help"
        exit 1
        ;;
esac

echo ""
echo "========================================"
echo "Installation Complete!"
echo "========================================"
echo ""
echo "Test completion by typing:"
echo "  tb bui<TAB>"
echo ""
