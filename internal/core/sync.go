package core

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ListPackagePresets returns list of .mdc files found in packageDir.
func ListPackagePresets(packageDir string) ([]string, error) {
	var out []string
	entries, err := fs.ReadDir(os.DirFS(packageDir), ".")
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() {
			// skip directories here; packages are handled via ListPackageDirs
			continue
		}
		if filepath.Ext(e.Name()) == ".mdc" {
			out = append(out, e.Name())
		}
	}
	return out, nil
}

// ListPackageDirs returns directories directly under packageDir which can be
// treated as packages (e.g., "frontend", "git").
func ListPackageDirs(packageDir string) ([]string, error) {
	var out []string
	entries, err := fs.ReadDir(os.DirFS(packageDir), ".")
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		if hasRuleFiles(filepath.Join(packageDir, e.Name())) {
			out = append(out, e.Name())
		}
	}
	return out, nil
}

func hasRuleFiles(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if ext == ".mdc" || ext == ".md" {
			return true
		}
	}
	return false
}

// SyncPackageRepo attempts to git pull if the packageDir is a git repo.
// If not a git repo, it is a no-op.
func SyncPackageRepo(packageDir string) error {
	gitDir := filepath.Join(packageDir, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		// not a git repo; nothing to do
		return nil
	}
	cmd := exec.Command("git", "-C", packageDir, "pull", "--ff-only")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git pull failed: %v: %s", err, string(output))
	}
	return nil
}
