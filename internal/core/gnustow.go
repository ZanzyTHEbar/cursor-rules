package core

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ZanzyTHEbar/cursor-rules/internal/errors"
)

// UseSymlink checks environment override to decide whether to create real symlinks
// instead of stub .mdc files. Default is false to preserve existing behavior.
func UseSymlink() bool {
	v := os.Getenv("CURSOR_RULES_SYMLINK")
	return v == "1" || strings.EqualFold(v, "true")
}

// HasStow returns true if GNU stow binary is available on PATH.
func HasStow() bool {
	if _, err := exec.LookPath("stow"); err == nil {
		return true
	}
	return false
}

// UseGNUStow reports whether GNU stow should be used (env requests and stow present).
// WantGNUStow reports whether the user requested GNU stow (regardless of availability).
func WantGNUStow() bool {
	v := strings.ToLower(os.Getenv("CURSOR_RULES_USE_GNUSTOW"))
	return v == "1" || v == "true"
}

// UseGNUStow reports whether GNU stow should be used (requested and available).
func UseGNUStow() bool {
	return WantGNUStow() && HasStow()
}

// CreateSymlink attempts to create a symlink from src -> dest. It will create parent
// directories if necessary and will be idempotent if dest already exists and points
// to the same target.
func CreateSymlink(src, dest string) error {
	// Ensure source exists
	if _, err := os.Stat(src); err != nil {
		return errors.Newf(errors.CodeNotFound, "source not found: %s", src)
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return errors.Wrapf(err, errors.CodeInternal, "create parent dirs for %s", dest)
	}

	// If destination exists, check whether it points to src
	if info, err := os.Lstat(dest); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			current, err := os.Readlink(dest)
			if err == nil && current == src {
				// already correct symlink
				return nil
			}
			// remove stale symlink
			_ = os.Remove(dest)
		} else {
			// file exists and is not a symlink - do not overwrite
			return errors.Newf(errors.CodeAlreadyExists, "destination exists and is not a symlink: %s", dest)
		}
	}

	if err := os.Symlink(src, dest); err != nil {
		return errors.Wrapf(err, errors.CodeInternal, "create symlink %s -> %s", dest, src)
	}
	return nil
}

// ApplyPresetWithOptionalSymlink applies a preset either by creating a stub file (default)
// or by creating a symlink if the environment requests it. If GNU Stow is requested via
// CURSOR_RULES_USE_GNUSTOW and stow is available, attempt to use it (best-effort).
func ApplyPresetWithOptionalSymlink(projectRoot, preset, packageDir string) (InstallStrategy, error) {
	// Ensure target rules directory exists for all strategies (stow/symlink/stub)
	rulesDir := filepath.Join(projectRoot, ".cursor", "rules")
	if err := os.MkdirAll(rulesDir, 0o755); err != nil {
		return StrategyUnknown, err
	}

	src := filepath.Join(packageDir, preset+".mdc")
	if _, err := os.Stat(src); err != nil {
		return StrategyUnknown, errors.Newf(errors.CodeNotFound, "source not found: %s", src)
	}

	dest := filepath.Join(projectRoot, ".cursor", "rules", preset+".mdc")

	// If env requests GNU stow and stow exists, attempt to use it. This expects the
	// packageDir to be structured for stow (package directories). If stow fails,
	// fallback to symlink creation.
	if WantGNUStow() && HasStow() {
		// #nosec G204 - packageDir, rulesDir, and preset are validated before this call
		cmd := exec.Command("stow", "-v", "-d", packageDir, "-t", rulesDir, preset)
		if _, err := cmd.CombinedOutput(); err == nil {
			return StrategyStow, nil
		}
		// else: fall through
	}

	// If user requested symlink behavior (explicitly or via GNU stow), create a symlink
	if UseSymlink() || WantGNUStow() {
		if err := CreateSymlink(src, dest); err == nil {
			return StrategySymlink, nil
		}
	}

	// Default behavior: delegate to shared helper which handles stow -> symlink -> stub
	return ApplySourceToDest(packageDir, src, dest, preset)
}
