package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// UseSymlink checks environment override to decide whether to create real symlinks
// instead of stub .mdc files. Default is false to preserve existing behavior.
func UseSymlink() bool {
	v := os.Getenv("CURSOR_RULES_SYMLINK")
	return v == "1" || strings.ToLower(v) == "true"
}

// HasStow returns true if GNU stow binary is available on PATH.
func HasStow() bool {
	if _, err := exec.LookPath("stow"); err == nil {
		return true
	}
	return false
}

// UseGNUStow reports whether GNU stow should be used (env requests and stow present).
func UseGNUStow() bool {
	return strings.ToLower(os.Getenv("CURSOR_RULES_USE_GNUSTOW")) == "1" && HasStow()
}

// CreateSymlink attempts to create a symlink from src -> dest. It will create parent
// directories if necessary and will be idempotent if dest already exists and points
// to the same target.
func CreateSymlink(src, dest string) error {
	// Ensure source exists
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("source not found: %s", src)
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("failed to create parent dirs for %s: %w", dest, err)
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
			return fmt.Errorf("destination exists and is not a symlink: %s", dest)
		}
	}

	if err := os.Symlink(src, dest); err != nil {
		return fmt.Errorf("failed to create symlink %s -> %s: %w", dest, src, err)
	}
	return nil
}

// ApplyPresetWithOptionalSymlink applies a preset either by creating a stub file (default)
// or by creating a symlink if the environment requests it. If GNU Stow is requested via
// CURSOR_RULES_USE_GNUSTOW and stow is available, attempt to use it (best-effort).
func ApplyPresetWithOptionalSymlink(projectRoot, preset, sharedDir string) error {
	// Ensure target rules directory exists for all strategies (stow/symlink/stub)
	rulesDir := filepath.Join(projectRoot, ".cursor", "rules")
	if err := os.MkdirAll(rulesDir, 0o755); err != nil {
		return err
	}

	src := filepath.Join(sharedDir, preset+".mdc")
	dest := filepath.Join(projectRoot, ".cursor", "rules", preset+".mdc")

	// If env requests GNU stow and stow exists, attempt to use it. This expects the
	// sharedDir to be structured for stow (package directories). If stow fails,
	// fallback to symlink creation.
	if strings.ToLower(os.Getenv("CURSOR_RULES_USE_GNUSTOW")) == "1" && HasStow() {
		cmd := exec.Command("stow", "-v", "-d", sharedDir, "-t", rulesDir, preset)
		if _, err := cmd.CombinedOutput(); err == nil {
			return nil
		}
		// else: fall through
	}

	// If user requested symlink behavior, create a symlink
	if UseSymlink() {
		if err := CreateSymlink(src, dest); err != nil {
			return err
		}
		return nil
	}

	// Default behavior: delegate to shared helper which handles stow -> symlink -> stub
	return ApplySourceToDest(sharedDir, src, dest, preset)
}
