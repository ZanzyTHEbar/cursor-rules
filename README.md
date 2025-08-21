# cursor-rules (MVP)

CLI to manage shared Cursor .mdc rule presets across projects.

## Build

make build

## Quick run (install preset into current project):

go run ./cmd/cursor-rules install frontend

## List effective rules:

go run ./cmd/cursor-rules effective --workdir /path/to/project

## Config

Shared presets live in `~/.cursor-rules/`.
Override with `$CURSOR_RULES_DIR`.

## Advanced usage

## VSCode extension

-   Build: `make ext-build`
-   Test: `make ext-test`
-   Package (.vsix): `cd extension && npx @vscode/vsce package`
-   Install locally: `make ext-install` or `code --install-extension extension/*.vsix`

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
