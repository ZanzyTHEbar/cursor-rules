package core

import (
	"fmt"
	"io"
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
	// If symlinking or stow support requested, prefer ApplyPresetWithOptionalSymlink
	if UseSymlink() || strings.ToLower(os.Getenv("CURSOR_RULES_USE_GNUSTOW")) == "1" {
		return ApplyPresetWithOptionalSymlink(projectRoot, normalizedPreset, sharedDir)
	}
	// create stub file that references shared path
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.WriteString(f, "---\n@file "+src+"\n")
	if err != nil {
		return err
	}
	return nil
}
