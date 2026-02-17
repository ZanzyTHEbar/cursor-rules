# Cursor Rules Manager

---

<p align="center">
  <img src="extension/assets/icon.webp" alt="Cursor Rules Manager" width="160" height="160" />
  <br/>
  <em>Manage shared Cursor rule presets from your editor</em>
  <br/>
</p>

CLI & Cursor Extension to manage shared Cursor `.mdc` rule presets across projects, with support for GitHub Copilot instructions and prompts.

#### Why

While Cursor has a built-in feature to manage rules, it's not very flexible.

- It's not possible to share rules with others
- It's not possible to override rules for a specific project
- Global rules are brittle, hard to maintain, not always applicable and constantly ignored by the Agent tooling
- Global rules have a terrible UX

I wanted a centralized way to manage rules that would be easy to maintain and share with others, while also being able to override rules for a specific project without all of the copy-pasting between projects.

**New:** Now supports installing rules as GitHub Copilot instructions (`.github/instructions/`) and prompts (`.github/prompts/`) for seamless multi-tool workflows. It also manages **Cursor skills**, **commands**, **subagents (agents)**, and **hooks** from a single package directory.

#### How

CLI and VSCode extension that wraps the CLI.

The CLI is a simple tool that allows you to install, uninstall, list presets and apply them to the current project (some advanced features are available).

The Cursor extension, a VS Code extension, allows you to call the CLI from the command palette.

## Build Everything & Install CLI

> [!NOTE]
> Cursor does not support installing extensions from the command line, so you need to install the extension manually.

### Prerequisites

- Go 1.25.2+ (required)
- Node.js and pnpm (for extension development)
- Docker (optional, for Dev Container)

### Local Build

```bash
make
```

### Dev Container (VS Code/Cursor)

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
- **User base:** When using `--global` or `--dir user`, rules/commands/skills/agents/hooks live under `CURSOR_USER_DIR` unless overridden by `CURSOR_RULES_DIR`, etc.

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




### Extension workflows

- Cursor Rules: Sync Shared Presets — fetch updates then pick presets offered on first run
- Cursor Rules: Show Effective Rules — opens a markdown preview of merged rules
- Cursor Rules: Install Preset — prompts for a preset and installs into the current workspace

Environment variables:

-   `CURSOR_RULES_PACKAGE_DIR`: package directory (default: `~/.cursor/rules`). When set, config dir defaults to its parent so one var gives one root.
-   `CURSOR_RULES_CONFIG_DIR`: override the config directory (default: `~/.cursor/rules`).
-   `CURSOR_RULES_SYMLINK=1`: when set, `install`/`apply` operations will create real filesystem symlinks instead of writing stub `.mdc` files. Use with caution.
-   `CURSOR_RULES_USE_GNUSTOW=1`: when set and GNU `stow` is available in PATH, the tool will attempt to use `stow` to manage symlinks from the package directory into project targets. This is best-effort and will fall back to symlinks or stubs if stow fails.

**User/global dirs and custom directories:**

-   `CURSOR_USER_DIR`: base for user/global context (default: `~/.cursor`). Used with `--global` or `--dir user` for install, list, remove.
-   `CURSOR_RULES_DIR`, `CURSOR_COMMANDS_DIR`, `CURSOR_SKILLS_DIR`, `CURSOR_AGENTS_DIR`, `CURSOR_HOOKS_DIR`, `CURSOR_HOOKS_JSON`: override where user-level rules, commands, skills, agents, hooks, and `hooks.json` live (default: `<CURSOR_USER_DIR>/rules`, etc.). Set these to point at custom dirs, then run `cursor-rules config link` to create symlinks from `~/.cursor/rules` (etc.) to those dirs so Cursor sees them as user globals.

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

### Installation targets

```bash
# Install to Cursor (default)
cursor-rules install frontend

# Install to Copilot Instructions (.github/instructions/)
cursor-rules install frontend --target copilot-instr

# Install to Copilot Prompts (.github/prompts/)
cursor-rules install frontend --target copilot-prompt

# Install to all targets defined in package manifest
cursor-rules install frontend --all-targets

# Install Cursor skills, agents, or hook presets from package dir into project
cursor-rules install my-skill --target cursor-skills
cursor-rules install my-agent --target cursor-agents
cursor-rules install my-hooks --target cursor-hooks
```

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

## Cursor context: rules, commands, skills, agents, hooks

The package directory (default `~/.cursor/rules`) can hold all five Cursor context types:

| Type      | Package dir layout              | Project location              | Install target      |
|-----------|----------------------------------|-------------------------------|---------------------|
| **Rules** | `*.mdc`, `<pkg>/*.mdc`           | `.cursor/rules/`              | `cursor`, `copilot-*` |
| **Commands** | `*.md` or `commands/<name>/` | `.cursor/commands/`           | `cursor-commands`   |
| **Skills**   | `skills/<name>/SKILL.md`     | `.cursor/skills/<name>/`      | `cursor-skills`     |
| **Agents**   | `agents/<name>.md`           | `.cursor/agents/<name>.md`    | `cursor-agents`     |
| **Hooks**    | `hooks/<preset>/hooks.json` + scripts | `.cursor/hooks.json`, `.cursor/hooks/` | `cursor-hooks` |

- **init** creates `.cursor/rules`, `.cursor/commands`, `.cursor/skills`, `.cursor/agents`, and `.cursor/hooks`.
- **list** and **sync** show presets, commands, skills, agents, and hook presets from the package dir.
- **install** with `--target cursor-commands`, `cursor-skills`, `cursor-agents`, or `cursor-hooks` installs from the package dir into the project.
- **remove** supports `--type rule|command|skill|agent|hooks` (e.g. `remove my-skill --type skill`, `remove --type hooks`).
- **install**, **list**, and **remove** support `--global` (or `--dir user`) to operate on user dirs (`~/.cursor/...`) instead of the project. Destination can be set with persistent `--dir`, `--workdir`/`-w`, or `--global`. Override the user base with `CURSOR_USER_DIR` or per-feature with `CURSOR_RULES_DIR`, `CURSOR_COMMANDS_DIR`, etc. Run `cursor-rules config link` to create symlinks from `~/.cursor` to your custom dirs when those env vars are set.

**Hooks:** Installing a hook preset replaces the project’s `.cursor/hooks.json` and populates `.cursor/hooks/` with scripts; script paths in the preset are rewritten to `.cursor/hooks/<name>`. Installing a second hook preset replaces the first unless merge support is added later.

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

**Note**: All packages are flattened by default when installed. This means files from packages will be placed directly in `.cursor/rules/` rather than preserving the package directory structure. Nested packages (containing `/` in the name) are always flattened regardless of flags.

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
