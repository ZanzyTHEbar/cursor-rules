package core

import (
	"fmt"
	"os"
	"path/filepath"
)

// DetectProjectType inspects common files and returns a simple project type hint.
// FIXME: This heuristic is not good enough, we need something more sophisticated.
func DetectProjectType(root string) (string, error) {
	checks := []struct {
		file string
		typ  string
	}{
		{"package.json", "node"},
		{"go.mod", "go"},
		{"pyproject.toml", "python"},
		{"requirements.txt", "python"},
		{"Cargo.toml", "rust"},
	}

	for _, c := range checks {
		if _, err := os.Stat(filepath.Join(root, c.file)); err == nil {
			return c.typ, nil
		}
	}
	return "unknown", nil
}

// EffectiveRules walks .cursor/rules and returns merged raw text (simple concatenation).
// FIXME: in v2 we will return a structured object with metadata.
func EffectiveRules(projectRoot string) (string, error) {
	rulesDir := filepath.Join(projectRoot, ".cursor", "rules")
	if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
		return "", fmt.Errorf("no .cursor/rules directory found in project")
	}
	out := ""
	err := filepath.Walk(rulesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		b, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		out += "\n\n---\n# " + filepath.Base(path) + "\n\n" + string(b)
		return nil
	})
	if err != nil {
		return "", err
	}
	return out, nil
}
