package core

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ApplyPresetToProject copies a shared preset file into the project's .cursor/rules as a stub (@file).
// If the stub already exists, it is left unchanged (idempotent).
func ApplyPresetToProject(projectRoot, preset, sharedDir string) error {
	// ensure source exists
	src := filepath.Join(sharedDir, preset+".mdc")
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("shared preset not found: %s", src)
	}
	rulesDir := filepath.Join(projectRoot, ".cursor", "rules")
	if err := os.MkdirAll(rulesDir, 0o755); err != nil {
		return err
	}
	dest := filepath.Join(rulesDir, preset+".mdc")
	if _, err := os.Stat(dest); err == nil {
		// already exists -> idempotent
		return nil
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
