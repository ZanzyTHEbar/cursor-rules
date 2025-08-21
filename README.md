# Cursor Rules Manager

---

<p align="center">
  <img src="extension/assets/icon.png" alt="Cursor Rules Manager" width="160" height="160" />
  <br/>
  <em>Manage shared Cursor rule presets from your editor</em>
  <br/>
</p>

CLI & Cursor Extension to manage shared Cursor `.mdc` rule presets across projects.

#### Why

While Cursor has a built-in feature to manage rules, it's not very flexible.

- It's not possible to share rules with others
- It's not possible to override rules for a specific project
- Global rules are brittle, hard to maintain, not always applicable and constantly ignored by the Agent tooling
- Global rules have a terrible UX

I wanted a centralized way to manage rules that would be easy to maintain and share with others, while also being able to override rules for a specific project without all of the copy-pasting between projects.

#### How

CLI and VSCode extension that wraps the CLI.

The CLI is a simple tool that allows you to install, uninstall, list presets and apply them to the current project (some advanced features are available).

The Cursor extension, a VS Code extension, allows you to call the CLI from the command palette.

## Build Everything & Install CLI

> [!NOTE]
> Cursor does not support installing extensions from the command line, so you need to install the extension manually.

`make`

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

# Install a preset into the current project
cursor-rules install frontend

# Show effective rules for the current workspace
cursor-rules effective
```

## Common workflows

### Bootstrap a new project

```bash
cursor-rules sync
cursor-rules install backend
cursor-rules effective > .cursor/effective.md
```

### Keep presets up to date

```bash
cursor-rules sync
# optionally re-apply a preset if structure changed
cursor-rules install shared
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

Shared presets live in `~/.cursor-rules/`.
Override with `$CURSOR_RULES_DIR`.

## Advanced usage

```bash
# Apply multiple presets
cursor-rules install frontend backend tooling

# Generate effective rules for a specific path
cursor-rules effective --workdir /path/to/project

# Watch shared directory and auto-apply to mapped projects
cursor-rules watch
```

### Extension workflows

- Cursor Rules: Sync Shared Presets — fetch updates then pick presets offered on first run
- Cursor Rules: Show Effective Rules — opens a markdown preview of merged rules
- Cursor Rules: Install Preset — prompts for a preset and installs into the current workspace

Environment variables:

-   `CURSOR_RULES_DIR`: override the default shared presets directory (default: `~/.cursor-rules`).
-   `CURSOR_RULES_SYMLINK=1`: when set, `install`/`apply` operations will create real filesystem symlinks instead of writing stub `.mdc` files. Use with caution.
-   `CURSOR_RULES_USE_GNUSTOW=1`: when set and GNU `stow` is available in PATH, the tool will attempt to use `stow` to manage symlinks from the shared directory into project targets. This is best-effort and will fall back to symlinks or stubs if stow fails.

Watcher mapping (auto-apply):

If you run the CLI with `watch` enabled and `autoApply=true` in your config, the watcher will consult a `watcher-mapping.yaml` file in your shared directory to determine which projects to auto-apply presets to. Example `watcher-mapping.yaml` format:

```yaml
presets:
    frontend:
        - /abs/path/to/project1
        - ../relative/path/to/project2
    backend:
        - /abs/path/to/backend
```

Relative project paths are resolved relative to the shared directory. The watcher will only auto-apply presets that appear in this mapping to avoid accidental writes.
