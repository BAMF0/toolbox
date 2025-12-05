# Shell Autocompletion Guide

ToolBox provides intelligent shell autocompletion that suggests commands based on your current project context.

## Features

- **Context-Aware Suggestions**: Autocompletion suggests commands available in your current project type
- **Command Descriptions**: See what each command does while typing
- **Flag Completion**: Tab-complete `--context` and `--config` flags
- **Multi-Shell Support**: Works with Bash, Zsh, Fish, and PowerShell

## Installation

### Bash

#### Linux
```bash
# Generate and install completion script
tb completion bash | sudo tee /etc/bash_completion.d/tb

# Reload your shell
source ~/.bashrc
```

#### macOS
```bash
# Install bash-completion if not already installed
brew install bash-completion

# Generate and install completion script
tb completion bash > $(brew --prefix)/etc/bash_completion.d/tb

# Add to your ~/.bash_profile if not already there
echo "[ -f $(brew --prefix)/etc/bash_completion ] && . $(brew --prefix)/etc/bash_completion" >> ~/.bash_profile

# Reload your shell
source ~/.bash_profile
```

### Zsh

```bash
# Enable Zsh completion system if not already enabled
# Add this to your ~/.zshrc if needed
echo "autoload -U compinit; compinit" >> ~/.zshrc

# Option 1: Use first directory in fpath
tb completion zsh > "${fpath[1]}/_tb"

# Option 2: Use custom completion directory (recommended)
mkdir -p ~/.zsh/completion
tb completion zsh > ~/.zsh/completion/_tb
echo 'fpath=(~/.zsh/completion $fpath)' >> ~/.zshrc

# Reload your shell
source ~/.zshrc
```

### Fish

```bash
# Generate and install completion script
tb completion fish > ~/.config/fish/completions/tb.fish

# Fish will automatically load completions
# No need to reload
```

### PowerShell

```powershell
# One-time setup for current session
tb completion powershell | Out-String | Invoke-Expression

# Persistent setup - add to PowerShell profile
tb completion powershell >> $PROFILE

# Reload profile
. $PROFILE
```

## How It Works

### Context Detection

When you press Tab, ToolBox:

1. **Detects your project type** by looking for marker files:
   - `package.json` → Node.js context
   - `go.mod` → Go context
   - `debian/control` → Ubuntu packaging context
   - And more...

2. **Loads available commands** for that context

3. **Suggests matching commands** with descriptions

### Example Usage

```bash
# In a Node.js project directory
$ cd ~/my-node-app
$ tb <TAB>
build    Build the project (npm run build)
test     Run tests (npm test)
start    Start the application (npm start)
dev      Start development server (npm run dev)
lint     Run linter (npm run lint)
install  Install dependencies (npm install)

# In a Go project directory  
$ cd ~/my-go-api
$ tb <TAB>
build    Build all packages (go build ./...)
test     Run all tests (go test ./...)
run      Run main program (go run ./cmd/...)
fmt      Format code (go fmt ./...)
lint     Run golangci-lint (golangci-lint run)
```

### Flag Completion

```bash
# Complete --context flag with available contexts
$ tb build --context <TAB>
node  go  python  rust  make  ubuntu-packaging

# Complete --config flag with file paths
$ tb build --config <TAB>
.toolbox.yaml  config.yaml  my-config.yml
```

## Customizing Completion

### Adding Descriptions to Custom Commands

In your `.toolbox.yaml`, add descriptions to make completion more helpful:

```yaml
contexts:
  node:
    commands:
      deploy: "npm run deploy"
      e2e: "npm run test:e2e"
    descriptions:
      deploy: "Deploy to production server"
      e2e: "Run end-to-end tests with Playwright"
```

Now when you tab-complete:

```bash
$ tb <TAB>
deploy   Deploy to production server
e2e      Run end-to-end tests with Playwright
```

### Plugin Commands

Commands from plugins also appear in autocompletion:

```bash
# In a Debian packaging directory
$ tb <TAB>
gbranch      Create/checkout git branch for bug fixes
ppa-status   Show PPA information from current branch
dch-auto     Auto-update changelog with version suffix
ubuild       Complete build and upload workflow
```

## Troubleshooting

### Completion Not Working

1. **Verify installation**:
   ```bash
   # Bash
   ls -l /etc/bash_completion.d/tb
   
   # Zsh
   ls -l ~/.zsh/completion/_tb
   
   # Fish
   ls -l ~/.config/fish/completions/tb.fish
   ```

2. **Check permissions**:
   ```bash
   # Completion files should be readable
   chmod +r /etc/bash_completion.d/tb
   ```

3. **Reload shell**:
   ```bash
   # Bash
   source ~/.bashrc
   
   # Zsh
   source ~/.zshrc
   
   # Fish - restart terminal or run:
   source ~/.config/fish/config.fish
   ```

### No Context-Specific Suggestions

If you only see common commands (build, test, run) instead of context-specific ones:

1. **Verify project detection**:
   ```bash
   tb build --dry-run
   # Should show: "Context: <detected-context>"
   ```

2. **Check for marker files**:
   ```bash
   # Node.js
   ls package.json
   
   # Go
   ls go.mod
   
   # Python
   ls pyproject.toml requirements.txt
   ```

3. **Force context manually**:
   ```bash
   tb --context node build
   ```

### Slow Completion

If completion feels slow:

1. **Reduce config file size**: Large configs take longer to parse
2. **Simplify detection logic**: Avoid deeply nested directories
3. **Use local cache**: Put `.toolbox.yaml` in project root for faster access

### Zsh Completion Issues

If Zsh shows "command not found: compdef":

```bash
# Add these lines to the top of your ~/.zshrc
autoload -Uz compinit
compinit
```

### Bash Completion Not Loading

On some systems, you may need to explicitly source bash-completion:

```bash
# Add to ~/.bashrc
if [ -f /etc/bash_completion ]; then
    . /etc/bash_completion
fi
```

## Advanced Usage

### Dynamic Context Switching

Completion adapts when you override the context:

```bash
# Get Python-specific completions even in a Node.js project
$ tb --context python <TAB>
test     Run tests with pytest
lint     Check code with ruff
fmt      Format code with black
install  Install dependencies from requirements.txt
```

### Multi-Project Workflows

Working on multiple projects? Completion switches automatically:

```bash
$ cd ~/frontend-app
$ tb <TAB>
# Shows Node.js commands

$ cd ~/backend-api  
$ tb <TAB>
# Shows Go commands

$ cd ~/ubuntu-packages
$ tb <TAB>
# Shows Debian packaging commands
```

### Completion in Scripts

While primarily designed for interactive use, you can generate completion lists programmatically:

```bash
# Get list of available commands for current context
tb __complete "" | head -n -1
```

## Best Practices

1. **Add descriptions**: Makes completion more user-friendly
2. **Use consistent naming**: Keep command names intuitive (build, test, deploy)
3. **Keep contexts focused**: Don't overload a single context with too many commands
4. **Document custom commands**: Add descriptions for any non-standard commands

## Next Steps

- Learn about [Plugin Development](plugin-development.md) to create custom contexts
- See [Configuration Guide](configuration.md) for advanced config options
- Check out [Example Configurations](../examples/example-config.yaml)
