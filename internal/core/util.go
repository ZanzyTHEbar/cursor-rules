package core

import (
	"io/fs"
	"os"
	"path/filepath"
)

// WorkingDir returns the current working directory or an error.
func WorkingDir() (string, error) {
	return os.Getwd()
}

// ListProjectPresets lists files in project's .cursor/rules directory (returns file names).
func ListProjectPresets(projectRoot string) ([]string, error) {
	rulesDir := filepath.Join(projectRoot, ".cursor", "rules")
	var out []string
	entries, err := fs.ReadDir(os.DirFS(rulesDir), ".")
	if err != nil {
		return nil, err
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
	rulesDir := filepath.Join(projectRoot, ".cursor", "rules")
	return os.MkdirAll(rulesDir, 0o755)
}
