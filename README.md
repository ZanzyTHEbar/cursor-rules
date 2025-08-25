# Cursor Rules Manager

---

<p align="center">
  <img src="extension/assets/icon.webp" alt="Cursor Rules Manager" width="160" height="160" />
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

## Command palette architecture (developer notes)

- The CLI uses a composable "palette" architecture implemented in `cli`.
- `cli.AppContext` contains shared dependencies: `Viper` (config) and `Logger`.
- Commands are authored as factories: `func(*cli.AppContext) *cobra.Command` so they can read configuration from `ctx.Viper` instead of global flags.
- Register all command factories in `cmd/cursor-rules/registry.go` using `cli.Register(...)`.
- Build the root command with `cli.BuildRoot(ctx)` or `cli.NewRoot(ctx, cli.DefaultPalette)` in `main`.

Migration tips:

- Prefer `ctx.Viper.GetString("workdir")` over `rootCmd.Flags().GetString("workdir")`, falling back to the flag if unset.
- For incremental migration, wrap existing `*cobra.Command` values with `cli.FromCommand(cmd)` and register the resulting factory.

Logger integration:

- We use a minimal `cli.Logger` interface (Printf) so you can plug any logger easily.
- By default `AppContext` uses the standard library logger. To use `go-basetools` create
  a logger adapter and pass it to `cli.NewAppContext`.

Example (main.go):

```go
import (
  gblogger "github.com/ZanzyTHEbar/go-basetools/logger"
)

// configure go-basetools logger
gblogger.InitLogger(&gblogger.Config{Logger: gblogger.Logger{Style: "text", Level: "info"}})

// adapt to cli.Logger
ctx := cli.NewAppContext(nil, cli.NewGoBasetoolsAdapter())
root := cli.NewRoot(ctx, cli.DefaultPalette)
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

## Packages

You can organize shared presets as packages under your shared dir (default: `~/.cursor-rules`).
Each package is a directory (for example `frontend/` or `git/`) containing one or more `.mdc` files.

### Nested Package Support

Packages can be organized in nested directory structures of arbitrary depth:

```
~/.cursor-rules/
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

**Note**: Nested packages (containing `/` in the name) are automatically flattened when installed. This means files from `frontend/react/` will be placed directly in `.cursor/rules/` rather than `.cursor/rules/frontend/react/`.

Package installs support exclusions via the `--exclude` flag and a `.cursor-rules-ignore` file placed in the package root.
The `--exclude` flag accepts repeated patterns which are merged with the `.cursor-rules-ignore` patterns.

Example:

```bash
cursor-rules install frontend --exclude "templates/*" --exclude "legacy.mdc"
```

You can also flatten regular package files into the project's `.cursor/rules/` root by passing `--flatten`:

```bash
cursor-rules install frontend --flatten
```

Patterns in `.cursor-rules-ignore` follow simple glob semantics (see `filepath.Match`). Lines starting with `#` are comments.
