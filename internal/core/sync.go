package core

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
)

// ListSharedPresets returns list of .mdc files found in sharedDir
func ListSharedPresets(sharedDir string) ([]string, error) {
	var out []string
	entries, err := fs.ReadDir(os.DirFS(sharedDir), ".")
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() {
			// skip directories here; packages are handled via ListSharedPackages
			continue
		}
		if filepath.Ext(e.Name()) == ".mdc" {
			out = append(out, e.Name())
		}
	}
	return out, nil
}

// ListSharedPackages returns directories directly under sharedDir which can be
// treated as packages (e.g., "frontend", "git").
func ListSharedPackages(sharedDir string) ([]string, error) {
	var out []string
	entries, err := fs.ReadDir(os.DirFS(sharedDir), ".")
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() {
			out = append(out, e.Name())
		}
	}
	return out, nil
}

// SyncSharedRepo attempts to git pull if the sharedDir is a git repo.
// If not a git repo, it is a no-op.
func SyncSharedRepo(sharedDir string) error {
	gitDir := filepath.Join(sharedDir, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		// not a git repo; nothing to do
		return nil
	}
	cmd := exec.Command("git", "-C", sharedDir, "pull", "--ff-only")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git pull failed: %v: %s", err, string(output))
	}
	return nil
}
