package core

import (
	"fmt"
	"os"
	"path/filepath"
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
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	t := template.Must(template.New("stub").Parse(stubTmpl))
	data := map[string]string{
		"Preset":     preset,
		"SourcePath": src,
	}
	if err := t.Execute(f, data); err != nil {
		return err
	}
	return nil
}

// DefaultSharedDir returns ~/.cursor-rules by default; environment overrides allowed.
func DefaultSharedDir() string {
	if v := os.Getenv("CURSOR_RULES_DIR"); v != "" {
		return v
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cursor-rules")
}
