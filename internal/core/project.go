package core

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
)

// DetectProjectType inspects common files and returns a simple project type hint.
// FIXME: This heuristic is not good enough, we need something more sophisticated.
func DetectProjectType(root string) (string, error) {
	// Fast helper to check for file existence
	exists := func(p string) bool {
		_, err := os.Stat(filepath.Join(root, p))
		return err == nil
	}

	// Monorepo / workspace indicators (prefer more specific detection)
	if exists("lerna.json") || exists("pnpm-workspace.yaml") || exists("pnpm-workspace.yml") || exists("turbo.json") {
		return "node-monorepo", nil
	}

	// Language-specific checks (ordered by clarity)
	if exists("go.mod") {
		return "go", nil
	}

	if exists("pyproject.toml") || exists("requirements.txt") {
		return "python", nil
	}

	if exists("Cargo.toml") {
		return "rust", nil
	}

	// Node ecosystem - look for package.json or lockfiles
	if exists("package.json") || exists("package-lock.json") || exists("yarn.lock") || exists("pnpm-lock.yaml") || exists("pnpm-lock.yml") {
		// Distinguish workspace (package.json with workspaces) if possible
		pkgPath := filepath.Join(root, "package.json")
		if _, err := os.Stat(pkgPath); err == nil {
			data, err := os.ReadFile(pkgPath)
			if err == nil {
				if bytes.Contains(data, []byte("\"workspaces\"")) || bytes.Contains(data, []byte("\"packages\"")) {
					return "node-monorepo", nil
				}
			}
		}
		return "node", nil
	}

	// Fallback: look for generic build files
	if exists("Makefile") {
		return "make", nil
	}
	if exists("Dockerfile") {
		return "docker", nil
	}
	if exists(".gitlab-ci.yml") || exists(".github/workflows") {
		return "ci", nil
	}

	return "unknown", nil
}

// EffectiveRules walks .cursor/rules and returns merged raw text (deterministic order).
// FIXME: in v2 we will return a structured object with metadata.
func EffectiveRules(projectRoot string) (string, error) {
	rulesDir := filepath.Join(projectRoot, ".cursor", "rules")
	if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
		return "", nil // Return empty string when directory doesn't exist
	}
	var files []string
	err := filepath.Walk(rulesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if ext != ".md" && ext != ".mdc" {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(files)
	out := ""
	for _, path := range files {
		b, readErr := os.ReadFile(path)
		if readErr != nil {
			return "", readErr
		}
		out += "\n\n---\n# " + filepath.Base(path) + "\n\n" + string(b)
	}
	return out, nil
}
