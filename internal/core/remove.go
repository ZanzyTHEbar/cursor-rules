package core

import (
	"fmt"
	"os"
	"path/filepath"
)

// RemovePreset removes the stub file for a preset from the project's .cursor/rules
func RemovePreset(projectRoot, preset string) error {
	rulesDir := filepath.Join(projectRoot, ".cursor", "rules")
	target := filepath.Join(rulesDir, preset+".mdc")
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return fmt.Errorf("preset stub not found: %s", target)
	}
	if err := os.Remove(target); err != nil {
		return err
	}
	return nil
}
