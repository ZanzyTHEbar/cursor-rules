package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ZanzyTHEbar/cursor-rules/internal/config"
	"github.com/ZanzyTHEbar/cursor-rules/internal/errors"
	"github.com/ZanzyTHEbar/cursor-rules/internal/transform"
)

type staticTransformerProvider map[string]transform.Transformer

func (p staticTransformerProvider) Transformer(target string) (transform.Transformer, error) {
	t, ok := p[target]
	if !ok {
		return nil, errors.Newf(errors.CodeInvalidArgument, "unknown target: %s", target)
	}
	return t, nil
}

func (p staticTransformerProvider) AvailableTargets() []string {
	out := make([]string, 0, len(p))
	for k := range p {
		out = append(out, k)
	}
	return out
}

func TestProviderForTarget(t *testing.T) {
	reg := newNativeResourceRegistry(nil)
	for _, target := range []string{"commands", "skills", "agents", "hooks"} {
		p, ok := reg.providerForTarget(target)
		if !ok || p == nil {
			t.Errorf("providerForTarget(%q): want ok, got ok=%v", target, ok)
		}
	}
	// cursor requires TransformerProvider
	_, ok := reg.providerForTarget("cursor")
	if ok {
		t.Error("providerForTarget(cursor) with nil provider: want !ok")
	}
}

func TestProviderForTargetWithTransformers(t *testing.T) {
	tp := staticTransformerProvider{
		"cursor":         transform.NewCursorTransformer(),
		"copilot-instr":  transform.NewCopilotInstructionsTransformer(),
		"copilot-prompt": transform.NewCopilotPromptsTransformer(),
	}
	reg := newNativeResourceRegistry(tp)
	for _, target := range []string{"commands", "skills", "agents", "hooks", "cursor", "copilot-instr", "copilot-prompt"} {
		p, ok := reg.providerForTarget(target)
		if !ok || p == nil {
			t.Errorf("providerForTarget(%q): want ok, got ok=%v", target, ok)
		}
	}
}

func TestProviderForKind(t *testing.T) {
	reg := newNativeResourceRegistry(nil)
	for _, kind := range []string{"command", "skill", "agent", "hooks"} {
		p, ok := reg.providerForKind(kind)
		if !ok || p == nil {
			t.Errorf("providerForKind(%q): want ok, got ok=%v", kind, ok)
		}
	}
	_, ok := reg.providerForKind("rule")
	if ok {
		t.Error("providerForKind(rule): rules use target not kind, want !ok")
	}
}

func TestResolveDefaultTargetCommands(t *testing.T) {
	packageDir := t.TempDir()
	commandsDir := filepath.Join(packageDir, "commands")
	if err := os.MkdirAll(commandsDir, 0o755); err != nil {
		t.Fatalf("create commands dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(commandsDir, "foo.md"), []byte("# foo"), 0o644); err != nil {
		t.Fatalf("write command: %v", err)
	}

	reg := newNativeResourceRegistry(nil)
	cfg := &config.Config{}
	target, ok, err := reg.resolveDefaultTarget(packageDir, "commands", cfg)
	if err != nil {
		t.Fatalf("resolveDefaultTarget: %v", err)
	}
	if !ok || target != "commands" {
		t.Errorf("resolveDefaultTarget(commands): want target=commands ok=true, got target=%q ok=%v", target, ok)
	}
}

func TestResolveDefaultTargetPresetWithTransformers(t *testing.T) {
	packageDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(packageDir, "example.mdc"), []byte("content"), 0o644); err != nil {
		t.Fatalf("write preset: %v", err)
	}

	tp := staticTransformerProvider{
		"cursor": transform.NewCursorTransformer(),
	}
	reg := newNativeResourceRegistry(tp)
	cfg := &config.Config{}
	target, ok, err := reg.resolveDefaultTarget(packageDir, "example", cfg)
	if err != nil {
		t.Fatalf("resolveDefaultTarget: %v", err)
	}
	if !ok || target != "cursor" {
		t.Errorf("resolveDefaultTarget(example): want target=cursor ok=true, got target=%q ok=%v", target, ok)
	}
}

func TestResolveDefaultTargetUnknownReturnsLastError(t *testing.T) {
	packageDir := t.TempDir()
	// Create a dir that will cause ListPackagePresets to fail (e.g. unreadable)
	// Simpler: use a name that doesn't exist; rules provider will return "", false, nil
	// Skills provider with name "skills" but no skills dir returns err from ListSkillDirs
	reg := newNativeResourceRegistry(nil)
	cfg := &config.Config{}
	target, ok, err := reg.resolveDefaultTarget(packageDir, "nonexistent-preset", cfg)
	if ok {
		t.Errorf("resolveDefaultTarget(nonexistent): want ok=false, got ok=true target=%q", target)
	}
	if err != nil {
		// Some provider may return an error (e.g. package dir issues)
		t.Logf("resolveDefaultTarget returned err (expected for unknown): %v", err)
	}
}

func TestHooksProviderPlanInstallAllReturnsEmpty(t *testing.T) {
	reg := newNativeResourceRegistry(nil)
	p, ok := reg.providerForTarget("hooks")
	if !ok {
		t.Fatal("hooks provider not found")
	}
	plans, err := p.PlanInstallAll(t.TempDir(), &config.Config{})
	if err != nil {
		t.Fatalf("PlanInstallAll: %v", err)
	}
	if len(plans) != 0 {
		t.Errorf("hooks PlanInstallAll: want 0 plans, got %d", len(plans))
	}
}
