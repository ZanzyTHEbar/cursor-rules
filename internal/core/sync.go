package core

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ZanzyTHEbar/cursor-rules/internal/errors"
)

// ResolveRulesPackageDir returns the directory that should be treated as the
// rules source. If a dedicated `rules/` subdirectory exists under packageDir,
// that subtree is authoritative. Otherwise packageDir itself is treated as the
// rules source for backward compatibility.
func ResolveRulesPackageDir(packageDir string) string {
	rulesDir := filepath.Join(packageDir, "rules")
	info, err := os.Stat(rulesDir)
	if err == nil && info.IsDir() {
		return rulesDir
	}
	return packageDir
}

// ListPackagePresets returns list of .mdc files found in packageDir.
func ListPackagePresets(packageDir string) ([]string, error) {
	packageDir = ResolveRulesPackageDir(packageDir)
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
	originalPackageDir := packageDir
	packageDir = ResolveRulesPackageDir(packageDir)
	isCompatibilityRoot := packageDir == originalPackageDir
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
		if isCompatibilityRoot {
			switch e.Name() {
			case defaultCommandsSubdir, defaultSkillsSubdir, defaultAgentsSubdir, legacyAgentsSubdir, defaultHooksSubdir:
				continue
			}
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
		return errors.Wrapf(err, errors.CodeInternal, "git pull failed: %s", string(output))
	}
	return nil
}
