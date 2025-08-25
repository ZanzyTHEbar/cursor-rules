package core

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

var stubTmpl = `---
description: "Shared preset: {{ .Preset }}"
alwaysApply: true
---
@file {{ .SourcePath }}
`

// InstallPreset writes a small stub .mdc in the project's .cursor/rules/
// pointing to the shared preset under sharedDir (default: ~/.cursor-rules).
func InstallPreset(projectRoot, preset string) error {
	sharedDir := DefaultSharedDir()
	src := filepath.Join(sharedDir, preset+".mdc")
	// If preset file not found, allow package-style layout when stow is enabled
	if _, err := os.Stat(src); os.IsNotExist(err) {
		if !(UseGNUStow() && func() bool {
			d := filepath.Join(sharedDir, preset)
			if info, err := os.Stat(d); err == nil && info.IsDir() {
				return true
			}
			return false
		}()) {
			return fmt.Errorf("preset not found: %s (expected %s)", preset, src)
		}
	}

	rulesDir := filepath.Join(projectRoot, ".cursor", "rules")
	if err := os.MkdirAll(rulesDir, 0o755); err != nil {
		return err
	}

	// If symlink/stow behavior requested, prefer that path
	if UseSymlink() || UseGNUStow() {
		return ApplyPresetWithOptionalSymlink(projectRoot, preset, sharedDir)
	}

	dest := filepath.Join(rulesDir, preset+".mdc")
	tmp, err := os.CreateTemp(rulesDir, ".stub-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()

	t := template.Must(template.New("stub").Parse(stubTmpl))
	data := map[string]string{
		"Preset":     preset,
		"SourcePath": src,
	}
	if err := t.Execute(tmp, data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, dest); err != nil {
		return err
	}
	_ = os.Chmod(dest, 0o644)
	return nil
}

// DefaultSharedDir returns ~/.cursor-rules by default; environment overrides allowed.
func DefaultSharedDir() string {
	if v := os.Getenv("CURSOR_RULES_DIR"); v != "" {
		return v
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		if env := os.Getenv("HOME"); env != "" {
			return filepath.Join(env, ".cursor-rules")
		}
		if cwd, cwdErr := os.Getwd(); cwdErr == nil && cwd != "" {
			return filepath.Join(cwd, ".cursor-rules")
		}
		return ".cursor-rules"
	}
	return filepath.Join(home, ".cursor-rules")
}

// InstallPackage installs an entire package directory from sharedDir into the project's
// .cursor/rules. The package is a directory under sharedDir (e.g. "frontend" or "git").
// It supports excluding specific files via the excludes slice and respects a
// .cursor-rules-ignore file placed inside the package which lists patterns to skip.
// By default, packages are flattened into .cursor/rules/. Use noFlatten=true to preserve structure.
func InstallPackage(projectRoot, packageName string, excludes []string, noFlatten bool) error {
	sharedDir := DefaultSharedDir()
	pkgDir := filepath.Join(sharedDir, packageName)
	info, err := os.Stat(pkgDir)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("package not found: %s", pkgDir)
	}

	// Read .cursor-rules-ignore if present in package dir
	ignorePath := filepath.Join(pkgDir, ".cursor-rules-ignore")
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
	rulesDir := filepath.Join(projectRoot, ".cursor", "rules")
	if err := os.MkdirAll(rulesDir, 0o755); err != nil {
		return err
	}

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

		rel, err := filepath.Rel(pkgDir, path)
		if err != nil {
			return err
		}

		// Check ignore patterns (simple glob match)
		for _, pat := range ignorePatterns {
			matched, _ := filepath.Match(pat, rel)
			if matched {
				return nil
			}
		}

		// Destination path preserves package name as prefix to avoid collisions
		// For nested packages (containing "/"), always flatten to avoid deep directory structures
		// By default, all packages are flattened unless noFlatten is explicitly set
		var dest string
		if !noFlatten || strings.Contains(packageName, "/") {
			dest = filepath.Join(rulesDir, filepath.Base(rel))
		} else {
			destName := filepath.Join(packageName, rel)
			dest = filepath.Join(rulesDir, destName)
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}

		// If symlink/stow requested, attempt to use ApplyPresetWithOptionalSymlink semantics
		// For package installs, prefer creating a symlink to the source file when available.
		if UseSymlink() {
			if err := CreateSymlink(path, dest); err == nil {
				return nil
			}
			// else fallthrough to copy
		}

		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.Create(dest)
		if err != nil {
			return err
		}
		defer out.Close()
		if _, err := io.Copy(out, in); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
