# Cursor Rules Manager - User Guide

**Version:** 1.0  
**Last Updated:** 2025-01-11

---

## Table of Contents

1. [Introduction](#introduction)
2. [Installation](#installation)
3. [Quick Start (5 Minutes)](#quick-start-5-minutes)
4. [Core Concepts](#core-concepts)
5. [Common Workflows](#common-workflows)
6. [Command Reference](#command-reference)
7. [Configuration](#configuration)
8. [Troubleshooting](#troubleshooting)
9. [Advanced Usage](#advanced-usage)
10. [FAQ](#faq)

---

## Introduction

Cursor Rules Manager is a CLI tool and VSCode extension that helps you manage shared Cursor `.mdc` rule presets across projects. It also supports transforming rules to GitHub Copilot instructions and prompts for seamless multi-tool workflows.

### Why Use Cursor Rules Manager?

- **Share rules** across projects and teams
- **Version control** your rules with Git
- **Transform** between Cursor and Copilot formats
- **Organize** rules into packages
- **Auto-apply** rules with file watching
- **Override** rules per project

### What You'll Learn

This guide will teach you how to:
- Install and configure the tool
- Manage shared rule presets
- Install rules into projects
- Transform between formats
- Use advanced features

---

## Installation

### Prerequisites

- **Go 1.25.2+** (for building from source)
- **Git** (for syncing shared presets)
- **VS Code or Cursor** (for the extension)

### Option 1: Install from Source

```bash
# Clone the repository
git clone https://github.com/ZanzyTHEbar/cursor-rules.git
cd cursor-rules

# Build and install
make install

# Verify installation
cursor-rules --version
```

### Option 2: Install Pre-built Binary

```bash
# Download latest release
# Visit: https://github.com/ZanzyTHEbar/cursor-rules/releases

# Extract and move to PATH
tar -xzf cursor-rules_*.tar.gz
sudo mv cursor-rules /usr/local/bin/

# Verify
cursor-rules --version
```

### Option 3: Install VSCode Extension

1. Download `.vsix` file from releases
2. Open VS Code/Cursor
3. Press `Cmd/Ctrl+Shift+P`
4. Type "Install from VSIX"
5. Select the downloaded file

---

## Quick Start (5 Minutes)

Let's get you up and running in 5 minutes!

### Step 1: Set Up Shared Directory (1 min)

```bash
# Create a directory for package presets
mkdir -p ~/cursor-rules-shared

# Set environment variable (add to ~/.bashrc or ~/.zshrc)
export CURSOR_RULES_PACKAGE_DIR=~/cursor-rules-shared
```

### Step 2: Create Your First Preset (2 min)

```bash
# Create a frontend preset
cat > ~/cursor-rules-shared/frontend.mdc <<'EOF'
---
description: "Frontend development rules"
apply_to: "**/*.{ts,tsx,js,jsx}"
priority: 1
---
# Frontend Best Practices

- Use TypeScript for type safety
- Follow React best practices
- Use functional components with hooks
- Write unit tests for components
EOF
```

### Step 3: Install to Your Project (1 min)

```bash
# Navigate to your project
cd ~/my-project

# Install the preset
cursor-rules install frontend

# Verify installation
ls .cursor/rules/
# Should show: frontend.mdc
```

### Step 4: View Effective Rules (1 min)

```bash
# See all active rules
cursor-rules effective

# Output shows merged rules from all sources
```

**🎉 Congratulations!** You've successfully installed your first preset!

---

## Core Concepts

### 1. Presets

**Presets** are individual rule files (`.mdc`) stored in your package directory.

```
~/cursor-rules-shared/
├── frontend.mdc
├── backend.mdc
├── testing.mdc
└── security.mdc
```

**Structure:**
```markdown
---
description: "Rule description"
apply_to: "**/*.ts"
priority: 1
---
Rule content here
```

### 2. Packages

**Packages** are directories containing multiple related presets.

```
~/cursor-rules-shared/
└── react-package/
    ├── components.mdc
    ├── hooks.mdc
    └── testing.mdc
```

**Install entire package:**
```bash
cursor-rules install react-package
```

### 3. Targets

**Targets** are output formats or context types:

| Target | Output Location | Format |
|--------|----------------|--------|
| `cursor` | `.cursor/rules/` | Cursor `.mdc` |
| `copilot-instr` | `.github/instructions/` | Copilot Instructions |
| `copilot-prompt` | `.github/prompts/` | Copilot Prompts |
| `cursor-commands` | `.cursor/commands/` | Command `.md` files |
| `cursor-skills` | `.cursor/skills/<name>/` | Skill dir with `SKILL.md` |
| `cursor-agents` | `.cursor/agents/<name>.md` | Agent markdown |
| `cursor-hooks` | `.cursor/hooks.json` + `.cursor/hooks/` | Hooks config and scripts |

### 4. Rules, commands, skills, agents, and hooks

The **package directory** (default `~/.cursor/rules`) can hold all five Cursor context types:

- **Rules:** `*.mdc` presets and `<package>/*.mdc` — install to `.cursor/rules/` or Copilot targets.
- **Commands:** `*.md` or `commands/<name>/` — install with `--target cursor-commands` to `.cursor/commands/`.
- **Skills:** `skills/<name>/SKILL.md` (plus optional `scripts/`, `references/`, `assets/`) — install with `--target cursor-skills` to `.cursor/skills/<name>/`.
- **Agents:** `agents/<name>.md` (YAML frontmatter + body) — install with `--target cursor-agents` to `.cursor/agents/<name>.md`.
- **Hooks:** `hooks/<preset>/hooks.json` plus script files — install with `--target cursor-hooks` to `.cursor/hooks.json` and `.cursor/hooks/`. Installing a hook preset **replaces** the project’s existing `.cursor/hooks.json`; script paths are rewritten to `.cursor/hooks/<script>`.

Use **init** to create all five project dirs (`.cursor/rules`, `commands`, `skills`, `agents`, `hooks`). Use **list** and **sync** to see presets, commands, skills, agents, and hook presets. Use **remove** with `--type rule|command|skill|agent|hooks` to remove by type (e.g. `remove my-skill --type skill`, `remove --type hooks`).

### 5. Manifest

**Manifest** (`cursor-rules-manifest.yaml`) defines package structure and targets.

```yaml
version: "1.0"
targets:
  - cursor
  - copilot-instr
  - copilot-prompt
```

---

## Common Workflows

### Workflow 1: Bootstrap a New Project

```bash
# 1. Navigate to project
cd ~/new-project

# 2. Sync shared presets
cursor-rules sync

# 3. Install presets
cursor-rules install frontend
cursor-rules install backend
cursor-rules install testing

# 4. Verify
cursor-rules effective
```

### Workflow 2: Share Rules with Team

```bash
# 1. Create Git repository for shared rules
cd ~/cursor-rules-shared
git init
git add .
git commit -m "Initial rules"
git remote add origin https://github.com/team/cursor-rules.git
git push -u origin main

# 2. Team members clone
git clone https://github.com/team/cursor-rules.git ~/cursor-rules-shared

# 3. Set environment variable
export CURSOR_RULES_PACKAGE_DIR=~/cursor-rules-shared
```

### Workflow 3: Use with GitHub Copilot

```bash
# Install to both Cursor and Copilot
cursor-rules install frontend --target cursor
cursor-rules install frontend --target copilot-instr

# Or install to all targets at once
cursor-rules install frontend --all-targets
```

### Workflow 4: Transform Existing Rules

```bash
# Convert Cursor rules to Copilot instructions
cursor-rules transform --from cursor --to copilot-instr

# Convert back
cursor-rules transform --from copilot-instr --to cursor
```

### Workflow 5: Auto-Apply with Watcher

```bash
# Start watcher (runs in background)
cursor-rules watch &

# Edit shared preset
vim ~/cursor-rules-shared/frontend.mdc

# Changes automatically applied to projects
```

---

## Command Reference

### `cursor-rules install`

Install a preset or package to your project.

**Usage:**
```bash
cursor-rules install <preset|package> [flags]
```

**Flags:**
- `--target <target>` - Output target: `cursor`, `copilot-instr`, `copilot-prompt`, `cursor-commands`, `cursor-skills`, `cursor-agents`, `cursor-hooks`
- `--all-targets` - Install to all targets in manifest
- `--global` (persistent) - Install to user dirs (~/.cursor/...) instead of project. Can be passed before the subcommand.
- `--dir <path|user>` (persistent) - Destination: path or `user` (same as `--global` when `user`).
- `--exclude <pattern>` - Exclude files matching pattern
- `-n, --no-flatten` - Preserve package directory structure
- `--workdir <dir>`, `-w` (persistent) - Project directory (default: current)

**Examples:**
```bash
# Install single preset
cursor-rules install frontend

# Install to user dirs (~/.cursor/...) instead of project
cursor-rules --global install frontend

# Install to Copilot instructions
cursor-rules install frontend --target copilot-instr

# Install Cursor skills, agents, or hook presets
cursor-rules install my-skill --target cursor-skills
cursor-rules install my-agent --target cursor-agents
cursor-rules install my-hooks --target cursor-hooks

# Install package with exclusions
cursor-rules install mypkg --exclude "*.test.mdc"

# Install to all targets
cursor-rules install frontend --all-targets
```

---

### `cursor-rules list`

List available presets and packages.

**Usage:**
```bash
cursor-rules list
```

**Flags:**
- `--global` (persistent) - list from user dirs (~/.cursor/...) instead of package directory. Can be passed before the subcommand: `cursor-rules --global list`.

**Output:** Lists presets, packages, commands, skills, agents, and hook presets from the package directory or user dirs when `--global` (each section shown when non-empty).

---

### `cursor-rules sync`

Sync package presets from Git repository.

**Usage:**
```bash
cursor-rules sync [flags]
```

**Flags:**
- `--apply` - Apply changes to projects after sync

**Examples:**
```bash
# Sync presets
cursor-rules sync

# Sync and apply
cursor-rules sync --apply
```

---

### `cursor-rules remove`

Remove a preset, command, skill, agent, or hooks from your project.

**Usage:**
```bash
cursor-rules remove [name] [flags]
```

**Flags:**
- `--type <type>` - Type: `rule`, `command`, `skill`, `agent`, or `hooks` (required unless removing a rule by name)
- `--workdir <dir>` / `-w` (persistent) - Project directory (default: current)
- `--global` (persistent) - Remove from user dirs (~/.cursor/...) instead of project. Can be passed before the subcommand: `cursor-rules --global remove frontend`.

**Examples:**
```bash
# Remove rule preset (default type)
cursor-rules remove frontend

# Remove skill or agent by name
cursor-rules remove my-skill --type skill
cursor-rules remove my-agent --type agent

# Remove installed hooks (no name needed)
cursor-rules remove --type hooks

# Remove from user dirs
cursor-rules --global remove frontend

# Remove from specific project
cursor-rules remove frontend --workdir ~/my-project
```

---

### `cursor-rules effective`

Show effective rules for the current project.

**Usage:**
```bash
cursor-rules effective [flags]
```

**Flags:**
- `--target <target>` - Target format to show
- `--workdir <dir>` - Project directory (default: current)

**Examples:**
```bash
# Show Cursor rules
cursor-rules effective

# Show Copilot instructions
cursor-rules effective --target copilot-instr
```

---

### `cursor-rules transform`

Transform rules between formats.

**Usage:**
```bash
cursor-rules transform [flags]
```

**Flags:**
- `--from <target>` - Source format
- `--to <target>` - Destination format
- `--workdir <dir>` - Project directory (default: current)

**Examples:**
```bash
# Cursor to Copilot
cursor-rules transform --from cursor --to copilot-instr

# Copilot to Cursor
cursor-rules transform --from copilot-instr --to cursor
```

---

### `cursor-rules watch`

Watch package presets and auto-apply changes.

**Usage:**
```bash
cursor-rules watch [flags]
```

**Flags:**
- `--config <file>` - Config file path

**Examples:**
```bash
# Start watcher
cursor-rules watch

# With custom config
cursor-rules watch --config ~/.cursor/rules/config.yaml
# or use $CURSOR_RULES_CONFIG_DIR/config.yaml
```

---

### `cursor-rules init`

Initialize a project with Cursor directories: `.cursor/rules`, `.cursor/commands`, `.cursor/skills`, `.cursor/agents`, and `.cursor/hooks`.

**Usage:**
```bash
cursor-rules init [workdir]
```

**Examples:**
```bash
# Initialize current directory
cursor-rules init

# Initialize specific directory
cursor-rules init ~/my-project
```

### `cursor-rules config init`

Scaffold (or overwrite with `--force`) the `config.yaml` file inside your config directory.

**Usage:**
```bash
cursor-rules config init [flags]
```

**Flags:**
- `-f, --force` - overwrite existing config (backs up the previous file)

**Examples:**
```bash
# Generate config.yaml with current defaults
cursor-rules config init

# Overwrite existing config and keep a backup
cursor-rules config init --force

# Create symlinks from ~/.cursor to CURSOR_*_DIR custom dirs (when those env vars are set)
cursor-rules config link
```

---

## Configuration

### Generate a default config

Use the helper command to scaffold `config.yaml` inside your config directory:

```bash
cursor-rules config init
# Overwrite existing config (backs up the previous file)
cursor-rules config init --force
```

If GNU `stow` is available on your PATH the generated file sets `enableStow: true`; otherwise it remains disabled until stow is installed.

### Directories and destination

Three concepts control where content lives:

| Concept | Primary | Override / shorthand |
|--------|---------|----------------------|
| **Source** (package + config) | `CURSOR_RULES_PACKAGE_DIR` (package dir; when set, config dir defaults to its parent for one-root) | `CURSOR_RULES_CONFIG_DIR` |
| **Destination** (where install/list/remove operate) | `--dir <path\|user>` | `--workdir` / `-w` (path), `--global` (user) |
| **User/global base** | `CURSOR_USER_DIR` (default `~/.cursor`) | `CURSOR_RULES_DIR`, `CURSOR_COMMANDS_DIR`, … (per-feature overrides) |

- **Source:** Where shared packages and config are read from. Set `CURSOR_RULES_PACKAGE_DIR` to your package directory; config dir is then the parent by default, or set `CURSOR_RULES_CONFIG_DIR` to override.
- **Destination:** Either a project path (e.g. `.` or `--workdir /path`) or user dirs (`--dir user` or `--global`). Use `-w`/`--workdir` for a path, or `--global` for user. These flags can be passed before the subcommand: `cursor-rules --global list`, `cursor-rules -w ~/my-project install frontend`.
- **User base:** When using `--global` or `--dir user`, rules/commands/skills/agents/hooks live under `CURSOR_USER_DIR` unless overridden by `CURSOR_RULES_DIR`, etc.

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `CURSOR_RULES_PACKAGE_DIR` | Package directory (when set, config dir defaults to its parent) | `~/.cursor/rules` |
| `CURSOR_RULES_CONFIG_DIR` | Config directory | `~/.cursor/rules` |
| `CURSOR_RULES_SYMLINK` | Use symlinks instead of copies | `false` |
| `CURSOR_USER_DIR` | User/global base for `--global` or `--dir user` | `~/.cursor` |
| `CURSOR_RULES_DIR` | User rules dir (overrides `<user-dir>/rules`) | — |
| `CURSOR_COMMANDS_DIR` | User commands dir | — |
| `CURSOR_SKILLS_DIR` | User skills dir | — |
| `CURSOR_AGENTS_DIR` | User agents dir | — |
| `CURSOR_HOOKS_DIR` | User hooks script dir | — |
| `CURSOR_HOOKS_JSON` | User hooks.json path | — |

**Set in shell:**
```bash
# Bash/Zsh
export CURSOR_RULES_PACKAGE_DIR=~/cursor-rules-shared
export CURSOR_RULES_CONFIG_DIR=~/.cursor/rules
export CURSOR_RULES_SYMLINK=1

# User/global and custom dirs (for install --global, list --global, remove --global)
export CURSOR_USER_DIR=~/.cursor
# Optional: point user context at custom dirs, then run config link
export CURSOR_RULES_DIR=~/my-rules
export CURSOR_COMMANDS_DIR=~/my-commands

# Add to ~/.bashrc or ~/.zshrc for persistence
```

### Custom directories and symlinks

When you set `CURSOR_RULES_DIR`, `CURSOR_COMMANDS_DIR`, etc. to custom paths, the CLI uses those for `--global` operations. To make Cursor (and other tools that only read `~/.cursor`) see the same content, create symlinks from the default user dir to your custom dirs:

```bash
# Set your custom dirs
export CURSOR_RULES_DIR=~/my-rules
export CURSOR_COMMANDS_DIR=~/my-commands

# Create ~/.cursor/rules -> ~/my-rules, ~/.cursor/commands -> ~/my-commands, etc.
cursor-rules config link
```

After that, `~/.cursor/rules` and `~/.cursor/commands` point to your custom dirs.

### Config File

Create `~/.cursor/rules/config.yaml` (or override with `$CURSOR_RULES_CONFIG_DIR`):

```yaml
# Shared presets directory
packageDir: ~/cursor-rules-shared

# Enable file watching
watch: true

# Auto-apply changes
autoApply: true

# Prefer GNU stow when available
enableStow: true

# Default presets to install
presets:
  - frontend
  - backend
  - testing
```

---

## Troubleshooting

### Issue: Command not found

**Problem:** `cursor-rules: command not found`

**Solution:**
```bash
# Check if installed
which cursor-rules

# If not in PATH, add to PATH
export PATH=$PATH:$GOPATH/bin

# Or reinstall
make install
```

---

### Issue: Preset not found

**Problem:** `Error: preset "frontend" not found`

**Solution:**
```bash
# List available presets
cursor-rules list

# Check package directory
ls $CURSOR_RULES_PACKAGE_DIR

# Sync presets
cursor-rules sync
```

---

### Issue: Permission denied

**Problem:** `Error: permission denied`

**Solution:**
```bash
# Check file permissions
ls -la .cursor/rules/

# Fix permissions
chmod 644 .cursor/rules/*.mdc
```

---

### Issue: Invalid frontmatter

**Problem:** `Error: invalid frontmatter format`

**Solution:**
```yaml
# Ensure proper YAML format
---
description: "Rule description"
apply_to: "**/*.ts"
---
Content here
```

---

## Advanced Usage

### Custom Transformers

Create custom transformation logic:

```go
// internal/transform/custom.go
type CustomTransformer struct {
    // ...
}

func (t *CustomTransformer) Transform(fm *yaml.Node, body string) (*yaml.Node, string, error) {
    // Custom transformation logic
    return fm, body, nil
}
```

### Package Manifests

Create complex package structures:

```yaml
# mypackage/cursor-rules-manifest.yaml
version: "1.0"
targets:
  - cursor
  - copilot-instr

overrides:
  copilot-instr:
    defaultMode: "agent"
    defaultTools:
      - "githubRepo"
```

### Watcher Mapping

Configure which presets apply to which projects:

```yaml
# ~/cursor-rules-shared/watcher-mapping.yaml
presets:
  frontend:
    - ~/projects/web-app
    - ~/projects/mobile-app
  backend:
    - ~/projects/api-server
```

---

## FAQ

### Q: Can I use this with GitHub Copilot?

**A:** Yes! Use `--target copilot-instr` or `--target copilot-prompt` to install rules as Copilot instructions or prompts.

### Q: How do I share rules with my team?

**A:** Create a Git repository for your shared presets directory and have team members clone it.

### Q: Can I override rules per project?

**A:** Yes! Project-local rules in `.cursor/rules/` take precedence over shared presets.

### Q: What's the difference between presets and packages?

**A:** Presets are individual files, packages are directories containing multiple related presets.

### Q: How do I update shared presets?

**A:** Run `cursor-rules sync` to pull latest changes from Git, then `cursor-rules sync --apply` to apply to projects.

### Q: Can I use symlinks instead of copying files?

**A:** Yes! Set `CURSOR_RULES_SYMLINK=1` environment variable.

### Q: How do I uninstall a preset?

**A:** Use `cursor-rules remove <preset>` to remove from current project.

### Q: What file formats are supported?

**A:** `.mdc` (Markdown with YAML frontmatter) for Cursor, `.instructions.md` for Copilot instructions, `.prompt.md` for Copilot prompts.

---

## Next Steps

- Read [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for common issues
- Check [CONTRIBUTING.md](../CONTRIBUTING.md) to contribute
- Join our community discussions
- Report bugs on GitHub Issues

---

**Need Help?**

- 📖 [Documentation](https://github.com/ZanzyTHEbar/cursor-rules)
- 🐛 [Report Issues](https://github.com/ZanzyTHEbar/cursor-rules/issues)
- 💬 [Discussions](https://github.com/ZanzyTHEbar/cursor-rules/discussions)

---

**Version:** 1.0  
**Last Updated:** 2025-01-11  
**Maintained By:** cursor-rules development team
