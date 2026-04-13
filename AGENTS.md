# AGENTS.md

## Repo Shape
- This is a Go CLI. Entry point: `cmd/cursor-rules/main.go`.
- Keep package boundaries intact: `internal/cli` wires Cobra/Viper and rendering, `internal/app` owns use-case orchestration and destination/config resolution, `internal/core` does filesystem/install/watch work, `internal/transform` owns Cursor/Copilot conversions, `internal/security` guards path handling.
- Native package layouts are not symmetric:
- Rules come from root `*.mdc` files or package directories.
- Commands come from `commands/<name>/` or `commands/<name>.md`, but `install commands ...` writes Cursor-compatible skills into `.cursor/skills/`, not `.cursor/commands/`.
- Skills come from `skills/<name>/SKILL.md`.
- Agents come from `agents/<name>.md`; code still falls back to legacy `agent/` if present.
- Hooks come from `hooks/<preset>/hooks.json` plus scripts.

## Commands
- Build/run: `make build`, `go run ./cmd/cursor-rules <subcommand>`.
- Focused test packages: `go test ./internal/cli/...`, `go test ./internal/core/...`, `go test ./cmd/cursor-rules/...`.
- CI-like verification: `go vet ./...`, `make test`, `make test-coverage`, `make build`, plus `golangci-lint run` if installed.
- Install/sync/watch changes: run `scripts/integration_test.sh`.
- Root e2e workflow: `make test-e2e`.
- Do not use `make test-unit`; it still targets `./cli/...` and currently fails because the code lives under `internal/cli`.

## Config And Runtime Gotchas
- `CURSOR_RULES_PACKAGE_DIR` changes both the source package dir and the default config-dir base: unless `CURSOR_RULES_CONFIG_DIR` is set, config defaults to the parent of the package dir.
- Destination precedence is `--dir user|global` > `--dir <path>` > `--global` > `--workdir`.
- Any CLI command can auto-start the watcher if config has `watch: true`; this happens in root `PersistentPreRunE`, not only in `cursor-rules watch`.
- Watcher auto-apply only writes presets listed in `watcher-mapping.yaml`. `autoApply: true` without a mapping still watches but skips writes.
- `sync --apply` uses `config.presets` if present; otherwise it applies every shared preset.
- `install all` with default `--target cursor` also installs commands, skills, agents, and hooks, not just rules.
- `CURSOR_RULES_SYMLINK=1` and `CURSOR_RULES_USE_GNUSTOW=1` change cursor-target install strategy; `enableStow: true` only takes effect when `stow` is on `PATH`.
- `cursor-rules config link` is the supported way to expose `CURSOR_*_DIR` overrides back under `~/.cursor/...`.

## Tests And Fixtures
- CLI/config tests should set both `CURSOR_RULES_PACKAGE_DIR` and a temp `CURSOR_RULES_CONFIG_DIR`; otherwise user config can leak into tests.
- `scripts/integration_test.sh` is the executable source of truth for `sync`, `--apply`, symlink mode, and watcher-mapping behavior.

## Conventions
- For new filesystem path handling, use `internal/security` (`SafeJoin`, validators) instead of raw joins on user input.
- For app/core domain errors, use `internal/errors` wrappers/helpers rather than ad-hoc `fmt.Errorf`.
- Release automation is semantic-release plus Conventional Commits. `.releaserc` maps `feat` to minor; `fix`, `refactor`, `test`, `build`, `style`, `perf`, and `ci` to patch; `chore` and `no-release` do not cut releases.
- Prefer `go.mod` and the Makefile over `CONTRIBUTING.md` when they disagree: the current toolchain is Go 1.25, and `CONTRIBUTING.md` still mentions stale targets like `make ext-build`.
