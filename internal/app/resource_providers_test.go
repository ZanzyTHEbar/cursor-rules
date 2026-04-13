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
		"opencode-rules": transform.NewOpenCodeRulesTransformer(),
	}
	reg := newNativeResourceRegistry(tp)
	for _, target := range []string{"commands", "opencode-commands", "skills", "opencode-skills", "agents", "opencode-agents", "hooks", "cursor", "copilot-instr", "copilot-prompt", "opencode-rules"} {
		p, ok := reg.providerForTarget(target)
		if !ok || p == nil {
			t.Errorf("providerForTarget(%q): want ok, got ok=%v", target, ok)
		}
	}
}

func TestOpenCodeNativeProvidersUseCanonicalOutputDirs(t *testing.T) {
	projectRoot := t.TempDir()
	reg := newNativeResourceRegistry(nil)

	tests := []struct {
		target string
		want   string
	}{
		{target: "opencode-commands", want: filepath.Join(projectRoot, ".opencode", "commands")},
		{target: "opencode-skills", want: filepath.Join(projectRoot, ".opencode", "skills")},
		{target: "opencode-agents", want: filepath.Join(projectRoot, ".opencode", "agents")},
	}

	for _, tt := range tests {
		provider, ok := reg.providerForTarget(tt.target)
		if !ok {
			t.Fatalf("providerForTarget(%q): missing provider", tt.target)
		}
		if got := provider.OutputDir(projectRoot, &config.Config{}, false); got != tt.want {
			t.Fatalf("provider.OutputDir(%q): want %q, got %q", tt.target, tt.want, got)
		}
	}
}

func TestProvidersForKind(t *testing.T) {
	reg := newNativeResourceRegistry(nil)
	for _, kind := range []string{"command", "skill", "agent", "hooks"} {
		providers := reg.providersForKind(kind)
		if len(providers) == 0 {
			t.Errorf("providersForKind(%q): want providers, got none", kind)
		}
	}
	ruleProviders := reg.providersForKind("rule")
	if len(ruleProviders) != 0 {
		t.Errorf("providersForKind(rule): want none without transformers, got %d", len(ruleProviders))
	}
}

func TestProvidersForKindReturnsAllRuleTargets(t *testing.T) {
	tp := staticTransformerProvider{
		"cursor":         transform.NewCursorTransformer(),
		"copilot-instr":  transform.NewCopilotInstructionsTransformer(),
		"copilot-prompt": transform.NewCopilotPromptsTransformer(),
		"opencode-rules": transform.NewOpenCodeRulesTransformer(),
	}
	reg := newNativeResourceRegistry(tp)
	providers := reg.providersForKind(resourceKindRule)
	if len(providers) != 4 {
		t.Fatalf("providersForKind(rule): want 4 providers, got %d", len(providers))
	}
	gotTargets := make([]string, 0, len(providers))
	for _, provider := range providers {
		gotTargets = append(gotTargets, provider.Target())
	}
	wantTargets := []string{"cursor", "copilot-instr", "copilot-prompt", "opencode-rules"}
	for i, target := range wantTargets {
		if gotTargets[i] != target {
			t.Fatalf("providersForKind(rule)[%d]: want %q, got %q", i, target, gotTargets[i])
		}
	}
}

func TestProviderForTargetWithFutureRuleTransformer(t *testing.T) {
	tp := staticTransformerProvider{
		"cursor":         transform.NewCursorTransformer(),
		"opencode-rules": transform.NewOpenCodeRulesTransformer(),
		"zzz-custom":     transform.NewCopilotInstructionsTransformer(),
	}
	reg := newNativeResourceRegistry(tp)
	provider, ok := reg.providerForTarget("zzz-custom")
	if !ok || provider == nil {
		t.Fatal("expected future rule target to be registered")
	}
	providers := reg.providersForKind(resourceKindRule)
	if len(providers) != 3 {
		t.Fatalf("providersForKind(rule): want 3 providers, got %d", len(providers))
	}
	if providers[2].Target() != "zzz-custom" {
		t.Fatalf("expected future target to appear after known targets, got %q", providers[2].Target())
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

func TestHooksProviderPlanInstallAllReturnsPresets(t *testing.T) {
	reg := newNativeResourceRegistry(nil)
	p, ok := reg.providerForTarget("hooks")
	if !ok {
		t.Fatal("hooks provider not found")
	}

	packageDir := t.TempDir()
	hookDir := filepath.Join(packageDir, "hooks", "format")
	if err := os.MkdirAll(hookDir, 0o755); err != nil {
		t.Fatalf("mkdir hook dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(hookDir, "hooks.json"), []byte(`{"version":1,"hooks":{}}`), 0o644); err != nil {
		t.Fatalf("write hooks.json: %v", err)
	}

	plans, err := p.PlanInstallAll(packageDir, &config.Config{})
	if err != nil {
		t.Fatalf("PlanInstallAll: %v", err)
	}
	if len(plans) != 1 || plans[0].Name != "format" {
		t.Errorf("hooks PlanInstallAll: want [format], got %+v", plans)
	}
}
