package core

import (
	"os"
	"path/filepath"
)

// RemovePreset removes the stub file for a preset from the project's .cursor/rules
func RemovePreset(projectRoot, preset string) error {
	rulesDir := filepath.Join(projectRoot, ".cursor", "rules")
	target := filepath.Join(rulesDir, preset+".mdc")
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return nil // Return nil if file doesn't exist (idempotent)
	}
	if err := os.Remove(target); err != nil {
		return err
	}
	return nil
}

// RemoveCommand removes a command file from the project's .cursor/commands
func RemoveCommand(projectRoot, command string) error {
	commandsDir := filepath.Join(projectRoot, ".cursor", "commands")
	target := filepath.Join(commandsDir, command+".md")
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return nil // Return nil if file doesn't exist (idempotent)
	}
	if err := os.Remove(target); err != nil {
		return err
	}
	return nil
}
