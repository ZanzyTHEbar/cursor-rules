package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/ZanzyTHEbar/cursor-rules/internal/cli"
	"github.com/ZanzyTHEbar/cursor-rules/internal/cli/display"
	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
)

func TestListCommandShowsRulesTree(t *testing.T) {
	shared := t.TempDir()
	configDir := t.TempDir() // empty so LoadConfig uses defaults
	os.Setenv("CURSOR_RULES_PACKAGE_DIR", shared)
	os.Setenv("CURSOR_RULES_CONFIG_DIR", configDir)
	defer func() {
		os.Unsetenv("CURSOR_RULES_PACKAGE_DIR")
		os.Unsetenv("CURSOR_RULES_CONFIG_DIR")
	}()

	writeFile(t, filepath.Join(shared, "base.mdc"), "base")
	writeFile(t, filepath.Join(shared, "pkg", "rule.mdc"), "pkg rule")

	tree, err := core.BuildRulesTree(shared)
	if err != nil {
		t.Fatalf("BuildRulesTree failed: %v", err)
	}
	expected := display.FormatRulesTree(tree) + "\n"

	var buf bytes.Buffer
	ctx := cli.NewAppContext(nil, nil)
	ctx.SetMessenger(cli.NewMessenger(&buf, &buf, "info"))

	cmd := NewListCmd(ctx)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("list command failed: %v", err)
	}

	if buf.String() != expected {
		t.Fatalf("unexpected output:\n--- got ---\n%s\n--- want ---\n%s", buf.String(), expected)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("failed to create dir for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write file %s: %v", path, err)
	}
}
