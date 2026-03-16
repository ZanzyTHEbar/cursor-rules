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
// Deprecated: prefer ListProjectPresetsFrom(rulesDir) with config.EffectiveRulesDir or config.ProjectCursorRulesDir.
func ListProjectPresets(projectRoot string) ([]string, error) {
	rulesDir, err := security.SafeJoin(projectRoot, ".cursor", "rules")
	if err != nil {
		return nil, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid project path")
	}
	return ListProjectPresetsFrom(rulesDir)
}

// ListProjectPresetsFrom lists .mdc files in the given rules directory.
func ListProjectPresetsFrom(rulesDir string) ([]string, error) {
	if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
		return []string{}, nil
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

// InitProject ensures .cursor/rules, .cursor/commands, .cursor/skills, .cursor/agents,
// and .cursor/hooks directories exist for a project.
func InitProject(projectRoot string) error {
	dirs := []string{"rules", "commands", "skills", "agents", "hooks"}
	for _, sub := range dirs {
		dir, err := security.SafeJoin(projectRoot, ".cursor", sub)
		if err != nil {
			return errors.Wrapf(err, errors.CodeInvalidArgument, "invalid project path for .cursor/%s", sub)
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}

// ListProjectCommands lists command entries in project's .cursor/commands directory.
func ListProjectCommands(projectRoot string) ([]string, error) {
	commandsDir, err := security.SafeJoin(projectRoot, ".cursor", "commands")
	if err != nil {
		return nil, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid project path")
	}

	return ListInstalledCommands(commandsDir)
}

// InitProjectCommands ensures the .cursor/commands directory exists for a project.
func InitProjectCommands(projectRoot string) error {
	commandsDir := filepath.Join(projectRoot, ".cursor", "commands")
	return os.MkdirAll(commandsDir, 0o755)
}
