# Shell Autocompletion for ToolBox

ToolBox provides intelligent shell autocompletion for all commands, flags, and contexts across multiple shells.

## Supported Shells

- ✅ **Bash** (3.2+)
- ✅ **Zsh** (5.0+)
- ✅ **Fish** (3.0+)
- ✅ **PowerShell** (5.0+)

---

## Quick Start

### Bash

**Linux:**
```bash
tb completion bash | sudo tee /etc/bash_completion.d/tb > /dev/null
source ~/.bashrc
```

**macOS (Homebrew):**
```bash
tb completion bash > $(brew --prefix)/etc/bash_completion.d/tb
source ~/.bashrc
```

**Manual:**
```bash
tb completion bash > ~/.tb-completion.bash
echo 'source ~/.tb-completion.bash' >> ~/.bashrc
source ~/.bashrc
```

### Zsh

**Standard:**
```bash
# Enable zsh completion if not already enabled
echo "autoload -U compinit; compinit" >> ~/.zshrc

# Install completion
tb completion zsh > "${fpath[1]}/_tb"
source ~/.zshrc
```

**Oh My Zsh:**
```bash
mkdir -p ~/.oh-my-zsh/custom/plugins/tb
tb completion zsh > ~/.oh-my-zsh/custom/plugins/tb/_tb
# Add 'tb' to plugins in ~/.zshrc
source ~/.zshrc
```

**Custom Directory:**
```bash
mkdir -p ~/.zsh/completion
tb completion zsh > ~/.zsh/completion/_tb
echo 'fpath=(~/.zsh/completion $fpath)' >> ~/.zshrc
source ~/.zshrc
```

### Fish

```bash
tb completion fish > ~/.config/fish/completions/tb.fish
```

Completions are loaded automatically in Fish!

### PowerShell

**Temporary (current session):**
```powershell
tb completion powershell | Out-String | Invoke-Expression
```

**Persistent:**
```powershell
# Add to profile
tb completion powershell >> $PROFILE

# Reload profile
. $PROFILE
```

---

## What Gets Completed

### 1. Commands (Context-Aware)

Autocompletion suggests commands based on your current project:

```bash
# In a Go project:
tb bu<TAB>  →  tb build
tb te<TAB>  →  tb test

# In a Docker project:
tb bu<TAB>  →  tb build
tb ru<TAB>  →  tb run
tb pu<TAB>  →  tb push

# In a Kubernetes project:
tb ap<TAB>  →  tb apply
tb ge<TAB>  →  tb get
```

### 2. Flags

All flags have completion:

```bash
tb --co<TAB>    →  tb --config
tb --ctx<TAB>   →  tb --context
tb --dr<TAB>    →  tb --dry-run
tb --ve<TAB>    →  tb --verbose
tb --ti<TAB>    →  tb --timeout
```

### 3. Context Names

The `--context` flag completes available contexts:

```bash
tb --context g<TAB>     →  tb --context go
tb --context no<TAB>    →  tb --context node
tb --context do<TAB>    →  tb --context docker
tb --context ku<TAB>    →  tb --context kubernetes
```

### 4. Config Files

The `--config` flag uses file completion:

```bash
tb --config <TAB>       →  Shows .yaml and .yml files
tb --config conf<TAB>   →  tb --config config.yaml
```

### 5. Subcommands

```bash
tb plu<TAB>      →  tb plugin
tb plugin l<TAB>  →  tb plugin list
tb plugin i<TAB>  →  tb plugin info
tb plugin c<TAB>  →  tb plugin contexts
```

---

## Installation Scripts

### One-Line Install

**Bash (Linux/macOS):**
```bash
curl -s https://raw.githubusercontent.com/bamf0/toolbox/main/scripts/install-completion.sh | bash
```

**Fish:**
```bash
tb completion fish > ~/.config/fish/completions/tb.fish
```

**PowerShell:**
```powershell
iwr -useb https://raw.githubusercontent.com/bamf0/toolbox/main/scripts/install-completion.ps1 | iex
```

---

## Examples

### Example 1: Building in Different Projects

```bash
# Navigate to Go project
cd ~/projects/my-go-app
tb bui<TAB>  # Suggests: build
# Command: go build ./...

# Navigate to Node project
cd ~/projects/my-node-app
tb bui<TAB>  # Suggests: build
# Command: npm run build

# Navigate to Docker project
cd ~/projects/my-docker-app
tb bui<TAB>  # Suggests: build
# Command: docker build -t my-app .
```

### Example 2: Kubernetes Workflows

```bash
cd ~/projects/k8s-manifests
tb <TAB>
# Suggests: apply, delete, describe, exec, get, logs, port-forward

tb ap<TAB>       # Completes to: apply
tb delete<TAB>   # Completes to: delete
tb get<TAB>      # Completes to: get
```

### Example 3: Flag Completion

```bash
tb build --<TAB>
# Suggests: --config, --context, --dry-run, --help, --timeout, --verbose

tb build --context <TAB>
# Suggests: docker, docker-compose, go, helm, kubernetes, node, python, rust

tb --config <TAB>
# Shows: .toolbox.yaml, config.yaml, etc.
```

---

## How It Works

### Context Detection

Completion is **context-aware**. When you press TAB:

1. **Detect Project Type** - Scans current directory for markers
   - Docker: Dockerfile, docker-compose.yml
   - Kubernetes: deployment.yaml, Chart.yaml
   - Go: go.mod
   - Node: package.json
   - etc.

2. **Load Available Commands** - Gets commands for detected context
   - From user config (`~/.toolbox/config.yaml`)
   - From plugins (Docker, Kubernetes, etc.)
   - From built-in defaults

3. **Filter by Prefix** - Shows only matching suggestions
   - `tb bu<TAB>` → filters to commands starting with "bu"

### Performance

Completion is **fast** (<10ms typical):
- Context detection is cached
- Config loaded once per session
- Minimal external calls

---

## Troubleshooting

### Completion Not Working

**1. Verify Installation**
```bash
# Check if completion script exists
ls -la ~/.tb-completion.bash  # Bash
ls -la ~/.config/fish/completions/tb.fish  # Fish
```

**2. Reload Shell**
```bash
source ~/.bashrc  # Bash
source ~/.zshrc   # Zsh
# Fish auto-loads
```

**3. Check Permissions**
```bash
chmod +x $(which tb)
```

**4. Verify tb in PATH**
```bash
which tb
# Should output: /usr/local/bin/tb (or similar)
```

### Completion Shows Wrong Commands

**Context not detected correctly:**
```bash
# Force specific context
tb --context docker build<TAB>
```

**Update configuration:**
```bash
# Check current context
tb --verbose build --dry-run
```

### Completion Slow

**Disable verbose mode in completion**
- Completion uses lightweight detection
- Should be <10ms

**Check file system performance:**
```bash
# Test detection speed
time tb completion bash > /dev/null
```

---

## Advanced Configuration

### Custom Completion Directory (Zsh)

```bash
# ~/.zshrc
fpath=(~/.zsh/completion $fpath)
autoload -Uz compinit && compinit
```

### Bash Completion with bash-completion Package

```bash
# Install bash-completion first
sudo apt-get install bash-completion  # Debian/Ubuntu
brew install bash-completion@2         # macOS

# Then install tb completion
tb completion bash > /etc/bash_completion.d/tb
```

### Fish Custom Completions

```bash
# Custom completion location
set -U fish_complete_path $fish_complete_path ~/.config/fish/completions
```

---

## Uninstallation

### Bash
```bash
sudo rm /etc/bash_completion.d/tb
# or
rm ~/.tb-completion.bash
# Remove source line from ~/.bashrc
```

### Zsh
```bash
rm "${fpath[1]}/_tb"
# or
rm ~/.zsh/completion/_tb
```

### Fish
```bash
rm ~/.config/fish/completions/tb.fish
```

### PowerShell
```powershell
# Remove completion line from $PROFILE
```

---

## Testing Completion

### Manual Testing

**Bash/Zsh/Fish:**
```bash
# Test command completion
tb bui<TAB>

# Test flag completion
tb --co<TAB>

# Test context completion
tb --context d<TAB>

# Test subcommand completion
tb plugin l<TAB>
```

**PowerShell:**
```powershell
# Test completion
tb bu<TAB>
```

### Verify Completion Functions

**Bash:**
```bash
complete -p tb
# Should show: complete -o default -F __start_tb tb
```

**Zsh:**
```bash
which _tb
# Should show: _tb is a shell function
```

**Fish:**
```bash
complete -c tb
# Should show completion definitions
```

---

## Best Practices

### 1. Keep tb in PATH
```bash
# Verify
which tb
echo $PATH
```

### 2. Update Completion on Upgrade
```bash
# After upgrading tb
tb completion bash > /etc/bash_completion.d/tb
source ~/.bashrc
```

### 3. Use Aliases Carefully
```bash
# If using alias, add completion
alias t='tb'
complete -F _tb t  # Bash
compdef t=tb      # Zsh
```

---

## FAQ

**Q: Does completion work offline?**  
A: Yes! Completion is entirely local.

**Q: Does it work with sudo?**  
A: Yes, with proper shell configuration.

**Q: Can I disable context detection?**  
A: Use `--context` flag to force a context.

**Q: Does it support custom plugins?**  
A: Yes! Custom plugins are auto-completed.

**Q: Performance impact?**  
A: Minimal (<10ms), completion is optimized.

---

## Support

For issues or questions:
1. Check this documentation
2. Run `tb completion --help`
3. Test with `tb --verbose build --dry-run`
4. File an issue on GitHub

---

## Summary

ToolBox provides **intelligent, context-aware autocompletion** for:

✅ Dynamic commands (based on project type)  
✅ All flags  
✅ Context names  
✅ Config files  
✅ Subcommands  
✅ Plugin commands  

**Installation**: One command per shell  
**Performance**: <10ms typical  
**Shells**: Bash, Zsh, Fish, PowerShell  

**Get started**:
```bash
tb completion bash > /etc/bash_completion.d/tb
source ~/.bashrc
```

Enjoy intelligent autocompletion! 🚀
