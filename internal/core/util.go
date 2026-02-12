package core

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/ZanzyTHEbar/cursor-rules/internal/errors"
	"github.com/ZanzyTHEbar/cursor-rules/internal/security"
)

// WorkingDir returns the current working directory or an error.
func WorkingDir() (string, error) {
	return os.Getwd()
}

// ListProjectPresets lists files in project's .cursor/rules directory (returns file names).
func ListProjectPresets(projectRoot string) ([]string, error) {
	// Safely construct rules directory path
	rulesDir, err := security.SafeJoin(projectRoot, ".cursor", "rules")
	if err != nil {
		return nil, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid project path")
	}

	// Check if directory exists
	if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
		return []string{}, nil // Return empty list when directory doesn't exist
	}

	var out []string
	entries, readErr := fs.ReadDir(os.DirFS(rulesDir), ".")
	if readErr != nil {
		return nil, readErr
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if filepath.Ext(e.Name()) == ".mdc" {
			out = append(out, e.Name())
		}
	}
	return out, nil
}

// InitProject ensures the .cursor/rules directory exists for a project.
func InitProject(projectRoot string) error {
	// Safely construct rules directory path
	rulesDir, err := security.SafeJoin(projectRoot, ".cursor", "rules")
	if err != nil {
		return errors.Wrapf(err, errors.CodeInvalidArgument, "invalid project path")
	}
	return os.MkdirAll(rulesDir, 0o755)
}

// ListProjectCommands lists files in project's .cursor/commands directory (returns file names).
func ListProjectCommands(projectRoot string) ([]string, error) {
	commandsDir := filepath.Join(projectRoot, ".cursor", "commands")

	// Check if directory exists
	if _, err := os.Stat(commandsDir); os.IsNotExist(err) {
		return []string{}, nil // Return empty list when directory doesn't exist
	}

	var out []string
	entries, err := fs.ReadDir(os.DirFS(commandsDir), ".")
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if filepath.Ext(e.Name()) == ".md" {
			out = append(out, e.Name())
		}
	}
	return out, nil
}

// InitProjectCommands ensures the .cursor/commands directory exists for a project.
func InitProjectCommands(projectRoot string) error {
	commandsDir := filepath.Join(projectRoot, ".cursor", "commands")
	return os.MkdirAll(commandsDir, 0o755)
}
