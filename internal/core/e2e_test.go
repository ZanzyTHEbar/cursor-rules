package core

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Simple end-to-end test for install -> effective -> remove flows
func TestE2EInstallApplyRemove(t *testing.T) {
	sharedDir, err := os.MkdirTemp("", "shared-e2e-")
	if err != nil {
		t.Fatalf("failed to create shared dir: %v", err)
	}
	defer os.RemoveAll(sharedDir)

	presetName := "e2e"
	presetFile := filepath.Join(sharedDir, presetName+".mdc")
	if err := os.WriteFile(presetFile, []byte("# e2e preset"), 0o644); err != nil {
		t.Fatalf("failed to write preset: %v", err)
	}

	projectDir, err := os.MkdirTemp("", "project-e2e-")
	if err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}
	defer os.RemoveAll(projectDir)

	// Apply preset via ApplyPresetToProject
	if _, err := ApplyPresetToProject(projectDir, presetName, sharedDir); err != nil {
		t.Fatalf("ApplyPresetToProject failed: %v", err)
	}

	// Check effective rules
	out, err := EffectiveRules(projectDir)
	if err != nil {
		t.Fatalf("EffectiveRules failed: %v", err)
	}
	if out == "" {
		t.Fatalf("Effective rules empty after apply")
	}

	// Remove
	if err := RemovePreset(projectDir, presetName); err != nil {
		t.Fatalf("RemovePreset failed: %v", err)
	}
}

func TestWatcherAutoApplyMapping(t *testing.T) {
	sharedDir, err := os.MkdirTemp("", "shared-watch-")
	if err != nil {
		t.Fatalf("failed to create shared dir: %v", err)
	}
	defer os.RemoveAll(sharedDir)

	presetName := "frontend"
	presetFile := filepath.Join(sharedDir, presetName+".mdc")
	if err := os.WriteFile(presetFile, []byte("# watcher preset"), 0o644); err != nil {
		t.Fatalf("failed to write preset: %v", err)
	}

	projDir, err := os.MkdirTemp("", "project-watch-")
	if err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}
	defer os.RemoveAll(projDir)

	// write watcher-mapping.yaml
	mapping := []byte("presets:\n  " + presetName + ":\n    - " + projDir + "\n")
	if err := os.WriteFile(filepath.Join(sharedDir, "watcher-mapping.yaml"), mapping, 0o644); err != nil {
		t.Fatalf("failed to write mapping: %v", err)
	}

	// start watcher with context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := StartWatcher(ctx, sharedDir, true); err != nil {
		t.Fatalf("failed to start watcher: %v", err)
	}

	// trigger change
	if err := os.WriteFile(presetFile, []byte("# updated"), 0o644); err != nil {
		t.Fatalf("failed to touch preset: %v", err)
	}

	// wait briefly for watcher to apply
	time.Sleep(500 * time.Millisecond)

	stub := filepath.Join(projDir, ".cursor", "rules", presetName+".mdc")
	if _, err := os.Stat(stub); err != nil {
		t.Fatalf("expected stub created by watcher, err: %v", err)
	}
}
