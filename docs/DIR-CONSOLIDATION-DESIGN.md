# Directory Flags and Env Vars: Consolidation Design

## Current Surface (What We Have)

| Kind | Name | Purpose |
|------|------|---------|
| **Persistent flag** | `--workdir` / `-w` | Project root (destination for install/remove, list context) |
| **Persistent flag** | `--config` | Explicit config file path |
| **Per-command flag** | `--global` | Use user dirs (~/.cursor) instead of project (install, install all, list, remove) |
| **Env** | `CURSOR_RULES_PACKAGE_DIR` | Where shared content lives (source) |
| **Env** | `CURSOR_RULES_CONFIG_DIR` | Where config.yaml is read from |
| **Env** | `CURSOR_USER_DIR` | User/global base (~/.cursor) |
| **Env** | `CURSOR_RULES_DIR`, `CURSOR_COMMANDS_DIR`, … (6 total) | Per-feature user dir overrides |
| **Config** | `packageDir` | Package dir when env not set |

So we have: **2 persistent flags**, **1 repeated flag** (`--global`), **8+ env vars**, and **1 config key** that all affect “where things live.”

---

## Goals

- Fewer concepts: one “source” root, one “destination” concept, one “user” root.
- Same or better flexibility for power users.
- Backward compatible: existing env and flags keep working.

---

## Proposal: Three Consolidations

### 1. Single env for source (package + config)

**Idea:** One env var for package dir; config dir is derived when package dir is set so one root is still easy.

- **`CURSOR_RULES_PACKAGE_DIR`** sets the package directory (default: `~/.cursor/rules`).
- **When set:** Config dir defaults to the **parent** of the package dir (e.g. `CURSOR_RULES_PACKAGE_DIR=~/cursor-rules/rules` → config dir `~/cursor-rules`), unless `CURSOR_RULES_CONFIG_DIR` is set.
- **When not set:** Package dir = `~/.cursor/rules`, config dir = same (current behavior).



Result: **one env var** (`CURSOR_RULES_PACKAGE_DIR`) for the common "everything under one root" case; `CURSOR_RULES_CONFIG_DIR` remains as override. (Previously we had `CURSOR_RULES_HOME` as well; it was removed as redundant with `CURSOR_RULES_PACKAGE_DIR`.)

### 2. One destination concept: `--dir` with special value `user`

**Idea:** Treat “destination” as a single concept: either a path (project) or “user” (global).

- Add a **persistent** flag: **`--dir`** (or reuse `--workdir` with extended semantics).
  - **Value = path** (e.g. `.`, `./my-project`, `/abs/path`): that path is the project root (same as today’s `--workdir`).
  - **Value = `user`** (or `global`): destination is user dirs; resolve via `GlobalProjectRoot()` (same as today’s `--global`).
- **Backward compatibility:**  
  - Keep **`--workdir`** / **`-w`**: when set, they set “destination path” (same as `--dir <path>`).  
  - Keep **`--global`**: when set, it forces destination = user (same as `--dir user`).  
  So we don’t remove flags; we add one unified `--dir` and document it as the single knob. Precedence can be: `--dir` overrides; if `--dir` not set, then `--global` → user, else `--workdir` / Viper `workdir` → path.

**CLI help:**  
- “Destination: project path (e.g. `.` or `/path`) or `user` for user dirs. Use `--workdir` / `-w` as shorthand for a path, or `--global` for user.”

Result: **one mental model** (destination = path or user); **one primary flag** (`--dir`); **existing flags** become shorthands.

### 3. User dir: one primary env, overrides optional

**Idea:** Don’t add new env vars; document a clear hierarchy.

- **Primary:** **`CURSOR_USER_DIR`** = user/global base (default `~/.cursor`). All user feature dirs default to `$CURSOR_USER_DIR/rules`, `$CURSOR_USER_DIR/commands`, etc.
- **Overrides:** Keep **`CURSOR_RULES_DIR`**, **`CURSOR_COMMANDS_DIR`**, … for the rare case where one feature lives elsewhere. Document as “optional overrides” in a single table.

Result: **One env var** for 99% of cases; **six optional overrides** for power users; no new vars, just clearer docs and maybe a short “Primary vs override” table in README/USER_GUIDE.

---

## Summary Table (After Consolidation)

| Concept | Primary | Override / Shorthand |
|--------|---------|----------------------|
| **Source (package + config)** | `CURSOR_RULES_PACKAGE_DIR` (when set, config dir = parent) | `CURSOR_RULES_CONFIG_DIR` |
| **Destination** | `--dir <path\|user>` | `--workdir`/`-w` (= path), `--global` (= user) |
| **User/global base** | `CURSOR_USER_DIR` | `CURSOR_RULES_DIR`, … (6 optional) |

---

## Implementation Order

1. **Phase 1 (done):** Use only `CURSOR_RULES_PACKAGE_DIR`; when set, config dir defaults to its parent (one-root). Document `--dir` as the preferred destination flag and map `--workdir`/`--global` to it in docs.
2. **Phase 2 (optional):** Add persistent `--dir` and in the app layer resolve “destination” once from `--dir` or `--workdir`/`--global`; use that single value everywhere.
3. **Phase 3 (docs only):** Update README and USER_GUIDE with the “Primary vs override” table and a single “Directory and destination” section that explains the three concepts above.

This keeps backward compatibility, reduces cognitive load (one home, one destination, one user base), and leaves room to later deprecate the redundant flags if desired.
