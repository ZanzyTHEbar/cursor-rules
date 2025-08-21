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
# rotate watcher log if too big
LOG=/tmp/cursor-rules-watch.log
if [ -f "$LOG" ]; then
  size=$(stat -c%s "$LOG" || echo 0)
  max=$((5 * 1024 * 1024))
  if [ "$size" -gt "$max" ]; then
    mv "$LOG" "$LOG."$(date +%s)
  fi
fi

"$BIN" watch --config "$CFG" >"$LOG" 2>&1 &
WATCH_PID=$!
echo "Watcher PID=$WATCH_PID (logs -> $LOG)"

echo "Touching preset to trigger watcher"
echo "\n# touch" >> "$PRESET_FILE"

# wait up to 10s for watcher to apply the preset (poll)
applied=0
for i in $(seq 1 20); do
  if [[ -f "$PROJECT_DIR/.cursor/rules/${PRESET}.mdc" ]]; then
    applied=1
    break
  fi
  sleep 0.5
done

if [ "$applied" -eq 1 ]; then
  echo "Watcher applied preset to project: OK"
else
  echo "ERROR: watcher did not apply preset"; echo "--- watcher log ---"; tail -n 200 "$LOG"; kill $WATCH_PID; exit 4
fi

echo "Cleaning up watcher"
kill $WATCH_PID || true

echo "Integration test succeeded"


