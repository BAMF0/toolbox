# ToolBox Man Pages

This directory contains manual pages for ToolBox (tb).

## Available Man Pages

### Section 1: User Commands
- **tb(1)** - Main ToolBox command
- **tb-plugin(1)** - Plugin management
- **tb-completion(1)** - Shell completion generation
- **tb-help(1)** - Command help

### Section 5: File Formats
- **tb-config(5)** - Configuration file format

## Installing Man Pages

### System-wide Installation (Linux/macOS)

```bash
# Install all man pages
sudo cp docs/man/*.1 /usr/share/man/man1/
sudo cp docs/man/*.5 /usr/share/man/man5/

# Update man database
sudo mandb  # Linux
# or
sudo makewhatis  # macOS
```

### User Installation (No root required)

```bash
# Create user man directories
mkdir -p ~/.local/share/man/man1
mkdir -p ~/.local/share/man/man5

# Copy man pages
cp docs/man/*.1 ~/.local/share/man/man1/
cp docs/man/*.5 ~/.local/share/man/man5/

# Add to MANPATH (add to ~/.bashrc or ~/.zshrc)
export MANPATH="$HOME/.local/share/man:$MANPATH"
```

### Quick Installation Script

Use the provided installation script:

```bash
# System-wide (requires sudo)
sudo ./docs/man/install-man.sh

# User installation
./docs/man/install-man.sh --user
```

## Viewing Man Pages

After installation:

```bash
# Main man page
man tb

# Plugin management
man tb-plugin

# Shell completion
man tb-completion

# Help command
man tb-help

# Configuration format
man tb-config
man 5 tb-config  # Explicitly request section 5
```

## Viewing Without Installation

You can view man pages directly without installing:

```bash
# View with man
man docs/man/tb.1

# View with less
less docs/man/tb.1

# Convert to text
man docs/man/tb.1 | col -b > tb.txt

# Convert to PDF (requires groff)
groff -man -Tpdf docs/man/tb.1 > tb.pdf

# Convert to HTML (requires groff)
groff -man -Thtml docs/man/tb.1 > tb.html
```

## Format

Man pages are written in roff format using the `man` macro package. This is the standard format for Unix/Linux manual pages.

### Man Page Sections

The man pages follow standard Unix conventions:

- **Section 1**: User commands (executable programs or shell commands)
- **Section 5**: File formats and conventions (configuration files)

### Macro Reference

Common macros used in these pages:

- `.TH` - Title heading
- `.SH` - Section heading
- `.SS` - Subsection heading
- `.TP` - Tagged paragraph
- `.PP` - New paragraph
- `.B` - Bold text
- `.I` - Italic text
- `.BR` - Bold + Roman
- `.in` - Indent
- `.nf` - No fill (verbatim)
- `.fi` - Fill (end verbatim)

## Updating Man Pages

When updating ToolBox:

1. Update the corresponding man page
2. Update the date in the `.TH` line
3. Test the rendering: `man ./docs/man/tb.1`
4. Reinstall if needed

## Testing

Test man pages before distribution:

```bash
# Check for syntax errors
for f in docs/man/*.1 docs/man/*.5; do
    man --warnings -l "$f" > /dev/null
done

# View to check formatting
man ./docs/man/tb.1

# Check cross-references
lexgrog docs/man/*.1 docs/man/*.5
```

## Resources

- [man(7)](https://man7.org/linux/man-pages/man7/man.7.html) - Man page format
- [groff_man(7)](https://man7.org/linux/man-pages/man7/groff_man.7.html) - groff man macros
