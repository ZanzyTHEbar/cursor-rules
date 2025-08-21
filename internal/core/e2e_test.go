package core

import (
	"os"
	"path/filepath"
	"testing"
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
	if err := ApplyPresetToProject(projectDir, presetName, sharedDir); err != nil {
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
