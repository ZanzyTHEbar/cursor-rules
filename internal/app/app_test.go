package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ZanzyTHEbar/cursor-rules/internal/errors"
	"github.com/ZanzyTHEbar/cursor-rules/internal/transform"
)

type staticProvider map[string]transform.Transformer

func (p staticProvider) Transformer(target string) (transform.Transformer, error) {
	t, ok := p[target]
	if !ok {
		return nil, errors.Newf(errors.CodeInvalidArgument, "unknown target: %s", target)
	}
	return t, nil
}

func (p staticProvider) AvailableTargets() []string {
	out := make([]string, 0, len(p))
	for k := range p {
		out = append(out, k)
	}
	return out
}

func TestListRulesUsesPackageDir(t *testing.T) {
	packageDir := t.TempDir()
	configDir := t.TempDir() // empty so no config file; LoadConfig returns defaults
	t.Setenv("CURSOR_RULES_PACKAGE_DIR", packageDir)
	t.Setenv("CURSOR_RULES_CONFIG_DIR", configDir)
	if err := os.WriteFile(filepath.Join(packageDir, "example.mdc"), []byte("content"), 0o644); err != nil {
		t.Fatalf("write preset: %v", err)
	}

	a := New(nil, nil)
	resp, err := a.ListRules(ListRequest{})
	if err != nil {
		t.Fatalf("ListRules failed: %v", err)
	}
	if resp.PackageDir != packageDir {
		t.Fatalf("expected packageDir %s, got %s", packageDir, resp.PackageDir)
	}
	if resp.Tree == nil {
		t.Fatal("expected rules tree to be present for default list")
	}
	if len(resp.Tree.Presets) != 1 || resp.Tree.Presets[0] != "example.mdc" {
		t.Fatalf("unexpected presets: %+v", resp.Tree.Presets)
	}
}

func TestListRulesOmitsTreeForNonRuleFilters(t *testing.T) {
	packageDir := t.TempDir()
	configDir := t.TempDir()
	t.Setenv("CURSOR_RULES_PACKAGE_DIR", packageDir)
	t.Setenv("CURSOR_RULES_CONFIG_DIR", configDir)
	if err := os.MkdirAll(filepath.Join(packageDir, "commands"), 0o755); err != nil {
		t.Fatalf("mkdir commands: %v", err)
	}
	if err := os.WriteFile(filepath.Join(packageDir, "commands", "review.command.mdc"), []byte("---\ndescription: review\n---\nbody"), 0o644); err != nil {
		t.Fatalf("write command: %v", err)
	}

	a := New(nil, nil)
	resp, err := a.ListRules(ListRequest{Kind: resourceKindCommand})
	if err != nil {
		t.Fatalf("ListRules failed: %v", err)
	}
	if resp.Tree != nil {
		t.Fatalf("expected no rules tree for non-rule filter, got %+v", resp.Tree)
	}
	if len(resp.Targets) == 0 || resp.Targets[0].Target != "commands" {
		t.Fatalf("unexpected target entries: %+v", resp.Targets)
	}
}

func TestTransformPreviewSingleFile(t *testing.T) {
	packageDir := t.TempDir()
	configDir := t.TempDir()
	t.Setenv("CURSOR_RULES_PACKAGE_DIR", packageDir)
	t.Setenv("CURSOR_RULES_CONFIG_DIR", configDir)
	content := `---
description: "Example"
apply_to: "**/*.ts"
---
Hello`
	if err := os.WriteFile(filepath.Join(packageDir, "example.mdc"), []byte(content), 0o644); err != nil {
		t.Fatalf("write preset: %v", err)
	}

	provider := staticProvider{
		"copilot-instr":  transform.NewCopilotInstructionsTransformer(),
		"opencode-rules": transform.NewOpenCodeRulesTransformer(),
	}
	a := New(nil, provider)
	resp, err := a.TransformPreview(TransformRequest{
		Name:   "example",
		Target: "copilot-instr",
	})
	if err != nil {
		t.Fatalf("TransformPreview failed: %v", err)
	}
	if len(resp.Items) != 1 {
		t.Fatalf("expected 1 transform item, got %d", len(resp.Items))
	}
	if resp.Items[0].Error != "" {
		t.Fatalf("unexpected transform error: %s", resp.Items[0].Error)
	}
	if resp.Items[0].Output == "" {
		t.Fatalf("expected transformed output")
	}
}
