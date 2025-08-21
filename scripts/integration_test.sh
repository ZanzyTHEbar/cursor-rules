#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)
cd "$ROOT_DIR"

echo "Building binary..."
make build

BIN=./bin/cursor-rules

echo "Setting up temp dirs..."
SHARED_DIR=$(mktemp -d)
PROJECT_DIR=$(mktemp -d)

echo "shared=$SHARED_DIR"; echo "project=$PROJECT_DIR"

PRESET=frontend
PRESET_FILE="$SHARED_DIR/${PRESET}.mdc"
echo "# preset for $PRESET" > "$PRESET_FILE"

# don't create a git repo here; SyncSharedRepo will be a no-op if not a git repo

export CURSOR_RULES_DIR="$SHARED_DIR"

echo "Syncing shared presets (should list presets)"
"$BIN" sync

echo "Installing preset into project (stub)"
"$BIN" install "$PRESET" -w "$PROJECT_DIR"

STUB="$PROJECT_DIR/.cursor/rules/${PRESET}.mdc"
if [[ ! -f "$STUB" ]]; then
  echo "ERROR: expected stub at $STUB"; exit 2
fi
echo "Stub created: $STUB"

echo "Testing apply via sync --apply"
"$BIN" sync --apply --workdir "$PROJECT_DIR"

echo "Testing symlink install"
rm -f "$STUB"
export CURSOR_RULES_SYMLINK=1
"$BIN" install "$PRESET" -w "$PROJECT_DIR"
if [[ ! -L "$STUB" ]]; then
  echo "ERROR: expected symlink at $STUB"; exit 3
fi
echo "Symlink created and points to: $(readlink -f "$STUB")"

echo "Testing watcher mapping auto-apply"
unset CURSOR_RULES_SYMLINK
MAPPING_FILE="$SHARED_DIR/watcher-mapping.yaml"
cat > "$MAPPING_FILE" <<YML
presets:
  $PRESET:
    - $PROJECT_DIR
YML

# create config enabling watch
CFG=$(mktemp)
cat > "$CFG" <<YAML
sharedDir: "$SHARED_DIR"
watch: true
autoApply: true
presets: []
YAML

echo "Starting watcher (background)"
# Use the dedicated 'watch' command to keep the process alive until signaled
"$BIN" watch --config "$CFG" >/tmp/cursor-rules-watch.log 2>&1 &
WATCH_PID=$!
echo "Watcher PID=$WATCH_PID"

sleep 1

echo "Touching preset to trigger watcher"
echo "\n# touch" >> "$PRESET_FILE"

sleep 2

if [[ -f "$PROJECT_DIR/.cursor/rules/${PRESET}.mdc" ]]; then
  echo "Watcher applied preset to project: OK"
else
  echo "ERROR: watcher did not apply preset"; cat /tmp/cursor-rules-watch.log; kill $WATCH_PID; exit 4
fi

echo "Cleaning up watcher"
kill $WATCH_PID || true

echo "Integration test succeeded"


