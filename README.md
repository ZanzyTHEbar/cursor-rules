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


