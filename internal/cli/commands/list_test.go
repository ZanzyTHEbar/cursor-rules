package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ZanzyTHEbar/cursor-rules/internal/cli"
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

	var buf bytes.Buffer
	ctx := cli.NewAppContext(nil, nil)
	ctx.SetMessenger(cli.NewMessenger(&buf, &buf, "info"))

	cmd := NewListCmd(ctx)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("list command failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "base.mdc") || !strings.Contains(out, "pkg/") {
		t.Fatalf("expected rules tree in output, got:\n%s", out)
	}
	if !strings.Contains(out, "cursor (rule):") || !strings.Contains(out, "  - base") {
		t.Fatalf("expected cursor target section, got:\n%s", out)
	}
}

func TestListCommandShowsMultipleTargetsDistinctly(t *testing.T) {
	shared := t.TempDir()
	configDir := t.TempDir()
	os.Setenv("CURSOR_RULES_PACKAGE_DIR", shared)
	os.Setenv("CURSOR_RULES_CONFIG_DIR", configDir)
	defer func() {
		os.Unsetenv("CURSOR_RULES_PACKAGE_DIR")
		os.Unsetenv("CURSOR_RULES_CONFIG_DIR")
	}()

	writeFile(t, filepath.Join(shared, "base.mdc"), "base")

	var buf bytes.Buffer
	ctx := cli.NewAppContext(nil, nil)
	ctx.SetMessenger(cli.NewMessenger(&buf, &buf, "info"))

	cmd := NewListCmd(ctx)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("list command failed: %v", err)
	}

	out := buf.String()
	for _, section := range []string{"cursor (rule):", "copilot-instr (rule):", "copilot-prompt (rule):", "opencode-rules (rule):"} {
		if !strings.Contains(out, section) {
			t.Fatalf("expected %q in output, got:\n%s", section, out)
		}
	}
}

func TestListCommandTargetFilter(t *testing.T) {
	shared := t.TempDir()
	configDir := t.TempDir()
	os.Setenv("CURSOR_RULES_PACKAGE_DIR", shared)
	os.Setenv("CURSOR_RULES_CONFIG_DIR", configDir)
	defer func() {
		os.Unsetenv("CURSOR_RULES_PACKAGE_DIR")
		os.Unsetenv("CURSOR_RULES_CONFIG_DIR")
	}()

	writeFile(t, filepath.Join(shared, "base.mdc"), "base")

	var buf bytes.Buffer
	ctx := cli.NewAppContext(nil, nil)
	ctx.SetMessenger(cli.NewMessenger(&buf, &buf, "info"))

	cmd := NewListCmd(ctx)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--target", "copilot-prompt"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("list --target failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "copilot-prompt (rule):") {
		t.Fatalf("expected copilot-prompt section, got:\n%s", out)
	}
	if strings.Contains(out, "cursor (rule):") || strings.Contains(out, "copilot-instr (rule):") || strings.Contains(out, "opencode-rules (rule):") {
		t.Fatalf("expected only filtered target section, got:\n%s", out)
	}
}

func TestListCommandKindFilterOmitsRulesTreeForNonRuleKinds(t *testing.T) {
	shared := t.TempDir()
	configDir := t.TempDir()
	os.Setenv("CURSOR_RULES_PACKAGE_DIR", shared)
	os.Setenv("CURSOR_RULES_CONFIG_DIR", configDir)
	defer func() {
		os.Unsetenv("CURSOR_RULES_PACKAGE_DIR")
		os.Unsetenv("CURSOR_RULES_CONFIG_DIR")
	}()

	writeFile(t, filepath.Join(shared, "base.mdc"), "base")
	writeFile(t, filepath.Join(shared, "commands", "review.command.mdc"), "---\ndescription: review\n---\nReview")

	var buf bytes.Buffer
	ctx := cli.NewAppContext(nil, nil)
	ctx.SetMessenger(cli.NewMessenger(&buf, &buf, "info"))

	cmd := NewListCmd(ctx)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--kind", "command"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("list --kind failed: %v", err)
	}

	out := buf.String()
	if strings.Contains(out, "package dir:") || strings.Contains(out, "presets:") || strings.Contains(out, "packages:") {
		t.Fatalf("expected no rules tree for non-rule kind filter, got:\n%s", out)
	}
	if !strings.Contains(out, "commands (command):") || !strings.Contains(out, "opencode-commands (command):") {
		t.Fatalf("expected command target sections, got:\n%s", out)
	}
}

func TestListCommandTargetFilterOmitsRulesTreeForNonRuleTarget(t *testing.T) {
	shared := t.TempDir()
	configDir := t.TempDir()
	os.Setenv("CURSOR_RULES_PACKAGE_DIR", shared)
	os.Setenv("CURSOR_RULES_CONFIG_DIR", configDir)
	defer func() {
		os.Unsetenv("CURSOR_RULES_PACKAGE_DIR")
		os.Unsetenv("CURSOR_RULES_CONFIG_DIR")
	}()

	writeFile(t, filepath.Join(shared, "base.mdc"), "base")
	writeFile(t, filepath.Join(shared, "skills", "deploy", "SKILL.md"), "---\nname: deploy\n---\nbody")

	var buf bytes.Buffer
	ctx := cli.NewAppContext(nil, nil)
	ctx.SetMessenger(cli.NewMessenger(&buf, &buf, "info"))

	cmd := NewListCmd(ctx)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--target", "opencode-skills"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("list --target opencode-skills failed: %v", err)
	}

	out := buf.String()
	if strings.Contains(out, "package dir:") || strings.Contains(out, "presets:") || strings.Contains(out, "packages:") {
		t.Fatalf("expected no rules tree for non-rule target filter, got:\n%s", out)
	}
	if !strings.Contains(out, "opencode-skills (skill):") {
		t.Fatalf("expected filtered target section, got:\n%s", out)
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
