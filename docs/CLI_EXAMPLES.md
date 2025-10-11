# CLI Examples

Comprehensive examples for all cursor-rules commands.

## Table of Contents

- [Installation](#installation)
- [Basic Usage](#basic-usage)
- [Install Command](#install-command)
- [Remove Command](#remove-command)
- [List Command](#list-command)
- [Effective Command](#effective-command)
- [Transform Command](#transform-command)
- [Sync Command](#sync-command)
- [Watch Command](#watch-command)
- [Init Command](#init-command)
- [Policy Command](#policy-command)
- [Advanced Workflows](#advanced-workflows)

---

## Installation

```bash
# Install from source
git clone https://github.com/ZanzyTHEbar/cursor-rules.git
cd cursor-rules
make install

# Verify installation
cursor-rules --version
```

---

## Basic Usage

```bash
# Get help
cursor-rules --help

# Get help for specific command
cursor-rules install --help

# Set working directory
cursor-rules --workdir /path/to/project install frontend

# Use environment variable for shared directory
export CURSOR_RULES_DIR=~/my-rules
cursor-rules list
```

---

## Install Command

Install presets or packages into your project.

### Basic Installation

```bash
# Install a single preset (default: Cursor format)
cursor-rules install frontend

# Install to specific directory
cursor-rules --workdir /path/to/project install frontend

# Install with verbose output
cursor-rules install frontend -v
```

### Target Formats

```bash
# Install to Cursor (.cursor/rules/)
cursor-rules install frontend --target cursor

# Install to GitHub Copilot Instructions (.github/instructions/)
cursor-rules install frontend --target copilot-instr

# Install to GitHub Copilot Prompts (.github/prompts/)
cursor-rules install frontend --target copilot-prompt

# Install to all targets defined in manifest
cursor-rules install frontend --all-targets
```

### Package Installation

```bash
# Install entire package
cursor-rules install frontend-package

# Install package with exclusions
cursor-rules install frontend-package --exclude "*.test.mdc"
cursor-rules install frontend-package --exclude "tests/*" --exclude "*.draft.mdc"

# Preserve directory structure (don't flatten)
cursor-rules install frontend-package --no-flatten

# Install nested package
cursor-rules install company/team/backend
```

### Advanced Installation

```bash
# Install multiple presets
cursor-rules install frontend
cursor-rules install backend
cursor-rules install testing

# Install with custom shared directory
CURSOR_RULES_DIR=~/custom-rules cursor-rules install frontend

# Install to multiple projects
for project in project1 project2 project3; do
  cursor-rules --workdir "$project" install frontend
done
```

---

## Remove Command

Remove installed presets from your project.

### Basic Removal

```bash
# Remove a preset
cursor-rules remove frontend

# Remove from specific directory
cursor-rules --workdir /path/to/project remove frontend

# Remove multiple presets
cursor-rules remove frontend
cursor-rules remove backend
cursor-rules remove testing
```

### Target-Specific Removal

```bash
# Remove from Cursor
cursor-rules remove frontend --target cursor

# Remove from Copilot Instructions
cursor-rules remove frontend --target copilot-instr

# Remove from Copilot Prompts
cursor-rules remove frontend --target copilot-prompt
```

---

## List Command

List available presets in shared directory.

### Basic Listing

```bash
# List all presets
cursor-rules list

# List with custom shared directory
CURSOR_RULES_DIR=~/custom-rules cursor-rules list

# List in specific format
cursor-rules list --format json
cursor-rules list --format yaml
```

### Filtering

```bash
# List only packages
cursor-rules list --packages-only

# List only single presets
cursor-rules list --presets-only

# List with pattern matching
cursor-rules list | grep frontend
cursor-rules list | grep -E "^(frontend|backend)"
```

---

## Effective Command

Show the effective rules for the current project.

### Basic Usage

```bash
# Show effective rules (default: Cursor format)
cursor-rules effective

# Show for specific directory
cursor-rules --workdir /path/to/project effective

# Show for specific target
cursor-rules effective --target cursor
cursor-rules effective --target copilot-instr
cursor-rules effective --target copilot-prompt
```

### Output Formats

```bash
# Output to file
cursor-rules effective > effective-rules.md

# Show with line numbers
cursor-rules effective | nl

# Show with syntax highlighting (if bat is installed)
cursor-rules effective | bat -l markdown

# Count rules
cursor-rules effective | grep -c "^#"
```

---

## Transform Command

Transform rules between different formats.

### Basic Transformation

```bash
# Transform to Copilot Instructions (default)
cursor-rules transform frontend

# Transform to specific target
cursor-rules transform frontend --target copilot-instr
cursor-rules transform frontend --target copilot-prompt
cursor-rules transform frontend --target cursor
```

### Batch Transformation

```bash
# Transform all presets in directory
for preset in ~/.cursor-rules/*.mdc; do
  cursor-rules transform "$(basename "$preset" .mdc)" --target copilot-instr
done

# Transform package
cursor-rules transform frontend-package --target copilot-instr
```

### Advanced Transformation

```bash
# Transform and save to custom location
cursor-rules transform frontend --target copilot-instr --workdir /custom/output

# Transform with validation
cursor-rules transform frontend --target copilot-instr --validate

# Transform and compare
cursor-rules transform frontend --target copilot-instr > new-format.md
diff .cursor/rules/frontend.mdc new-format.md
```

---

## Sync Command

Sync presets from remote repository.

### Basic Sync

```bash
# Sync from default remote
cursor-rules sync

# Sync with custom shared directory
CURSOR_RULES_DIR=~/custom-rules cursor-rules sync

# Sync and show changes
cursor-rules sync --verbose
```

### Git Integration

```bash
# Sync from specific branch
cd ~/.cursor-rules
git checkout develop
cursor-rules sync

# Sync and auto-apply to projects
cursor-rules sync --auto-apply

# Sync with conflict resolution
cursor-rules sync --strategy merge
cursor-rules sync --strategy rebase
```

---

## Watch Command

Watch for changes and auto-apply to projects.

### Basic Watching

```bash
# Start watcher
cursor-rules watch

# Watch with custom shared directory
CURSOR_RULES_DIR=~/custom-rules cursor-rules watch

# Watch with verbose output
cursor-rules watch --verbose
```

### Advanced Watching

```bash
# Watch and apply to specific projects
cursor-rules watch --projects /path/to/project1,/path/to/project2

# Watch with custom mapping file
cursor-rules watch --mapping ~/.cursor-rules/custom-mapping.yaml

# Watch in background
cursor-rules watch &
echo $! > ~/.cursor-rules-watch.pid

# Stop watcher
kill $(cat ~/.cursor-rules-watch.pid)
```

### Watcher Configuration

Create `~/.cursor-rules/watcher-mapping.yaml`:

```yaml
presets:
  frontend:
    - /home/user/project1
    - /home/user/project2
  backend:
    - /home/user/project1
  testing:
    - /home/user/project1
    - /home/user/project2
    - /home/user/project3
```

Then run:
```bash
cursor-rules watch
```

---

## Init Command

Initialize a new project with cursor-rules.

### Basic Initialization

```bash
# Initialize current directory
cursor-rules init

# Initialize specific directory
cursor-rules init /path/to/project

# Initialize with default presets
cursor-rules init --with-presets frontend,backend,testing
```

### Advanced Initialization

```bash
# Initialize with custom template
cursor-rules init --template company/standard

# Initialize and install all presets
cursor-rules init --install-all

# Initialize with specific target
cursor-rules init --target copilot-instr
```

---

## Policy Command

Manage policy rules for the project.

### Basic Policy Management

```bash
# Show current policy
cursor-rules policy show

# Set policy
cursor-rules policy set --enforce-frontmatter
cursor-rules policy set --require-description
cursor-rules policy set --max-file-size 100KB

# Validate against policy
cursor-rules policy validate
```

### Policy Configuration

Create `.cursor-rules-policy.yaml`:

```yaml
enforce:
  frontmatter: true
  description: true
  tags: true
  
limits:
  max_file_size: 102400  # 100KB
  max_rules_per_file: 50
  
allowed_tags:
  - frontend
  - backend
  - testing
  - security
  - performance
```

Then run:
```bash
cursor-rules policy validate
```

---

## Advanced Workflows

### Multi-Project Setup

```bash
#!/bin/bash
# setup-projects.sh

PROJECTS=(
  "/home/user/project1"
  "/home/user/project2"
  "/home/user/project3"
)

PRESETS=(
  "frontend"
  "backend"
  "testing"
)

for project in "${PROJECTS[@]}"; do
  echo "Setting up $project..."
  for preset in "${PRESETS[@]}"; do
    cursor-rules --workdir "$project" install "$preset"
  done
done

echo "✅ All projects configured!"
```

### Automated Sync and Apply

```bash
#!/bin/bash
# sync-and-apply.sh

# Sync presets
cursor-rules sync

# Apply to all projects
PROJECTS=$(find ~/projects -name ".cursor" -type d -exec dirname {} \;)

for project in $PROJECTS; do
  echo "Updating $project..."
  cd "$project"
  
  # Get installed presets
  INSTALLED=$(cursor-rules list --installed)
  
  # Reinstall each preset
  for preset in $INSTALLED; do
    cursor-rules install "$preset"
  done
done

echo "✅ All projects updated!"
```

### CI/CD Integration

```yaml
# .github/workflows/cursor-rules.yml
name: Update Cursor Rules

on:
  schedule:
    - cron: '0 0 * * *'  # Daily at midnight
  workflow_dispatch:

jobs:
  update-rules:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.25'
      
      - name: Install cursor-rules
        run: |
          go install github.com/ZanzyTHEbar/cursor-rules@latest
      
      - name: Sync rules
        run: |
          cursor-rules sync
      
      - name: Update project rules
        run: |
          cursor-rules install frontend --all-targets
          cursor-rules install backend --all-targets
      
      - name: Commit changes
        run: |
          git config user.name "GitHub Actions"
          git config user.email "actions@github.com"
          git add .cursor/ .github/
          git commit -m "chore: update cursor rules" || true
          git push
```

### Custom Preset Development

```bash
# Create new preset
mkdir -p ~/.cursor-rules/my-preset
cat > ~/.cursor-rules/my-preset.mdc <<'EOF'
---
description: "My custom preset"
tags:
  - custom
  - development
---

# My Custom Preset

Custom rules for my workflow.

## Rules

1. Always use TypeScript
2. Follow ESLint configuration
3. Write tests for all features
EOF

# Test preset
cursor-rules install my-preset --target cursor

# Verify installation
cursor-rules effective | grep "My Custom Preset"

# Share preset
cd ~/.cursor-rules
git add my-preset.mdc
git commit -m "feat: add my-preset"
git push
```

### Preset Validation Script

```bash
#!/bin/bash
# validate-presets.sh

SHARED_DIR="${CURSOR_RULES_DIR:-$HOME/.cursor-rules}"

echo "Validating presets in $SHARED_DIR..."

for preset in "$SHARED_DIR"/*.mdc; do
  echo "Checking $(basename "$preset")..."
  
  # Check frontmatter
  if ! head -1 "$preset" | grep -q "^---$"; then
    echo "  ❌ Missing frontmatter"
    continue
  fi
  
  # Check description
  if ! grep -q "^description:" "$preset"; then
    echo "  ⚠️  Missing description"
  fi
  
  # Check tags
  if ! grep -q "^tags:" "$preset"; then
    echo "  ⚠️  Missing tags"
  fi
  
  echo "  ✅ Valid"
done

echo "Validation complete!"
```

### Backup and Restore

```bash
# Backup current rules
backup_rules() {
  local backup_dir="$HOME/.cursor-rules-backup-$(date +%Y%m%d-%H%M%S)"
  mkdir -p "$backup_dir"
  
  # Backup shared rules
  cp -r "$CURSOR_RULES_DIR" "$backup_dir/shared"
  
  # Backup project rules
  find ~/projects -name ".cursor" -type d | while read cursor_dir; do
    project_name=$(basename "$(dirname "$cursor_dir")")
    cp -r "$cursor_dir" "$backup_dir/projects/$project_name"
  done
  
  echo "✅ Backup created: $backup_dir"
}

# Restore from backup
restore_rules() {
  local backup_dir="$1"
  
  if [ -z "$backup_dir" ] || [ ! -d "$backup_dir" ]; then
    echo "❌ Invalid backup directory"
    return 1
  fi
  
  # Restore shared rules
  cp -r "$backup_dir/shared/"* "$CURSOR_RULES_DIR/"
  
  # Restore project rules
  find "$backup_dir/projects" -type d -mindepth 1 -maxdepth 1 | while read project_backup; do
    project_name=$(basename "$project_backup")
    project_path="$HOME/projects/$project_name"
    
    if [ -d "$project_path" ]; then
      cp -r "$project_backup/.cursor" "$project_path/"
    fi
  done
  
  echo "✅ Restore complete"
}

# Usage
backup_rules
# restore_rules ~/.cursor-rules-backup-20250111-120000
```

---

## Tips and Tricks

### Quick Aliases

Add to your `~/.bashrc` or `~/.zshrc`:

```bash
# Cursor rules aliases
alias cr='cursor-rules'
alias cri='cursor-rules install'
alias crr='cursor-rules remove'
alias crl='cursor-rules list'
alias cre='cursor-rules effective'
alias crs='cursor-rules sync'
alias crw='cursor-rules watch'

# Quick install common presets
alias cr-frontend='cursor-rules install frontend --all-targets'
alias cr-backend='cursor-rules install backend --all-targets'
alias cr-full='cursor-rules install frontend backend testing --all-targets'
```

### Shell Functions

```bash
# Install preset to current project
cri() {
  cursor-rules install "$@"
}

# Show effective rules with pager
cre() {
  cursor-rules effective | less
}

# Sync and show changes
crs() {
  cd "$CURSOR_RULES_DIR"
  git pull
  cursor-rules sync
}

# Watch in background
crw() {
  cursor-rules watch > ~/.cursor-rules-watch.log 2>&1 &
  echo $! > ~/.cursor-rules-watch.pid
  echo "Watcher started (PID: $(cat ~/.cursor-rules-watch.pid))"
}

# Stop watcher
crw-stop() {
  if [ -f ~/.cursor-rules-watch.pid ]; then
    kill $(cat ~/.cursor-rules-watch.pid)
    rm ~/.cursor-rules-watch.pid
    echo "Watcher stopped"
  else
    echo "No watcher running"
  fi
}
```

### Environment Setup

```bash
# ~/.cursor-rules-env

# Shared directory
export CURSOR_RULES_DIR="$HOME/.cursor-rules"

# Enable symlinks (optional)
export CURSOR_RULES_SYMLINK=1

# Enable GNU Stow mode (optional)
export CURSOR_RULES_USE_STOW=1

# Default target
export CURSOR_RULES_DEFAULT_TARGET="cursor"

# Load in shell
source ~/.cursor-rules-env
```

---

## Troubleshooting

### Common Issues

```bash
# Preset not found
cursor-rules list  # Check available presets
echo $CURSOR_RULES_DIR  # Verify shared directory

# Permission denied
sudo chown -R $USER:$USER ~/.cursor-rules
chmod -R 755 ~/.cursor-rules

# Git sync fails
cd ~/.cursor-rules
git status
git pull --rebase

# Watcher not working
pkill -f cursor-rules  # Kill existing watchers
cursor-rules watch --verbose  # Start with verbose output
```

### Debug Mode

```bash
# Enable debug logging
export LOG_LEVEL=debug
cursor-rules install frontend

# Trace execution
set -x
cursor-rules install frontend
set +x

# Check file operations
strace -e open,openat cursor-rules install frontend 2>&1 | grep -E "\.mdc|\.cursor"
```

---

## See Also

- [User Guide](USER_GUIDE.md) - Comprehensive user documentation
- [Troubleshooting Guide](TROUBLESHOOTING.md) - Common issues and solutions
- [Architecture](ARCHITECTURE.md) - System design and components
- [Contributing](../CONTRIBUTING.md) - How to contribute

---

**Last Updated:** 2025-01-11  
**Version:** 1.0
