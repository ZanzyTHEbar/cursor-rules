# Cursor Rules Manager (VS Code/Cursor extension)

---

<p align="center">
  <img src="assets/icon.png" alt="Cursor Rules Manager" width="160" height="160" />
  <br/>
  <em>Manage shared Cursor rule presets from your editor</em>
  <br/>
</p>

Manage shared Cursor `.mdc` rule presets from your editor. This extension is a thin UI over the `cursor-rules` CLI.

### Requirements

- `cursor-rules` CLI installed and available on PATH
- VS Code/Cursor 1.75+ (extension targets `^1.75.0`)

### Commands

- Cursor Rules: Sync Shared Presets — fetch/update shared presets
- Cursor Rules: Show Effective Rules — preview the merged rules as markdown
- Cursor Rules: Install Preset — install a preset into the current workspace

#### Examples

```text
Command Palette → "Cursor Rules: Sync Shared Presets"
Command Palette → "Cursor Rules: Show Effective Rules"
Command Palette → "Cursor Rules: Install Preset" → enter "frontend"
```

### First‑run Experience

On activation the extension samples files in your workspace and offers presets based on detected languages. You can also choose from any presets found by `cursor-rules sync`.

If the CLI is missing, you’ll be prompted with a link to the repository.

### Status & Output

- All operations show a progress notification
- Detailed logs are written to the `Cursor Rules` output channel

### CLI + Extension workflow examples

#### Bootstrap a workspace
```bash
cursor-rules sync
cursor-rules install backend
# In the editor: "Cursor Rules: Show Effective Rules"
```

#### Update presets periodically
```bash
cursor-rules sync
# In the editor: "Cursor Rules: Install Preset" → choose "shared"
```

### Troubleshooting

- “cursor-rules CLI not found”: install the CLI and ensure it’s on PATH. The error dialog has a link to the repository.
- Permission errors when writing `.cursor/*.mdc`: ensure your workspace is writable; see CLI docs for stow/symlink modes.

### Development

- Build, Package, install CLI: `make`
- Install VSIX (Cursor): Command Palette → “Extensions: Install from VSIX…” or use the versionless file created by the packaging step: `extension/$(node -e "console.log(require('./package.json').name)").vsix`


