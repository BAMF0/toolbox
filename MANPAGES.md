# ToolBox Man Pages

Comprehensive manual pages have been created for ToolBox.

## Available Man Pages

### Section 1: User Commands

| Man Page | Command | Description |
|----------|---------|-------------|
| **tb(1)** | `man tb` | Main ToolBox manual - overview, options, contexts |
| **tb-plugin(1)** | `man tb-plugin` | Plugin management commands |
| **tb-completion(1)** | `man tb-completion` | Shell completion setup for bash, zsh, fish, powershell |
| **tb-help(1)** | `man tb-help` | Getting help for commands |

### Section 5: File Formats

| Man Page | Command | Description |
|----------|---------|-------------|
| **tb-config(5)** | `man tb-config` | Configuration file format (.toolbox.yaml) |

## Quick Installation

### User Install (No sudo required)

```bash
make install-man-user
```

This installs man pages to `~/.local/share/man/` and is immediately usable.

### System-wide Install

```bash
sudo make install-man
```

This installs to `/usr/share/man/` or `/usr/local/share/man/`.

## Viewing Man Pages

After installation:

```bash
man tb              # Main manual
man tb-plugin       # Plugin commands
man tb-completion   # Completion setup
man tb-help         # Help system
man tb-config       # Config format
man 5 tb-config     # Explicitly request section 5
```

### Without Installation

View directly from the source:

```bash
man ./docs/man/tb.1
man ./docs/man/tb-plugin.1
man ./docs/man/tb-completion.1
man ./docs/man/tb-help.1
man ./docs/man/tb-config.5
```

## Features

The man pages include:

- **Complete command reference** - All options, flags, and subcommands
- **Context documentation** - All built-in and plugin contexts
- **Examples** - Real-world usage examples
- **Configuration guide** - Complete YAML config documentation
- **Installation instructions** - Shell completion for all supported shells
- **Cross-references** - Links between related man pages

## File Locations

Man page source files are located in:

```
docs/man/
├── README.md           # Detailed man pages documentation
├── install-man.sh      # Installation script
├── tb.1               # Main manual (section 1)
├── tb-plugin.1        # Plugin management (section 1)
├── tb-completion.1    # Completion setup (section 1)
├── tb-help.1          # Help command (section 1)
└── tb-config.5        # Config format (section 5)
```

## Converting to Other Formats

### PDF

```bash
groff -man -Tpdf docs/man/tb.1 > tb.pdf
```

### HTML

```bash
groff -man -Thtml docs/man/tb.1 > tb.html
```

### Plain Text

```bash
man docs/man/tb.1 | col -b > tb.txt
```

## Makefile Targets

The Makefile includes convenient targets for man page management:

```bash
make install-man        # System-wide install (requires sudo)
make install-man-user   # User install (no sudo)
make uninstall-man      # Remove system-wide man pages
make install-all        # Install binary + man pages
make uninstall-all      # Remove binary + man pages
```

## Standards Compliance

The man pages follow:

- **Unix man page conventions** - Standard sections and formatting
- **groff_man(7)** - Using standard man macros
- **man-pages(7)** - Linux man-pages project guidelines
- **POSIX** - Portable to all Unix-like systems

## Testing

Test man pages for warnings:

```bash
for f in docs/man/*.1 docs/man/*.5; do
    man --warnings -l "$f" > /dev/null
done
```

## See Also

- [docs/man/README.md](docs/man/README.md) - Detailed documentation
- [docs/user-guide.md](docs/user-guide.md) - User guide
- [docs/configuration.md](docs/configuration.md) - Configuration guide
- [docs/command-reference.md](docs/command-reference.md) - Command reference

## Contributing

When adding new features to ToolBox:

1. Update the relevant man page
2. Update the date in the `.TH` line
3. Test rendering: `man ./docs/man/<file>`
4. Verify no warnings: `man --warnings -l docs/man/<file>`

## License

Man pages are part of ToolBox and released under the same MIT License.
