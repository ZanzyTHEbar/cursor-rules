package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ApplyPresetToProject copies a shared preset file into the project's .cursor/rules as a stub (@file).
// If the stub already exists, it is left unchanged (idempotent).
func ApplyPresetToProject(projectRoot, preset, sharedDir string) error {
	// Normalize preset name: remove .mdc extension if present
	normalizedPreset := strings.TrimSuffix(preset, ".mdc")

	// ensure source exists
	src := filepath.Join(sharedDir, normalizedPreset+".mdc")
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("shared preset not found: %s", src)
	}
	rulesDir := filepath.Join(projectRoot, ".cursor", "rules")
	if err := os.MkdirAll(rulesDir, 0o755); err != nil {
		return err
	}
	dest := filepath.Join(rulesDir, normalizedPreset+".mdc")

	// Ensure destination directory exists (handles nested paths like emissium/behaviour/)
	destDir := filepath.Dir(dest)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}

	if _, err := os.Stat(dest); err == nil {
		// already exists -> idempotent
		return nil
	}
	// Delegate to shared ApplySourceToDest which handles stow -> symlink -> stub
	return ApplySourceToDest(sharedDir, src, dest, normalizedPreset)
}
