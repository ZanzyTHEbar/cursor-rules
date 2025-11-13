package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ZanzyTHEbar/cursor-rules/cli"
)

func TestConfigInitCreatesFile(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CURSOR_RULES_DIR", tempDir)
	t.Setenv("PATH", "")

	ctx := cli.NewAppContext(nil, nil)
	cmd := NewConfigCmd(ctx)
	cmd.SetArgs([]string{"init"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("config init failed: %v", err)
	}

	cfgPath := filepath.Join(tempDir, "config.yaml")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("config file not created: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "sharedDir: "+tempDir) {
		t.Fatalf("expected sharedDir entry, got:\n%s", content)
	}
	if !strings.Contains(content, "enableStow: false") {
		t.Fatalf("expected enableStow entry, got:\n%s", content)
	}
}

func TestConfigInitRequiresForce(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CURSOR_RULES_DIR", tempDir)

	cfgPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte("existing"), 0o644); err != nil {
		t.Fatalf("failed to seed config: %v", err)
	}

	ctx := cli.NewAppContext(nil, nil)
	cmd := NewConfigCmd(ctx)
	cmd.SetArgs([]string{"init"})

	if err := cmd.Execute(); err == nil {
		t.Fatalf("expected error when config exists without --force")
	}
}

func TestConfigInitForceCreatesBackup(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CURSOR_RULES_DIR", tempDir)
	t.Setenv("PATH", "")

	cfgPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte("existing"), 0o644); err != nil {
		t.Fatalf("failed to seed config: %v", err)
	}

	ctx := cli.NewAppContext(nil, nil)
	cmd := NewConfigCmd(ctx)
	cmd.SetArgs([]string{"init", "--force"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("config init --force failed: %v", err)
	}

	backups, err := filepath.Glob(cfgPath + ".*.bak")
	if err != nil {
		t.Fatalf("failed to glob backups: %v", err)
	}
	if len(backups) != 1 {
		t.Fatalf("expected one backup file, found %d", len(backups))
	}

	newContent, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("failed to read new config: %v", err)
	}
	if !strings.Contains(string(newContent), "enableStow: false") {
		t.Fatalf("new config missing enableStow entry")
	}
}
