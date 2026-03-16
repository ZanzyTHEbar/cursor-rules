package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRemoveFallbackRemovesCommandWhenRuleMissing(t *testing.T) {
	projectDir := t.TempDir()
	commandDir := filepath.Join(projectDir, ".cursor", "commands")
	if err := os.MkdirAll(commandDir, 0o755); err != nil {
		t.Fatalf("create commands dir: %v", err)
	}
	commandPath := filepath.Join(commandDir, "hello.md")
	if err := os.WriteFile(commandPath, []byte("# hello"), 0o644); err != nil {
		t.Fatalf("write command: %v", err)
	}

	app := New(nil, nil)
	resp, err := app.Remove(RemoveRequest{
		Name:    "hello",
		Workdir: projectDir,
	})
	if err != nil {
		t.Fatalf("Remove returned error: %v", err)
	}
	if !resp.RemovedCommand {
		t.Fatalf("expected command to be removed: %+v", resp)
	}
	if resp.RemovedPreset {
		t.Fatalf("expected missing preset not to be reported as removed: %+v", resp)
	}
	if _, err := os.Stat(commandPath); !os.IsNotExist(err) {
		t.Fatalf("expected command file removed, got: %v", err)
	}
}

func TestRemoveTypeRuleWhenPresetDoesNotExist(t *testing.T) {
	projectDir := t.TempDir()
	rulesDir := filepath.Join(projectDir, ".cursor", "rules")
	if err := os.MkdirAll(rulesDir, 0o755); err != nil {
		t.Fatalf("create rules dir: %v", err)
	}

	app := New(nil, nil)
	resp, err := app.Remove(RemoveRequest{
		Name:    "nonexistent",
		Type:    "rule",
		Workdir: projectDir,
	})
	if err != nil {
		t.Fatalf("Remove returned error: %v", err)
	}
	if resp.RemovedPreset {
		t.Fatalf("expected RemovedPreset=false when preset does not exist, got: %+v", resp)
	}
}
