# Changelog

All notable changes to ToolBox will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2025-12-05

### Added
- Initial release of ToolBox (tb)
- Context-aware command aliasing system
- Support for multiple project types:
  - Node.js (package.json)
  - Go (go.mod)
  - Python (pyproject.toml, requirements.txt)
  - Rust (Cargo.toml)
  - Make (Makefile)
- Plugin system for extensibility:
  - Docker plugin (Dockerfile, docker-compose.yml)
  - Kubernetes plugin (k8s YAML manifests)
  - Ubuntu packaging plugin (debian/control, debian/changelog)
- YAML-based configuration system
  - Project-level config (.toolbox.yaml)
  - User-level config (~/.toolbox/config.yaml)
  - Customizable contexts and commands
- Command-line features:
  - `--version` flag and `version` subcommand
  - `--dry-run` flag for previewing commands
  - `--verbose` flag for debugging
  - `--context` flag to force specific context
  - `--timeout` flag for command execution limits
- Shell completion support:
  - Bash
  - Zsh
  - Fish
  - PowerShell
- Comprehensive documentation:
  - Man pages (tb.1, tb-plugin.1, tb-completion.1, tb-help.1, tb-config.5)
  - User guide
  - Configuration guide
  - Plugin development guide
  - API reference
  - Command reference
- Build system:
  - Makefile with multiple targets
  - Version embedding via ldflags
  - Git commit and build time tracking
  - Optimized release builds
- Security features:
  - Command validation
  - Argument length limits
  - No shell injection vulnerabilities
  - Safe command execution

### Plugin Features
- **Ubuntu Packaging Plugin**:
  - Smart PPA-based workflow
  - Git branch management (gbranch)
  - Automatic changelog updates (dch-auto)
  - Build automation (sb-auto, ubuild)
  - Upload automation (dput-auto)
  - Branch-aware PPA detection

- **Docker Plugin**:
  - Container lifecycle management
  - Docker Compose support
  - Interactive shell access

- **Kubernetes Plugin**:
  - Resource management
  - Pod inspection and logs
  - kubectl integration

### Developer Experience
- Built-in help system (`tb help <command>`)
- Context detection with detailed output
- Command-specific descriptions
- Error messages with context
- Timeout protection for long-running commands

[0.1.0]: https://github.com/bamf0/toolbox/releases/tag/v0.1.0
