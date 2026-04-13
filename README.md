# Cursor Rules Manager

---

<p align="center">
  <img src="assets/icon.webp" alt="Cursor Rules Manager" width="160" height="160" />
  <br/>
  <em>Manage shared Cursor rule presets with ease</em>
  <br/>
</p>

CLI to manage shared Cursor `.mdc` rule presets across projects, with support for GitHub Copilot instructions and prompts.

#### Why

While Cursor has a built-in feature to manage rules, it's not very flexible.

- It's not possible to share rules with others
- It's not possible to override rules for a specific project
- Global rules are brittle, hard to maintain, not always applicable and constantly ignored by the Agent tooling
- Global rules have a terrible UX

I wanted a centralized way to manage rules that would be easy to maintain and share with others, while also being able to override rules for a specific project without all of the copy-pasting between projects.

**New:** Now supports installing rules as GitHub Copilot instructions (`.github/instructions/`) and prompts (`.github/prompts/`) for seamless multi-tool workflows. It also manages **Cursor skills**, **commands**, **subagents (agents)**, and **hooks** from a single package directory.

#### How

The CLI is a simple tool that allows you to install, uninstall, list presets and apply them to the current project (some advanced features are available).

## Build Everything & Install CLI

### Prerequisites

- Go 1.25.2+ (required)
- Docker (optional, for Dev Container)

### Local Build

```bash
make
```

### Dev Container

For a consistent development environment:

1. Open project in VS Code or Cursor
2. Install "Dev Containers" extension
3. Press `F1` → "Dev Containers: Reopen in Container"
4. Container includes Go 1.25.2 and all development tools

See [.devcontainer/BUILD_INSTRUCTIONS.md](.devcontainer/BUILD_INSTRUCTIONS.md) for detailed instructions.

## Quick run (install a preset into current project):

```bash
go run ./cmd/cursor-rules install frontend
```

## List effective rules:

```bash
go run ./cmd/cursor-rules effective --workdir /path/to/project
```

## Quickstart

```bash
# Install the CLI (example)
go install github.com/ZanzyTHEbar/cursor-rules/cmd/cursor-rules@latest

# Sync shared presets from your source (git or local)
cursor-rules sync

# Install a preset into the current project (Cursor)
cursor-rules install frontend

# Install a preset as GitHub Copilot instructions
cursor-rules install frontend --target copilot-instr

# Install a preset as GitHub Copilot prompts
cursor-rules install frontend --target copilot-prompt

# Show effective rules for the current workspace
cursor-rules effective

# Show effective Copilot instructions
cursor-rules effective --target copilot-instr
```

## Common workflows

### Bootstrap a new project

```bash
cursor-rules sync
cursor-rules install backend
cursor-rules effective > .cursor/effective.md
```

### Bootstrap with GitHub Copilot support

```bash
cursor-rules sync
# Install to both Cursor and Copilot
cursor-rules install frontend --target cursor
cursor-rules install frontend --target copilot-instr
# Or use manifest to install to all targets at once
cursor-rules install frontend --all-targets
```

### Keep presets up to date

```bash
cursor-rules sync
# optionally re-apply a preset if structure changed
cursor-rules install shared
```

### Preview transformations before installing

```bash
# See how Cursor rules will be transformed to Copilot format
cursor-rules transform frontend --target copilot-instr
```

### Work with symlinks or GNU stow

```bash
# Use real symlinks instead of stub .mdc files
export CURSOR_RULES_SYMLINK=1
```

> [!NOTE]
> This requires gnustow to be installed

```bash
# Prefer GNU stow if available
export CURSOR_RULES_USE_GNUSTOW=1

cursor-rules install frontend
```

## Config

Shared presets live in `~/.cursor/rules/` by default.
Override with `$CURSOR_RULES_PACKAGE_DIR` or `packageDir` in `config.yaml`.

The config file lives in the config directory (default: `~/.cursor/rules/`), which can be
overridden with `$CURSOR_RULES_CONFIG_DIR` or `--config`.

### Directories and destination

Three concepts control where content lives:

| Concept | Primary | Override / shorthand |
|--------|---------|----------------------|
| **Source** (package + config) | `CURSOR_RULES_PACKAGE_DIR` (package dir; when set, config dir defaults to its parent for one-root) | `CURSOR_RULES_CONFIG_DIR` |
| **Destination** (where install/list/remove operate) | `--dir <path\|user>` | `--workdir` / `-w` (path), `--global` (user) |
| **User/global base** | `CURSOR_USER_DIR` (default `~/.cursor`) | `CURSOR_RULES_DIR`, `CURSOR_COMMANDS_DIR`, … (per-feature overrides) |

- **Source:** Where shared packages and config are read from. Set `CURSOR_RULES_PACKAGE_DIR` to your package directory (e.g. `~/cursor-rules/rules`); config dir is then the parent by default (e.g. `~/cursor-rules`), or set `CURSOR_RULES_CONFIG_DIR` to override.
- **Destination:** Either a project path (e.g. `.` or `--workdir /path`) or user dirs (`--dir user` or `--global`). Use `-w`/`--workdir` for a path, or `--global` for user.
- **User base:** When using `--global` or `--dir user`, rules/commands/skills/agents/hooks live under `CURSOR_USER_DIR` (default `~/.cursor`). Per-feature overrides (`CURSOR_RULES_DIR`, etc.) are supported for advanced use.

### Generate a default config

```bash
cursor-rules config init
```

This writes `config.yaml` in your config directory using the current defaults. Pass `--force` to overwrite (the previous file is backed up automatically). If GNU `stow` is available, the generated file sets `enableStow: true`; otherwise it remains disabled until you install stow.

## Advanced usage

```bash
# Apply multiple presets
cursor-rules install frontend backend tooling

# Generate effective rules for a specific path
cursor-rules effective --workdir /path/to/project

# Watch package directory and auto-apply to mapped projects
cursor-rules watch
```

## Command palette architecture (developer notes)

- The CLI uses a composable "palette" architecture implemented in `cli`.
- `cli.AppContext` contains shared dependencies: `Viper` (config) and `Logger`.
- Commands are authored as factories: `func(*cli.AppContext) *cobra.Command` so they can read configuration from `ctx.Viper` instead of global flags.
- Register all command factories in `internal/cli/commands` using `commands.RegisterAll()`.
- Build and execute the root command via `cli.BuildRoot(...)` or `cli.Execute()` (which calls `BuildRoot`).

Migration tips:

- Prefer `ctx.Viper.GetString("workdir")` over `rootCmd.Flags().GetString("workdir")`, falling back to the flag if unset.
- For incremental migration, wrap existing `*cobra.Command` values with `cli.FromCommand(cmd)` and register the resulting factory.

Logger integration:

- We use a minimal `cli.Logger` interface (Printf) so you can plug any logger easily.
- By default `AppContext` uses the standard library logger.

### Environment variables

-   `CURSOR_RULES_PACKAGE_DIR`: package directory (default: `~/.cursor/rules`). When set, config dir defaults to its parent so one var gives one root.
-   `CURSOR_RULES_CONFIG_DIR`: override the config directory (default: `~/.cursor/rules`).
-   `CURSOR_RULES_SYMLINK=1`: when set, `install`/`apply` operations will create real filesystem symlinks instead of writing stub `.mdc` files. Use with caution.
-   `CURSOR_RULES_USE_GNUSTOW=1`: when set and GNU `stow` is available in PATH, the tool will attempt to use `stow` to manage symlinks from the package directory into project targets. This is best-effort and will fall back to symlinks or stubs if stow fails.

**User/global dirs:**

-   `CURSOR_USER_DIR`: base for user/global context (default: `~/.cursor`). Used with `--global` or `--dir user` for install, list, remove. Most users need only this.
-   `CURSOR_RULES_DIR`, `CURSOR_COMMANDS_DIR`, etc.: optional per-feature overrides. When set, the CLI writes directly to those paths for `--global`. Run `cursor-rules config link` to symlink `~/.cursor/*` to custom dirs so Cursor sees them.

Watcher mapping (auto-apply):

If you run the CLI with `watch` enabled and `autoApply=true` in your config, the watcher will consult a `watcher-mapping.yaml` file in your package directory to determine which projects to auto-apply presets to. Example `watcher-mapping.yaml` format:

```yaml
presets:
    frontend:
        - /abs/path/to/project1
        - ../relative/path/to/project2
    backend:
        - /abs/path/to/backend
```

Relative project paths are resolved relative to the package directory. The watcher will only auto-apply presets that appear in this mapping to avoid accidental writes.

## GitHub Copilot Integration

Cursor Rules Manager now supports installing your Cursor rules as GitHub Copilot instructions and prompts, enabling seamless multi-tool workflows.

### What's the difference?

- **Cursor Rules (`.mdc`)**: Project-scoped AI guidance for Cursor IDE, auto-included based on relevance
- **Copilot Instructions (`.instructions.md`)**: Ambient behavioral rules for GitHub Copilot, auto-merged into all Chat/agent interactions
- **Copilot Prompts (`.prompt.md`)**: Reusable task templates for GitHub Copilot, invoked via slash commands
- **OpenCode Rules (`.mdc`)**: Rule files for the `opencode-rules` plugin, installed into `.opencode/rules/` or `~/.config/opencode/rules/`

### Installation targets

```bash
# Install to Cursor (default)
cursor-rules install frontend

# Install to Copilot Instructions (.github/instructions/)
cursor-rules install frontend --target copilot-instr

# Install to Copilot Prompts (.github/prompts/)
cursor-rules install frontend --target copilot-prompt

# Install to OpenCode rules (.opencode/rules/)
cursor-rules install frontend --target opencode-rules

# Install to all targets defined in package manifest
cursor-rules install frontend --all-targets

# Install commands, skills, agents, or hooks (subcommands)
cursor-rules install commands my-cmd
cursor-rules install commands all
cursor-rules install skills deploy
cursor-rules install skills all
cursor-rules install agents code-reviewer
cursor-rules install hooks my-hooks

# Install native OpenCode commands, skills, or agents
cursor-rules install commands review --target opencode
cursor-rules install skills deploy --target opencode
cursor-rules install agents reviewer --target opencode

cursor-rules install all
```

### List and remove by target

`list` is target-aware. By default it shows shared package content grouped by the concrete install targets it can feed. With `--global` (or `--dir user`), it lists installed user resources instead.

```bash
# Show everything grouped by target
cursor-rules list

# Show only rule targets
cursor-rules list --kind rule

# Show only command targets
cursor-rules list --kind command

# Show one concrete target
cursor-rules list --target opencode-skills

# Show installed global OpenCode rules
cursor-rules list --global --target opencode-rules
```

`remove` is also target-aware. Pass `--target` to remove from one concrete install target. Without `--target`, removal only succeeds when exactly one installed target matches; if the same name exists in multiple targets, the CLI errors and asks for `--target`.

```bash
# Remove a Cursor rule install
cursor-rules remove frontend --target cursor

# Remove a Copilot prompt install of a rule
cursor-rules remove base --target copilot-prompt

# Remove an OpenCode command install
cursor-rules remove review --target opencode-commands

# Remove configured hooks
cursor-rules remove --type hooks

# Remove a global OpenCode skill install
cursor-rules remove deploy --target opencode-skills --global
```

Concrete target names used by `list --target` and `remove --target`:

- `cursor`, `copilot-instr`, `copilot-prompt`, `opencode-rules` for rules
- `commands`, `opencode-commands` for commands
- `skills`, `opencode-skills` for skills
- `agents`, `opencode-agents` for agents
- `hooks` for hooks

Native installs use the higher-level `--target cursor|opencode` selector on `install commands|skills|agents`. `list` and `remove` then operate on the concrete target names above.

### Frontmatter transformation

Cursor rules are automatically transformed to Copilot-compatible formats:

**Cursor `.mdc`:**
```yaml
---
description: "React best practices"
apply_to: "**/*.tsx,**/*.jsx"
priority: 1
alwaysApply: true
---
```

**Copilot `.instructions.md`:**
```yaml
---
description: "React best practices"
applyTo: "**/*.tsx,**/*.jsx"
---
```

**Copilot `.prompt.md`:**
```yaml
---
description: "React best practices"
mode: "chat"
---
```

Key transformations:
- `apply_to` → `applyTo` (renamed)
- `priority`, `alwaysApply` → removed (Copilot doesn't support)
- `mode`, `tools` → added for prompts
- Array globs → comma-separated strings

### Package manifests

Create a `cursor-rules-manifest.yaml` in your package root to define multi-target support:

```yaml
version: "1.0"
targets:
  - cursor
  - copilot-instr
  - copilot-prompt

# Optional: target-specific overrides
overrides:
  copilot-prompt:
    defaultMode: "agent"
    defaultTools: ["githubRepo"]

# Optional: exclusions
exclude:
  - "templates/*"
  - "legacy.mdc"
```

Then install to all targets at once:

```bash
cursor-rules install frontend --all-targets
```

### Preview transformations

Before installing, preview how your rules will be transformed:

```bash
cursor-rules transform frontend --target copilot-instr
```

### View effective rules

See merged rules for any target:

```bash
# Cursor rules
cursor-rules effective

# Copilot instructions
cursor-rules effective --target copilot-instr

# Copilot prompts
cursor-rules effective --target copilot-prompt

# OpenCode rules
cursor-rules effective --target opencode-rules
```

### Migration workflow

1. **Sync existing Cursor rules:**
   ```bash
   cursor-rules sync
   ```

2. **Preview transformation:**
   ```bash
   cursor-rules transform frontend --target copilot-instr
   ```

3. **Install to Copilot (non-destructive):**
   ```bash
   cursor-rules install frontend --target copilot-instr
   ```

4. **Verify in VS Code:**
   - Open Command Palette → "Copilot: Show Custom Instructions"

5. **Test with Copilot Chat:**
   - Ask: "Generate a React component" (should follow instructions)

For detailed examples and best practices, see [docs/copilot-integration.md](docs/copilot-integration.md).

## Shared context: rules, commands, skills, agents, hooks

The package directory (default `~/.cursor/rules`) can hold all five shared context types:

| Type      | Package dir layout              | Project location              | Install command      |
|-----------|----------------------------------|-------------------------------|----------------------|
| **Rules** | `*.mdc`, `<pkg>/*.mdc`           | `.cursor/rules/`              | `install [name]` or `install rules [name]` (use `--target` for copilot) |
| **Commands** | `commands/<name>.md`, `commands/<name>.command.mdc`, or `commands/<name>/` | `.cursor/skills/<name>/SKILL.md` (Cursor-compatible) | `install commands [name\|all]` |
| **Skills**   | `skills/<name>/SKILL.md`     | `.cursor/skills/<name>/`      | `install skills [name\|all]` |
| **Agents**   | `agents/<name>.md`           | `.cursor/agents/<name>.md`    | `install agents [name\|all]` |
| **Hooks**    | `hooks/<preset>/hooks.json` + scripts | `.cursor/hooks.json`, `.cursor/hooks/` | `install hooks [preset]` |

- **init** creates `.cursor/rules`, `.cursor/commands`, `.cursor/skills`, `.cursor/agents`, and `.cursor/hooks`.
- **list** shows package content by default, grouped by concrete target. Use `--kind` and `--target` to filter sections, or `--global` / `--dir user` to list installed user resources.
- **install** uses subcommands: `install [name]` for rules (default), `install commands`, `install skills`, `install agents`, `install hooks`, `install all`.
- **remove** supports `--type rule|command|skill|agent|hooks` and `--target`. Without `--target`, removal only succeeds when the installed name is unique across targets.
- **install**, **list**, and **remove** support `--global` (or `--dir user`) to operate on user dirs (`~/.cursor/...`) instead of the project. Destination can be set with persistent `--dir`, `--workdir`/`-w`, or `--global`. Override the user base with `CURSOR_USER_DIR` or per-feature with `CURSOR_RULES_DIR`, `CURSOR_COMMANDS_DIR`, etc. Run `cursor-rules config link` to create symlinks from `~/.cursor` to your custom dirs when those env vars are set.

**Hooks:** Installing a hook preset replaces the project’s `.cursor/hooks.json` and populates `.cursor/hooks/` with scripts; script paths in the preset are rewritten to `.cursor/hooks/<name>`. Installing a second hook preset replaces the first unless merge support is added later.

**Commands:** Cursor installs convert shared commands into Cursor-compatible skills under `.cursor/skills/`. OpenCode installs keep native command files under `.opencode/commands/`.

### Migration: Subcommand-based install (breaking)

Native resources now use subcommands instead of `--target`:

| Old | New |
|-----|-----|
| `install my-cmd --target commands` | `install commands my-cmd` |
| `install deploy --target skills` | `install skills deploy` |
| `install code-reviewer --target agents` | `install agents code-reviewer` |
| `install my-hooks --target hooks` | `install hooks my-hooks` |
| `install commands` (collection) | `install commands all` |

Rules keep `--target` for output format: `install frontend --target copilot-instr`.
For native resources, `install ... --target opencode` maps to the concrete `list` / `remove` targets `opencode-commands`, `opencode-skills`, and `opencode-agents`.

## Packages

You can organize shared presets as packages under your package dir (default: `~/.cursor/rules`).
Each package is a directory (for example `frontend/` or `git/`) containing one or more `.mdc` files.

### Nested Package Support

Packages can be organized in nested directory structures of arbitrary depth:

```
~/.cursor/rules/
├── frontend/
│   ├── react/
│   │   ├── hooks.mdc
│   │   └── components.mdc
│   └── vue/
│       └── composition.mdc
└── backend/
    └── nodejs/
        └── express/
            ├── middleware.mdc
            └── routes.mdc
```

Install a package into a project:

```bash
# Install a simple package
cursor-rules install frontend

# Install a nested package
cursor-rules install frontend/react

# Install a deeply nested package
cursor-rules install backend/nodejs/express
```

> [!NOTE]
> All packages are flattened by default when installed. This means files from packages will be placed directly in `.cursor/rules/` rather than preserving the package directory structure. Nested packages (containing `/` in the name) are always flattened regardless of flags.

Package installs support exclusions via the `--exclude` flag and a `.cursor-rules-ignore` file placed in the package root.
The `--exclude` flag accepts repeated patterns which are merged with the `.cursor-rules-ignore` patterns.

Example:

```bash
cursor-rules install frontend --exclude "templates/*" --exclude "legacy.mdc"
```

You can preserve the package directory structure by passing `--no-flatten` (or `-n`):

```bash
cursor-rules install frontend --no-flatten
cursor-rules install frontend -n
```

Patterns in `.cursor-rules-ignore` follow simple glob semantics (see `filepath.Match`). Lines starting with `#` are comments.
