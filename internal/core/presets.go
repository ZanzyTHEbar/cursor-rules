package core

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/ZanzyTHEbar/cursor-rules/internal/config"
	"github.com/ZanzyTHEbar/cursor-rules/internal/security"
)

var stubTmpl = `---
description: "Package preset: {{ .Preset }}"
alwaysApply: true
---
@file {{ .SourcePath }}
`

// InstallPreset writes a small stub .mdc in the project's .cursor/rules/
// pointing to the package preset under packageDir (default: ~/.cursor/rules).
func InstallPreset(projectRoot, preset string) error {
	packageDir := DefaultPackageDir()

	// Normalize preset name: remove .mdc extension if present
	normalizedPreset := strings.TrimSuffix(preset, ".mdc")

	// Validate preset name for security (use package validation since presets can have nested paths)
	if err := security.ValidatePackageName(normalizedPreset); err != nil {
		return fmt.Errorf("invalid preset name: %w", err)
	}

	// Safely construct source path
	src, err := security.SafeJoin(packageDir, normalizedPreset+".mdc")
	if err != nil {
		return fmt.Errorf("invalid preset path: %w", err)
	}

	// If preset file not found, allow package-style layout when stow is enabled
	if _, statErr := os.Stat(src); os.IsNotExist(statErr) {
		if !(WantGNUStow() && func() bool {
			d, joinErr := security.SafeJoin(packageDir, normalizedPreset)
			if joinErr != nil {
				return false
			}
			if info, statErr := os.Stat(d); statErr == nil && info.IsDir() {
				return true
			}
			return false
		}()) {
			return fmt.Errorf("preset not found: %s (expected %s)", preset, src)
		}
	}

	// Safely construct rules directory path
	rulesDir, err := security.SafeJoin(projectRoot, ".cursor", "rules")
	if err != nil {
		return fmt.Errorf("invalid project path: %w", err)
	}
	if err := os.MkdirAll(rulesDir, 0o755); err != nil {
		return err
	}

	// If symlink/stow behavior requested, prefer that path
	if UseSymlink() || WantGNUStow() {
		_, err := ApplyPresetWithOptionalSymlink(projectRoot, normalizedPreset, packageDir)
		return err
	}

	// Safely construct destination path
	dest, err := security.SafeJoin(rulesDir, normalizedPreset+".mdc")
	if err != nil {
		return fmt.Errorf("invalid destination path: %w", err)
	}

	// Ensure destination directory exists (handles nested paths like emissium/behavior/)
	destDir := filepath.Dir(dest)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}

	t := template.Must(template.New("stub").Parse(stubTmpl))
	data := map[string]string{
		"Preset":     normalizedPreset,
		"SourcePath": src,
	}
	return AtomicWriteTemplate(destDir, dest, t, data, 0o644)
}

// DefaultPackageDir returns ~/.cursor/rules by default; environment overrides allowed.
func DefaultPackageDir() string {
	if v := strings.TrimSpace(os.Getenv(config.EnvPackageDir)); v != "" {
		return v
	}
	return config.DefaultPackageDir()
}

// InstallPackage installs an entire package directory from packageDir into the project's
// .cursor/rules. The package is a directory under packageDir (e.g. "frontend" or "git").
// It supports excluding specific files via the excludes slice and respects a
// .cursor-rules-ignore file placed inside the package which lists patterns to skip.
// By default, packages are flattened into .cursor/rules/. Use noFlatten=true to preserve structure.
func InstallPackage(projectRoot, packageName string, excludes []string, noFlatten bool) (InstallStrategy, error) {
	return InstallPackageFromPackageDir(projectRoot, DefaultPackageDir(), packageName, excludes, noFlatten)
}

// InstallPackageFromPackageDir installs a package directory from an explicit packageDir into the project's
// .cursor/rules. This is the same behavior as InstallPackage, but allows callers (like the CLI) to
// honor config-provided package directories without relying on environment variables.
func InstallPackageFromPackageDir(projectRoot, packageDir, packageName string, excludes []string, noFlatten bool) (InstallStrategy, error) {
	// Validate package name for security
	if err := security.ValidatePackageName(packageName); err != nil {
		return StrategyUnknown, fmt.Errorf("invalid package name: %w", err)
	}

	// Safely construct package directory path
	pkgDir, err := security.SafeJoin(packageDir, packageName)
	if err != nil {
		return StrategyUnknown, fmt.Errorf("invalid package path: %w", err)
	}
	info, statErr := os.Stat(pkgDir)
	if statErr != nil || !info.IsDir() {
		return StrategyUnknown, fmt.Errorf("package not found: %s", pkgDir)
	}

	// Read .cursor-rules-ignore if present in package dir
	ignorePath, err := security.SafeJoin(pkgDir, ".cursor-rules-ignore")
	if err != nil {
		return StrategyUnknown, fmt.Errorf("invalid ignore file path: %w", err)
	}
	var ignorePatterns []string
	if b, err := os.ReadFile(ignorePath); err == nil {
		lines := strings.Split(string(b), "\n")
		for _, l := range lines {
			l = strings.TrimSpace(l)
			if l == "" || strings.HasPrefix(l, "#") {
				continue
			}
			ignorePatterns = append(ignorePatterns, l)
		}
	}

	// Merge excludes param into ignorePatterns
	for _, ex := range excludes {
		ex = strings.TrimSpace(ex)
		if ex != "" {
			ignorePatterns = append(ignorePatterns, ex)
		}
	}

	// Walk package dir and install each .mdc file unless excluded
	rulesDir, err := security.SafeJoin(projectRoot, ".cursor", "rules")
	if err != nil {
		return StrategyUnknown, fmt.Errorf("invalid rules directory path: %w", err)
	}
	if err := os.MkdirAll(rulesDir, 0o755); err != nil {
		return StrategyUnknown, err
	}

	usedStrategy := StrategyCopy

	err = filepath.Walk(pkgDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".mdc" && filepath.Ext(path) != ".md" {
			return nil
		}

		rel, relErr := filepath.Rel(pkgDir, path)
		if relErr != nil {
			return relErr
		}

		// Validate relative path for security
		if validErr := security.ValidatePath(rel); validErr != nil {
			return fmt.Errorf("invalid file path in package: %w", validErr)
		}

		// Check ignore patterns (simple glob match)
		for _, pat := range ignorePatterns {
			matched, matchErr := filepath.Match(pat, rel)
			if matchErr == nil && matched {
				return nil
			}
		}

		// Destination path preserves package name as prefix to avoid collisions
		// For nested packages (containing "/"), always flatten to avoid deep directory structures
		// By default, all packages are flattened unless noFlatten is explicitly set
		var dest string
		var destErr error
		if !noFlatten || strings.Contains(packageName, "/") {
			dest, destErr = security.SafeJoin(rulesDir, filepath.Base(rel))
		} else {
			destName := filepath.Join(packageName, rel)
			dest, destErr = security.SafeJoin(rulesDir, destName)
		}
		if destErr != nil {
			return fmt.Errorf("invalid destination path: %w", destErr)
		}
		if mkdirErr := os.MkdirAll(filepath.Dir(dest), 0o755); mkdirErr != nil {
			return mkdirErr
		}

		// If symlink/stow requested, attempt to use ApplyPresetWithOptionalSymlink semantics
		// For package installs, prefer creating a symlink to the source file when available.
		if UseSymlink() || WantGNUStow() {
			if symlinkErr := CreateSymlink(path, dest); symlinkErr == nil {
				usedStrategy = StrategySymlink
				return nil
			}
			// else fallthrough to copy
		}

		// #nosec G304 - path is validated above and constructed from trusted sources
		in, openErr := os.Open(path)
		if openErr != nil {
			return openErr
		}
		defer in.Close()
		out, createErr := os.Create(dest)
		if createErr != nil {
			return createErr
		}
		defer out.Close()
		if _, copyErr := io.Copy(out, in); copyErr != nil {
			return copyErr
		}
		return nil
	})
	if err != nil {
		return StrategyUnknown, err
	}
	return usedStrategy, nil
}
