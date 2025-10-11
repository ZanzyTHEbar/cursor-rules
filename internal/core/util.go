package core

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

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
		return nil, fmt.Errorf("invalid project path: %w", err)
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
		return fmt.Errorf("invalid project path: %w", err)
	}
	return os.MkdirAll(rulesDir, 0o755)
}
