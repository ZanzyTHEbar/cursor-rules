package app

import (
	"os"
	"path/filepath"

	"github.com/ZanzyTHEbar/cursor-rules/internal/config"
	"github.com/ZanzyTHEbar/cursor-rules/internal/errors"
)

// LinkGlobalRequest describes a request to create symlinks from default user dir to env-specified custom dirs.
type LinkGlobalRequest struct{}

// LinkGlobalResult describes one symlink creation.
type LinkGlobalResult struct {
	Link   string // path where symlink was created
	Target string
	Error  string
}

// LinkGlobalResponse captures link results.
type LinkGlobalResponse struct {
	BaseDir string
	Results []LinkGlobalResult
}

// LinkGlobal creates symlinks at DefaultUserCursorDir() so that ~/.cursor/rules (etc.) point to
// CURSOR_RULES_DIR, CURSOR_COMMANDS_DIR, etc. when those env vars are set. Call this so that
// Cursor sees your custom dirs as the user globals.
func (a *App) LinkGlobal(req LinkGlobalRequest) (*LinkGlobalResponse, error) {
	base := config.DefaultUserCursorDir()
	if err := os.MkdirAll(base, 0o755); err != nil {
		return nil, errors.Wrapf(err, errors.CodeInternal, "create user dir")
	}
	var results []LinkGlobalResult
	pairs := []struct {
		envKey string
		subdir string
	}{
		{config.EnvUserRules, "rules"},
		{config.EnvUserCommands, "commands"},
		{config.EnvUserSkills, "skills"},
		{config.EnvUserAgents, "agents"},
		{config.EnvUserHooks, "hooks"},
	}
	for _, p := range pairs {
		target := os.Getenv(p.envKey)
		if target == "" {
			continue
		}
		target = filepath.Clean(target)
		linkPath := filepath.Join(base, p.subdir)
		if err := createSymlinkTo(linkPath, target); err != nil {
			results = append(results, LinkGlobalResult{Link: linkPath, Target: target, Error: err.Error()})
			continue
		}
		results = append(results, LinkGlobalResult{Link: linkPath, Target: target})
	}
	if target := os.Getenv(config.EnvUserHooksJSON); target != "" {
		target = filepath.Clean(target)
		linkPath := filepath.Join(base, "hooks.json")
		if err := createSymlinkTo(linkPath, target); err != nil {
			results = append(results, LinkGlobalResult{Link: linkPath, Target: target, Error: err.Error()})
		} else {
			results = append(results, LinkGlobalResult{Link: linkPath, Target: target})
		}
	}
	return &LinkGlobalResponse{BaseDir: base, Results: results}, nil
}

// createSymlinkTo creates or replaces linkPath as a symlink to target. Removes existing file/symlink if needed.
func createSymlinkTo(linkPath, target string) error {
	info, err := os.Lstat(linkPath)
	if err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			cur, _ := os.Readlink(linkPath)
			if filepath.Clean(cur) == target {
				return nil
			}
		}
		if err := os.Remove(linkPath); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	return os.Symlink(target, linkPath)
}
