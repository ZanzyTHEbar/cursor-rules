package core

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestListHookPresets(t *testing.T) {
	packageDir, err := os.MkdirTemp("", "hooks-pkg-")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(packageDir)
	hooksRoot := filepath.Join(packageDir, "hooks")
	if err := os.MkdirAll(hooksRoot, 0o755); err != nil {
		t.Fatalf("create hooks root: %v", err)
	}
	preset1 := filepath.Join(hooksRoot, "audit")
	if err := os.MkdirAll(preset1, 0o755); err != nil {
		t.Fatalf("create preset dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(preset1, "hooks.json"), []byte(`{"version":1,"hooks":{}}`), 0o644); err != nil {
		t.Fatalf("write hooks.json: %v", err)
	}
	preset2 := filepath.Join(hooksRoot, "format")
	if err := os.MkdirAll(preset2, 0o755); err != nil {
		t.Fatalf("create preset dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(preset2, "hooks.json"), []byte(`{"version":1,"hooks":{}}`), 0o644); err != nil {
		t.Fatalf("write hooks.json: %v", err)
	}
	// dir without hooks.json is ignored
	if err := os.MkdirAll(filepath.Join(hooksRoot, "empty"), 0o755); err != nil {
		t.Fatalf("create empty: %v", err)
	}

	names, err := ListHookPresets(packageDir, "")
	if err != nil {
		t.Fatalf("ListHookPresets: %v", err)
	}
	if len(names) != 2 {
		t.Fatalf("expected 2 presets, got %d: %v", len(names), names)
	}
	if !slices.Contains(names, "audit") || !slices.Contains(names, "format") {
		t.Fatalf("expected audit and format, got %v", names)
	}
}

func TestInstallHookPresetToProject(t *testing.T) {
	packageDir, err := os.MkdirTemp("", "hooks-pkg-")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(packageDir)
	presetDir := filepath.Join(packageDir, "hooks", "format-preset")
	if err := os.MkdirAll(presetDir, 0o755); err != nil {
		t.Fatalf("create preset dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(presetDir, "scripts"), 0o755); err != nil {
		t.Fatalf("create scripts: %v", err)
	}
	if err := os.WriteFile(filepath.Join(presetDir, "scripts", "format.sh"), []byte("#!/bin/bash\necho format"), 0o755); err != nil {
		t.Fatalf("write script: %v", err)
	}
	hookJSON := `{"version":1,"hooks":{"afterFileEdit":[{"command":"./scripts/format.sh"}]}}`
	if err := os.WriteFile(filepath.Join(presetDir, "hooks.json"), []byte(hookJSON), 0o644); err != nil {
		t.Fatalf("write hooks.json: %v", err)
	}

	projectDir, err := os.MkdirTemp("", "project-")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	defer os.RemoveAll(projectDir)

	strategy, err := InstallHookPresetToProject(projectDir, packageDir, "format-preset", "")
	if err != nil {
		t.Fatalf("InstallHookPresetToProject: %v", err)
	}
	if strategy != StrategyCopy && strategy != StrategySymlink {
		t.Fatalf("unexpected strategy: %s", strategy)
	}
	destJSON := filepath.Join(projectDir, ".cursor", "hooks.json")
	if _, err := os.Stat(destJSON); err != nil {
		t.Fatalf("expected hooks.json at %s: %v", destJSON, err)
	}
	destScript := filepath.Join(projectDir, ".cursor", "hooks", "format.sh")
	if _, err := os.Stat(destScript); err != nil {
		t.Fatalf("expected format.sh at %s: %v", destScript, err)
	}
	// Check that command was rewritten to .cursor/hooks/format.sh
	data, _ := os.ReadFile(destJSON)
	if !strings.Contains(string(data), ".cursor") || !strings.Contains(string(data), "format.sh") {
		t.Fatalf("expected rewritten command path in hooks.json, got: %s", string(data))
	}
}

func TestRemoveHookPresetFromProject(t *testing.T) {
	projectDir, err := os.MkdirTemp("", "project-")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	defer os.RemoveAll(projectDir)
	if err := os.MkdirAll(filepath.Join(projectDir, ".cursor", "hooks"), 0o755); err != nil {
		t.Fatalf("create hooks dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, ".cursor", "hooks.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("write hooks.json: %v", err)
	}

	if err := RemoveHookPresetFromProject(projectDir); err != nil {
		t.Fatalf("RemoveHookPresetFromProject: %v", err)
	}
	if _, err := os.Stat(filepath.Join(projectDir, ".cursor", "hooks.json")); !os.IsNotExist(err) {
		t.Fatal("expected hooks.json to be removed")
	}
	if _, err := os.Stat(filepath.Join(projectDir, ".cursor", "hooks")); !os.IsNotExist(err) {
		t.Fatal("expected hooks dir to be removed")
	}
}
