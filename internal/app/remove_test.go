package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ZanzyTHEbar/cursor-rules/internal/config"
	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
	apperrors "github.com/ZanzyTHEbar/cursor-rules/internal/errors"
	"github.com/ZanzyTHEbar/cursor-rules/internal/transform"
)

func TestRemoveTargetRemovesOnlyThatTarget(t *testing.T) {
	projectDir := t.TempDir()
	writeInstalledRuleFile(t, filepath.Join(projectDir, ".cursor", "rules", "shared.mdc"), "cursor")
	writeInstalledRuleFile(t, filepath.Join(projectDir, ".github", "instructions", "shared.instructions.md"), "copilot")

	app := New(nil, staticProvider{
		"cursor":        transform.NewCursorTransformer(),
		"copilot-instr": transform.NewCopilotInstructionsTransformer(),
	})
	resp, err := app.Remove(RemoveRequest{
		Name:    "shared",
		Target:  "copilot-instr",
		Workdir: projectDir,
	})
	if err != nil {
		t.Fatalf("Remove returned error: %v", err)
	}
	if len(resp.Matches) != 1 || !resp.Matches[0].Removed || resp.Matches[0].Target != "copilot-instr" {
		t.Fatalf("unexpected matches: %+v", resp.Matches)
	}
	assertExists(t, filepath.Join(projectDir, ".cursor", "rules", "shared.mdc"))
	assertNotExists(t, filepath.Join(projectDir, ".github", "instructions", "shared.instructions.md"))
}

func TestRemoveWithoutTargetRemovesUniqueMatch(t *testing.T) {
	projectDir := t.TempDir()
	writeInstalledRuleFile(t, filepath.Join(projectDir, ".github", "instructions", "shared.instructions.md"), "copilot")

	app := New(nil, staticProvider{
		"cursor":        transform.NewCursorTransformer(),
		"copilot-instr": transform.NewCopilotInstructionsTransformer(),
	})
	resp, err := app.Remove(RemoveRequest{Name: "shared", Workdir: projectDir})
	if err != nil {
		t.Fatalf("Remove returned error: %v", err)
	}
	if len(resp.Matches) != 1 || !resp.Matches[0].Removed || resp.Matches[0].Target != "copilot-instr" {
		t.Fatalf("unexpected matches: %+v", resp.Matches)
	}
	assertNotExists(t, filepath.Join(projectDir, ".github", "instructions", "shared.instructions.md"))
}

func TestRemoveWithoutTargetErrorsOnAmbiguity(t *testing.T) {
	projectDir := t.TempDir()
	writeInstalledRuleFile(t, filepath.Join(projectDir, ".cursor", "rules", "shared.mdc"), "cursor")
	writeInstalledRuleFile(t, filepath.Join(projectDir, ".github", "instructions", "shared.instructions.md"), "copilot")

	app := New(nil, staticProvider{
		"cursor":        transform.NewCursorTransformer(),
		"copilot-instr": transform.NewCopilotInstructionsTransformer(),
	})
	_, err := app.Remove(RemoveRequest{Name: "shared", Workdir: projectDir})
	if err == nil {
		t.Fatal("expected ambiguity error, got nil")
	}
	if apperrors.CodeOf(err) != apperrors.CodeFailedPrecondition {
		t.Fatalf("unexpected error code: %v", apperrors.CodeOf(err))
	}
	if got := err.Error(); got == "" || !strings.Contains(got, "ambiguous") || !strings.Contains(got, "cursor") || !strings.Contains(got, "copilot-instr") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestRemoveWithoutTargetDoesNothingWhenNoMatchExists(t *testing.T) {
	projectDir := t.TempDir()
	app := New(nil, staticProvider{"cursor": transform.NewCursorTransformer()})
	resp, err := app.Remove(RemoveRequest{Name: "missing", Workdir: projectDir})
	if err != nil {
		t.Fatalf("Remove returned error: %v", err)
	}
	if len(resp.Matches) != 0 {
		t.Fatalf("expected no matches, got %+v", resp.Matches)
	}
}

func TestRemoveRuleTargetsByExplicitTarget(t *testing.T) {
	projectDir := t.TempDir()
	writeInstalledRuleFile(t, filepath.Join(projectDir, ".github", "prompts", "review.prompt.md"), "prompt")
	writeInstalledRuleFile(t, filepath.Join(projectDir, ".opencode", "rules", "review.mdc"), "opencode")

	app := New(nil, staticProvider{
		"copilot-prompt": transform.NewCopilotPromptsTransformer(),
		"opencode-rules": transform.NewOpenCodeRulesTransformer(),
	})

	resp, err := app.Remove(RemoveRequest{Name: "review", Target: "copilot-prompt", Workdir: projectDir})
	if err != nil {
		t.Fatalf("Remove copilot-prompt returned error: %v", err)
	}
	if len(resp.Matches) != 1 || !resp.Matches[0].Removed || resp.Matches[0].Target != "copilot-prompt" {
		t.Fatalf("unexpected remove response: %+v", resp)
	}
	assertNotExists(t, filepath.Join(projectDir, ".github", "prompts", "review.prompt.md"))
	assertExists(t, filepath.Join(projectDir, ".opencode", "rules", "review.mdc"))

	resp, err = app.Remove(RemoveRequest{Name: "review", Target: "opencode-rules", Workdir: projectDir})
	if err != nil {
		t.Fatalf("Remove opencode-rules returned error: %v", err)
	}
	if len(resp.Matches) != 1 || !resp.Matches[0].Removed || resp.Matches[0].Target != "opencode-rules" {
		t.Fatalf("unexpected remove response: %+v", resp)
	}
	assertNotExists(t, filepath.Join(projectDir, ".opencode", "rules", "review.mdc"))
}

func TestRemoveUsesAllProvidersForFutureTargets(t *testing.T) {
	p1 := stubRemoveProvider{target: "future-a", kind: resourceKindRule, installed: []string{"shared"}}
	p2 := stubRemoveProvider{target: "future-b", kind: resourceKindRule, installed: []string{"shared"}}
	registry := &nativeResourceRegistry{
		ordered: []nativeResourceProvider{
			p1,
			p2,
		},
		kindOrdered: []string{resourceKindRule},
		byTarget: map[string]nativeResourceProvider{
			"future-a": p1,
			"future-b": p2,
		},
		byKind: map[string][]nativeResourceProvider{
			resourceKindRule: {
				p1,
				p2,
			},
		},
	}
	app := &App{Resources: registry}
	_, err := app.Remove(RemoveRequest{Name: "shared", Workdir: t.TempDir()})
	if err == nil || !strings.Contains(err.Error(), "future-a") || !strings.Contains(err.Error(), "future-b") {
		t.Fatalf("expected ambiguity across future targets, got: %v", err)
	}
}

func TestRemoveConfiguredDoesNotImplicitlyMatchHooks(t *testing.T) {
	projectDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(projectDir, ".cursor", "hooks"), 0o755); err != nil {
		t.Fatalf("mkdir hooks dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, ".cursor", "hooks.json"), []byte(`{"version":1,"hooks":{}}`), 0o644); err != nil {
		t.Fatalf("write hooks.json: %v", err)
	}

	app := New(nil, staticProvider{"cursor": transform.NewCursorTransformer()})
	resp, err := app.Remove(RemoveRequest{Name: "configured", Workdir: projectDir})
	if err != nil {
		t.Fatalf("Remove returned error: %v", err)
	}
	if len(resp.Matches) != 0 {
		t.Fatalf("expected no matches, got %+v", resp.Matches)
	}
	assertExists(t, filepath.Join(projectDir, ".cursor", "hooks.json"))
	assertExists(t, filepath.Join(projectDir, ".cursor", "hooks"))
}

func TestRemoveTypeHooksWithoutNameRemovesHooks(t *testing.T) {
	projectDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(projectDir, ".cursor", "hooks"), 0o755); err != nil {
		t.Fatalf("mkdir hooks dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, ".cursor", "hooks.json"), []byte(`{"version":1,"hooks":{}}`), 0o644); err != nil {
		t.Fatalf("write hooks.json: %v", err)
	}

	app := New(nil, staticProvider{"cursor": transform.NewCursorTransformer()})
	resp, err := app.Remove(RemoveRequest{Type: "hooks", Workdir: projectDir})
	if err != nil {
		t.Fatalf("Remove returned error: %v", err)
	}
	if len(resp.Matches) != 1 || !resp.Matches[0].Removed || resp.Matches[0].Target != "hooks" {
		t.Fatalf("unexpected matches: %+v", resp.Matches)
	}
	assertNotExists(t, filepath.Join(projectDir, ".cursor", "hooks.json"))
	assertNotExists(t, filepath.Join(projectDir, ".cursor", "hooks"))
}

type stubRemoveProvider struct {
	target    string
	kind      string
	installed []string
}

func (p stubRemoveProvider) Kind() string   { return p.kind }
func (p stubRemoveProvider) Target() string { return p.target }
func (p stubRemoveProvider) OutputDir(projectRoot string, _ *config.Config, _ bool) string {
	return filepath.Join(projectRoot, p.target)
}
func (p stubRemoveProvider) RequiresName() bool                                     { return true }
func (p stubRemoveProvider) ListAvailable(string, *config.Config) ([]string, error) { return nil, nil }
func (p stubRemoveProvider) ListInstalled(string, *config.Config, bool) ([]string, error) {
	return p.installed, nil
}
func (p stubRemoveProvider) Install(string, string, string, *config.Config, nativeResourceInstallOptions) (core.InstallStrategy, error) {
	return core.StrategyCopy, nil
}
func (p stubRemoveProvider) PlanInstallAll(string, *config.Config) ([]nativeResourceInstallAllPlan, error) {
	return nil, nil
}
func (p stubRemoveProvider) IncludeInDefaultInstallAll() bool { return false }
func (p stubRemoveProvider) DetectDefaultTarget(string, string, *config.Config) (string, bool, error) {
	return "", false, nil
}
func (p stubRemoveProvider) Remove(string, string, *config.Config, bool) (bool, error) {
	return true, nil
}

func writeInstalledRuleFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}

func assertExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected %s to exist: %v", path, err)
	}
}

func assertNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected %s to be removed, got: %v", path, err)
	}
}
