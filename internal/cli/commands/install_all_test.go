package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ZanzyTHEbar/cursor-rules/internal/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// quoteYAML double-quotes a string for use in YAML (escapes \ and ").
func quoteYAML(s string) string {
	return `"` + strings.ReplaceAll(strings.ReplaceAll(s, `\`, `\\`), `"`, `\"`) + `"`
}

func TestInstallAllInstallsAllPackages(t *testing.T) {
	packageDir := t.TempDir()
	os.Setenv("CURSOR_RULES_PACKAGE_DIR", packageDir)
	os.Setenv("CURSOR_RULES_CONFIG_DIR", t.TempDir())
	defer func() {
		os.Unsetenv("CURSOR_RULES_PACKAGE_DIR")
		os.Unsetenv("CURSOR_RULES_CONFIG_DIR")
	}()

	// pkgA
	if err := os.MkdirAll(filepath.Join(packageDir, "pkgA"), 0o755); err != nil {
		t.Fatalf("failed to create pkgA: %v", err)
	}
	if err := os.WriteFile(filepath.Join(packageDir, "pkgA", "a1.mdc"), []byte("---\ndescription: a1\n---\nbody"), 0o644); err != nil {
		t.Fatalf("failed to write pkgA/a1.mdc: %v", err)
	}

	// pkgB
	if err := os.MkdirAll(filepath.Join(packageDir, "pkgB"), 0o755); err != nil {
		t.Fatalf("failed to create pkgB: %v", err)
	}
	if err := os.WriteFile(filepath.Join(packageDir, "pkgB", "b1.mdc"), []byte("---\ndescription: b1\n---\nbody"), 0o644); err != nil {
		t.Fatalf("failed to write pkgB/b1.mdc: %v", err)
	}

	// dot directory should be ignored
	if err := os.MkdirAll(filepath.Join(packageDir, ".git"), 0o755); err != nil {
		t.Fatalf("failed to create .git: %v", err)
	}
	if err := os.WriteFile(filepath.Join(packageDir, ".git", "config"), []byte("noop"), 0o644); err != nil {
		t.Fatalf("failed to write .git/config: %v", err)
	}

	workdir := t.TempDir()
	v := viper.New()
	v.Set("workdir", workdir)
	ctx := cli.NewAppContext(v, nil)

	cmd := NewInstallCmd(ctx)
	cmd.SetArgs([]string{"all"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("install all failed: %v", err)
	}

	expectExists(t, filepath.Join(workdir, ".cursor", "rules", "a1.mdc"))
	expectExists(t, filepath.Join(workdir, ".cursor", "rules", "b1.mdc"))

	// ensure no stray files from ignored dirs
	files, err := os.ReadDir(filepath.Join(workdir, ".cursor", "rules"))
	if err != nil {
		t.Fatalf("read rules dir: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected only 2 files installed, got %d", len(files))
	}
}

func TestInstallAllRespectsViperPackageDirWhenEnvUnset(t *testing.T) {
	packageDir := t.TempDir()
	configDir := t.TempDir()
	os.Unsetenv("CURSOR_RULES_PACKAGE_DIR")
	// Config file so LoadConfig uses this packageDir (no sharedDir)
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("packageDir: "+quoteYAML(packageDir)+"\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	os.Setenv("CURSOR_RULES_CONFIG_DIR", configDir)
	defer os.Unsetenv("CURSOR_RULES_CONFIG_DIR")

	// pkgA
	if err := os.MkdirAll(filepath.Join(packageDir, "pkgA"), 0o755); err != nil {
		t.Fatalf("failed to create pkgA: %v", err)
	}
	if err := os.WriteFile(filepath.Join(packageDir, "pkgA", "a1.mdc"), []byte("---\ndescription: a1\n---\nbody"), 0o644); err != nil {
		t.Fatalf("failed to write pkgA/a1.mdc: %v", err)
	}

	workdir := t.TempDir()
	v := viper.New()
	v.Set("workdir", workdir)
	// Simulate config.yaml being loaded into viper by PersistentPreRunE.
	v.Set("packageDir", packageDir)
	ctx := cli.NewAppContext(v, nil)

	cmd := NewInstallCmd(ctx)
	cmd.SetArgs([]string{"all"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("install all failed: %v", err)
	}

	expectExists(t, filepath.Join(workdir, ".cursor", "rules", "a1.mdc"))
}

func TestInstallAllInstallsAllResourceCollections(t *testing.T) {
	packageDir := t.TempDir()
	os.Setenv("CURSOR_RULES_PACKAGE_DIR", packageDir)
	os.Setenv("CURSOR_RULES_CONFIG_DIR", t.TempDir())
	defer func() {
		os.Unsetenv("CURSOR_RULES_PACKAGE_DIR")
		os.Unsetenv("CURSOR_RULES_CONFIG_DIR")
	}()

	if err := os.MkdirAll(filepath.Join(packageDir, "pkgA"), 0o755); err != nil {
		t.Fatalf("failed to create pkgA: %v", err)
	}
	if err := os.WriteFile(filepath.Join(packageDir, "pkgA", "a1.mdc"), []byte("---\ndescription: a1\n---\nbody"), 0o644); err != nil {
		t.Fatalf("failed to write pkgA/a1.mdc: %v", err)
	}

	commandsDir := filepath.Join(packageDir, "commands")
	if err := os.MkdirAll(commandsDir, 0o755); err != nil {
		t.Fatalf("failed to create commands dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(commandsDir, "review.command.mdc"), []byte("---\ndescription: review code changes\n---\n# `/review`\n\nReview the current worktree."), 0o644); err != nil {
		t.Fatalf("failed to write commands/review.command.mdc: %v", err)
	}

	skillDir := filepath.Join(packageDir, "skills", "deploy")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("failed to create skills/deploy: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: deploy\n---\nbody"), 0o644); err != nil {
		t.Fatalf("failed to write skills/deploy/SKILL.md: %v", err)
	}

	agentDir := filepath.Join(packageDir, "agent")
	if err := os.MkdirAll(agentDir, 0o755); err != nil {
		t.Fatalf("failed to create agent dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(agentDir, "reviewer.md"), []byte("# reviewer"), 0o644); err != nil {
		t.Fatalf("failed to write agent/reviewer.md: %v", err)
	}

	hookPresetDir := filepath.Join(packageDir, "hooks", "format")
	if err := os.MkdirAll(filepath.Join(hookPresetDir, "scripts"), 0o755); err != nil {
		t.Fatalf("failed to create hooks/format/scripts: %v", err)
	}
	if err := os.WriteFile(filepath.Join(hookPresetDir, "hooks.json"), []byte(`{"version":1,"hooks":{"afterFileEdit":[{"command":"./scripts/format.sh"}]}}`), 0o644); err != nil {
		t.Fatalf("failed to write hooks.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(hookPresetDir, "scripts", "format.sh"), []byte("#!/usr/bin/env bash\necho ok\n"), 0o755); err != nil {
		t.Fatalf("failed to write hooks script: %v", err)
	}

	workdir := t.TempDir()
	v := viper.New()
	v.Set("workdir", workdir)
	ctx := cli.NewAppContext(v, nil)

	cmd := NewInstallCmd(ctx)
	cmd.SetArgs([]string{"all"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("install all failed: %v", err)
	}

	expectExists(t, filepath.Join(workdir, ".cursor", "rules", "a1.mdc"))
	expectExists(t, filepath.Join(workdir, ".cursor", "skills", "review", "SKILL.md"))
	expectExists(t, filepath.Join(workdir, ".cursor", "skills", "deploy", "SKILL.md"))
	expectExists(t, filepath.Join(workdir, ".cursor", "agents", "reviewer.md"))
	expectExists(t, filepath.Join(workdir, ".cursor", "hooks.json"))
	expectExists(t, filepath.Join(workdir, ".cursor", "hooks", "format.sh"))
	if _, err := os.Stat(filepath.Join(workdir, ".cursor", "commands")); err == nil {
		t.Fatalf("expected commands to be converted to skills, not installed to .cursor/commands")
	}
}

func TestInstallSkillsWithoutNameInstallsAllSkills(t *testing.T) {
	packageDir := t.TempDir()
	os.Setenv("CURSOR_RULES_PACKAGE_DIR", packageDir)
	os.Setenv("CURSOR_RULES_CONFIG_DIR", t.TempDir())
	defer func() {
		os.Unsetenv("CURSOR_RULES_PACKAGE_DIR")
		os.Unsetenv("CURSOR_RULES_CONFIG_DIR")
	}()

	skillDir := filepath.Join(packageDir, "skills", "deploy")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("failed to create skills/deploy: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: deploy\n---\nbody"), 0o644); err != nil {
		t.Fatalf("failed to write skills/deploy/SKILL.md: %v", err)
	}

	workdir := t.TempDir()
	v := viper.New()
	v.Set("workdir", workdir)
	ctx := cli.NewAppContext(v, nil)

	cmd := NewInstallCmd(ctx)
	cmd.SetArgs([]string{"skills"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("install skills failed: %v", err)
	}

	expectExists(t, filepath.Join(workdir, ".cursor", "skills", "deploy", "SKILL.md"))
}

func TestInstallCommandsWithoutNameInstallsAsSkills(t *testing.T) {
	packageDir := t.TempDir()
	os.Setenv("CURSOR_RULES_PACKAGE_DIR", packageDir)
	os.Setenv("CURSOR_RULES_CONFIG_DIR", t.TempDir())
	defer func() {
		os.Unsetenv("CURSOR_RULES_PACKAGE_DIR")
		os.Unsetenv("CURSOR_RULES_CONFIG_DIR")
	}()

	commandsDir := filepath.Join(packageDir, "commands")
	if err := os.MkdirAll(commandsDir, 0o755); err != nil {
		t.Fatalf("failed to create commands dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(commandsDir, "start.command.mdc"), []byte("---\ndescription: start workflow\n---\n# `/start`\n\nBegin the workflow."), 0o644); err != nil {
		t.Fatalf("failed to write command file: %v", err)
	}

	workdir := t.TempDir()
	v := viper.New()
	v.Set("workdir", workdir)
	ctx := cli.NewAppContext(v, nil)

	cmd := NewInstallCmd(ctx)
	cmd.SetArgs([]string{"commands"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("install commands failed: %v", err)
	}

	expectExists(t, filepath.Join(workdir, ".cursor", "skills", "start", "SKILL.md"))
	if _, err := os.Stat(filepath.Join(workdir, ".cursor", "commands")); err == nil {
		t.Fatalf("expected commands to be installed as skills")
	}
}

func TestInstallAllSupportsSkillTarget(t *testing.T) {
	packageDir := t.TempDir()
	os.Setenv("CURSOR_RULES_PACKAGE_DIR", packageDir)
	os.Setenv("CURSOR_RULES_CONFIG_DIR", t.TempDir())
	defer func() {
		os.Unsetenv("CURSOR_RULES_PACKAGE_DIR")
		os.Unsetenv("CURSOR_RULES_CONFIG_DIR")
	}()

	skillDir := filepath.Join(packageDir, "skills", "deploy")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("failed to create skills/deploy: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: deploy\n---\nbody"), 0o644); err != nil {
		t.Fatalf("failed to write skills/deploy/SKILL.md: %v", err)
	}

	workdir := t.TempDir()
	v := viper.New()
	v.Set("workdir", workdir)
	ctx := cli.NewAppContext(v, nil)

	cmd := NewInstallCmd(ctx)
	cmd.SetArgs([]string{"all", "--target", "skills"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("install all --target skills failed: %v", err)
	}

	expectExists(t, filepath.Join(workdir, ".cursor", "skills", "deploy", "SKILL.md"))
}

func TestInstallAllUsesRulesSubdirAsRulesSource(t *testing.T) {
	packageDir := t.TempDir()
	os.Setenv("CURSOR_RULES_PACKAGE_DIR", packageDir)
	os.Setenv("CURSOR_RULES_CONFIG_DIR", t.TempDir())
	defer func() {
		os.Unsetenv("CURSOR_RULES_PACKAGE_DIR")
		os.Unsetenv("CURSOR_RULES_CONFIG_DIR")
	}()

	writeRule := func(rel, body string) {
		t.Helper()
		path := filepath.Join(packageDir, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", path, err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}

	writeRule(filepath.Join("rules", "agent", "base.mdc"), "---\ndescription: base\n---\nbody")
	writeRule(filepath.Join("agent", "AGENT.md"), "# agent should not be treated as rules package")

	workdir := t.TempDir()
	v := viper.New()
	v.Set("workdir", workdir)
	ctx := cli.NewAppContext(v, nil)

	cmd := NewInstallCmd(ctx)
	cmd.SetArgs([]string{"all"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("install all failed: %v", err)
	}

	expectExists(t, filepath.Join(workdir, ".cursor", "rules", "base.mdc"))
	if _, err := os.Stat(filepath.Join(workdir, ".cursor", "rules", "AGENT.md")); err == nil {
		t.Fatalf("expected root agent collection to be ignored by rules install-all")
	}
}

func TestInstallCommandsDefaultsToCursorSkills(t *testing.T) {
	packageDir := t.TempDir()
	os.Setenv("CURSOR_RULES_PACKAGE_DIR", packageDir)
	os.Setenv("CURSOR_RULES_CONFIG_DIR", t.TempDir())
	defer func() {
		os.Unsetenv("CURSOR_RULES_PACKAGE_DIR")
		os.Unsetenv("CURSOR_RULES_CONFIG_DIR")
	}()

	commandsDir := filepath.Join(packageDir, "commands")
	if err := os.MkdirAll(commandsDir, 0o755); err != nil {
		t.Fatalf("failed to create commands dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(commandsDir, "review.command.mdc"), []byte("---\ndescription: review\n---\nReview body"), 0o644); err != nil {
		t.Fatalf("failed to write command file: %v", err)
	}

	workdir := t.TempDir()
	v := viper.New()
	v.Set("workdir", workdir)
	ctx := cli.NewAppContext(v, nil)

	cmd := NewInstallCmd(ctx)
	cmd.SetArgs([]string{"commands", "review"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("install commands failed: %v", err)
	}

	expectExists(t, filepath.Join(workdir, ".cursor", "skills", "review", "SKILL.md"))
	if _, err := os.Stat(filepath.Join(workdir, ".opencode", "commands", "review.md")); err == nil {
		t.Fatalf("expected default command install to avoid OpenCode output")
	}
}

func TestInstallCommandsSupportsOpenCodeTarget(t *testing.T) {
	packageDir := t.TempDir()
	os.Setenv("CURSOR_RULES_PACKAGE_DIR", packageDir)
	os.Setenv("CURSOR_RULES_CONFIG_DIR", t.TempDir())
	defer func() {
		os.Unsetenv("CURSOR_RULES_PACKAGE_DIR")
		os.Unsetenv("CURSOR_RULES_CONFIG_DIR")
	}()

	commandsDir := filepath.Join(packageDir, "commands")
	if err := os.MkdirAll(commandsDir, 0o755); err != nil {
		t.Fatalf("failed to create commands dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(commandsDir, "review.command.mdc"), []byte("---\ndescription: review\n---\nReview body"), 0o644); err != nil {
		t.Fatalf("failed to write command file: %v", err)
	}

	workdir := t.TempDir()
	v := viper.New()
	v.Set("workdir", workdir)
	ctx := cli.NewAppContext(v, nil)

	cmd := NewInstallCmd(ctx)
	cmd.SetArgs([]string{"commands", "review", "--target", "opencode"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("install commands --target opencode failed: %v", err)
	}

	path := filepath.Join(workdir, ".opencode", "commands", "review.md")
	expectExists(t, path)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read opencode command: %v", err)
	}
	if strings.Contains(string(data), "description:") {
		t.Fatalf("expected opencode command to omit mdc frontmatter, got %q", string(data))
	}
	if _, err := os.Stat(filepath.Join(workdir, ".cursor", "skills", "review", "SKILL.md")); err == nil {
		t.Fatalf("expected opencode command install to avoid Cursor skill output")
	}
}

func TestInstallSkillsSupportsOpenCodeTarget(t *testing.T) {
	packageDir := t.TempDir()
	os.Setenv("CURSOR_RULES_PACKAGE_DIR", packageDir)
	os.Setenv("CURSOR_RULES_CONFIG_DIR", t.TempDir())
	defer func() {
		os.Unsetenv("CURSOR_RULES_PACKAGE_DIR")
		os.Unsetenv("CURSOR_RULES_CONFIG_DIR")
	}()

	skillDir := filepath.Join(packageDir, "skills", "deploy")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("failed to create skill dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: deploy\n---\nbody"), 0o644); err != nil {
		t.Fatalf("failed to write skill file: %v", err)
	}

	workdir := t.TempDir()
	v := viper.New()
	v.Set("workdir", workdir)
	ctx := cli.NewAppContext(v, nil)

	cmd := NewInstallCmd(ctx)
	cmd.SetArgs([]string{"skills", "deploy", "--target", "opencode"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("install skills --target opencode failed: %v", err)
	}

	expectExists(t, filepath.Join(workdir, ".opencode", "skills", "deploy", "SKILL.md"))
	if _, err := os.Stat(filepath.Join(workdir, ".cursor", "skills", "deploy", "SKILL.md")); err == nil {
		t.Fatalf("expected opencode skill install to avoid Cursor output")
	}
}

func TestInstallAgentsSupportsOpenCodeTarget(t *testing.T) {
	packageDir := t.TempDir()
	os.Setenv("CURSOR_RULES_PACKAGE_DIR", packageDir)
	os.Setenv("CURSOR_RULES_CONFIG_DIR", t.TempDir())
	defer func() {
		os.Unsetenv("CURSOR_RULES_PACKAGE_DIR")
		os.Unsetenv("CURSOR_RULES_CONFIG_DIR")
	}()

	agentDir := filepath.Join(packageDir, "agents")
	if err := os.MkdirAll(agentDir, 0o755); err != nil {
		t.Fatalf("failed to create agents dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(agentDir, "reviewer.md"), []byte("# reviewer"), 0o644); err != nil {
		t.Fatalf("failed to write agent file: %v", err)
	}

	workdir := t.TempDir()
	v := viper.New()
	v.Set("workdir", workdir)
	ctx := cli.NewAppContext(v, nil)

	cmd := NewInstallCmd(ctx)
	cmd.SetArgs([]string{"agents", "reviewer", "--target", "opencode"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("install agents --target opencode failed: %v", err)
	}

	expectExists(t, filepath.Join(workdir, ".opencode", "agents", "reviewer.md"))
	if _, err := os.Stat(filepath.Join(workdir, ".cursor", "agents", "reviewer.md")); err == nil {
		t.Fatalf("expected opencode agent install to avoid Cursor output")
	}
}

func TestInstallNativeResourcesGlobalOpenCodeUsesConfigDir(t *testing.T) {
	packageDir := t.TempDir()
	configDir := t.TempDir()
	openCodeDir := t.TempDir()
	os.Setenv("CURSOR_RULES_PACKAGE_DIR", packageDir)
	os.Setenv("CURSOR_RULES_CONFIG_DIR", configDir)
	os.Setenv("OPENCODE_CONFIG_DIR", openCodeDir)
	defer func() {
		os.Unsetenv("CURSOR_RULES_PACKAGE_DIR")
		os.Unsetenv("CURSOR_RULES_CONFIG_DIR")
		os.Unsetenv("OPENCODE_CONFIG_DIR")
	}()

	if err := os.MkdirAll(filepath.Join(packageDir, "commands"), 0o755); err != nil {
		t.Fatalf("create commands dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(packageDir, "commands", "review.command.mdc"), []byte("---\ndescription: review\n---\nReview body"), 0o644); err != nil {
		t.Fatalf("write command: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(packageDir, "skills", "deploy"), 0o755); err != nil {
		t.Fatalf("create skill dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(packageDir, "skills", "deploy", "SKILL.md"), []byte("---\nname: deploy\n---\nbody"), 0o644); err != nil {
		t.Fatalf("write skill: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(packageDir, "agents"), 0o755); err != nil {
		t.Fatalf("create agents dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(packageDir, "agents", "reviewer.md"), []byte("# reviewer"), 0o644); err != nil {
		t.Fatalf("write agent: %v", err)
	}

	workdir := t.TempDir()
	v := viper.New()
	v.Set("workdir", workdir)
	ctx := cli.NewAppContext(v, nil)

	newRoot := func() *cobra.Command {
		root := cli.NewRoot(ctx, cli.Palette{NewInstallCmd})
		root.PersistentFlags().StringP("workdir", "w", "", "workspace root (defaults to current directory)")
		root.PersistentFlags().String("dir", "", "destination: path or 'user' (shorthand: -w/--workdir for path, --global for user)")
		root.PersistentFlags().Bool("global", false, "use user dirs (~/.cursor/...) as destination (same as --dir user)")
		return root
	}

	root1 := newRoot()
	root1.SetArgs([]string{"install", "commands", "review", "--target", "opencode", "--global"})
	cmd1 := root1
	if err := cmd1.Execute(); err != nil {
		t.Fatalf("global opencode command install failed: %v", err)
	}
	root2 := newRoot()
	root2.SetArgs([]string{"install", "skills", "deploy", "--target", "opencode", "--global"})
	cmd2 := root2
	if err := cmd2.Execute(); err != nil {
		t.Fatalf("global opencode skill install failed: %v", err)
	}
	root3 := newRoot()
	root3.SetArgs([]string{"install", "agents", "reviewer", "--target", "opencode", "--global"})
	cmd3 := root3
	if err := cmd3.Execute(); err != nil {
		t.Fatalf("global opencode agent install failed: %v", err)
	}

	expectExists(t, filepath.Join(openCodeDir, "commands", "review.md"))
	expectExists(t, filepath.Join(openCodeDir, "skills", "deploy", "SKILL.md"))
	expectExists(t, filepath.Join(openCodeDir, "agents", "reviewer.md"))
}

func expectExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file to exist: %s (%v)", path, err)
	}
}
