package app

import (
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/ZanzyTHEbar/cursor-rules/internal/config"
	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
	"github.com/ZanzyTHEbar/cursor-rules/internal/errors"
	"github.com/ZanzyTHEbar/cursor-rules/internal/security"
)

const (
	resourceKindCommand = "command"
	resourceKindSkill   = "skill"
	resourceKindAgent   = "agent"
	resourceKindHooks   = "hooks"
	resourceKindRule    = "rule"
)

type nativeResourceInstallOptions struct {
	Excludes  []string
	NoFlatten bool
	IsUser    bool // when true, use UserCursor* dirs (CURSOR_USER_DIR, per-feature overrides)
}

type nativeResourceInstallAllPlan struct {
	Name  string
	Label string
}

type nativeResourceProvider interface {
	Kind() string
	Target() string
	OutputDir() string
	RequiresName() bool
	ListAvailable(packageDir string, cfg *config.Config) ([]string, error)
	ListInstalled(projectRoot string, cfg *config.Config, isUser bool) ([]string, error)
	Install(projectRoot, packageDir, name string, cfg *config.Config, opts nativeResourceInstallOptions) (core.InstallStrategy, error)
	PlanInstallAll(packageDir string, cfg *config.Config) ([]nativeResourceInstallAllPlan, error)
	IncludeInDefaultInstallAll() bool
	DetectDefaultTarget(packageDir, name string, cfg *config.Config) (string, bool, error)
	Remove(projectRoot, name string, cfg *config.Config, isUser bool) (bool, error)
}

type nativeResourceRegistry struct {
	ordered  []nativeResourceProvider
	byTarget map[string]nativeResourceProvider
	byKind   map[string]nativeResourceProvider
}

func newNativeResourceRegistry(transformerProvider TransformerProvider) *nativeResourceRegistry {
	providers := []nativeResourceProvider{
		commandResourceProvider{},
		skillResourceProvider{},
		agentResourceProvider{},
		hooksResourceProvider{},
	}
	if transformerProvider != nil {
		providers = append(providers,
			rulesResourceProvider{target: "cursor", tp: transformerProvider},
			rulesResourceProvider{target: "copilot-instr", tp: transformerProvider},
			rulesResourceProvider{target: "copilot-prompt", tp: transformerProvider},
		)
	}
	registry := &nativeResourceRegistry{
		ordered:  providers,
		byTarget: make(map[string]nativeResourceProvider, len(providers)),
		byKind:   make(map[string]nativeResourceProvider, len(providers)),
	}
	for _, provider := range providers {
		registry.byTarget[provider.Target()] = provider
		if provider.Kind() != resourceKindRule {
			registry.byKind[provider.Kind()] = provider
		}
	}
	return registry
}

func (r *nativeResourceRegistry) providerForTarget(target string) (nativeResourceProvider, bool) {
	if r == nil {
		return nil, false
	}
	provider, ok := r.byTarget[strings.TrimSpace(target)]
	return provider, ok
}

func (r *nativeResourceRegistry) providerForKind(kind string) (nativeResourceProvider, bool) {
	if r == nil {
		return nil, false
	}
	provider, ok := r.byKind[strings.TrimSpace(kind)]
	return provider, ok
}

func (r *nativeResourceRegistry) providers() []nativeResourceProvider {
	if r == nil {
		return nil
	}
	return r.ordered
}

func (r *nativeResourceRegistry) resolveDefaultTarget(packageDir, name string, cfg *config.Config) (target string, ok bool, err error) {
	var lastErr error
	for _, provider := range r.providers() {
		target, ok, err = provider.DetectDefaultTarget(packageDir, name, cfg)
		if ok {
			return target, true, nil
		}
		if err != nil {
			lastErr = err
		}
	}
	return "", false, lastErr
}

func removeFromInstalledList(name string, installed []string, remove func() error) (bool, error) {
	if !containsInstalledResource(installed, name) {
		return false, nil
	}
	return true, remove()
}

func containsInstalledResource(installed []string, name string) bool {
	if slices.Contains(installed, name) {
		return true
	}
	if slices.Contains(installed, name+".md") {
		return true
	}
	return false
}

func installAllPlansToEntries(provider nativeResourceProvider, plans []nativeResourceInstallAllPlan) []installAllEntry {
	entries := make([]installAllEntry, 0, len(plans))
	for _, plan := range plans {
		entries = append(entries, installAllEntry{
			Name:   plan.Name,
			Target: provider.Target(),
			Label:  plan.Label,
		})
	}
	return entries
}

type commandResourceProvider struct{}

func (commandResourceProvider) Kind() string      { return resourceKindCommand }
func (commandResourceProvider) Target() string    { return "commands" }
func (commandResourceProvider) OutputDir() string { return ".cursor/skills" }
func (commandResourceProvider) RequiresName() bool {
	return true
}
func (commandResourceProvider) ListAvailable(packageDir string, _ *config.Config) ([]string, error) {
	return core.ListCursorCompatibleCommands(packageDir)
}
func (commandResourceProvider) ListInstalled(projectRoot string, cfg *config.Config, isUser bool) ([]string, error) {
	legacyCommandsDir := config.EffectiveCommandsDir(projectRoot, isUser, cfg)
	skillsDir := config.EffectiveSkillsDir(projectRoot, isUser, cfg)
	compat, err := core.ListInstalledCommandSkills(skillsDir)
	if err != nil {
		return nil, err
	}
	legacy, err := core.ListInstalledCommands(legacyCommandsDir)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{}, len(compat)+len(legacy))
	out := make([]string, 0, len(compat)+len(legacy))
	for _, name := range compat {
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	for _, name := range legacy {
		normalized := strings.TrimSuffix(name, ".md")
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	sort.Strings(out)
	return out, nil
}
func (commandResourceProvider) Install(projectRoot, packageDir, name string, cfg *config.Config, opts nativeResourceInstallOptions) (core.InstallStrategy, error) {
	skillsDir := config.EffectiveSkillsDir(projectRoot, opts.IsUser, cfg)
	if strings.TrimSpace(name) == core.CommandsSubdir() {
		return core.InstallCommandCollectionAsSkillsToDir(skillsDir, packageDir, opts.Excludes)
	}
	return core.InstallCommandAsSkillToDir(skillsDir, packageDir, name, opts.Excludes)
}
func (commandResourceProvider) PlanInstallAll(packageDir string, _ *config.Config) ([]nativeResourceInstallAllPlan, error) {
	names, err := core.ListCursorCompatibleCommands(packageDir)
	if err != nil {
		return nil, err
	}
	plans := make([]nativeResourceInstallAllPlan, 0, len(names))
	for _, name := range names {
		plans = append(plans, nativeResourceInstallAllPlan{
			Name:  name,
			Label: filepath.Join(core.CommandsSubdir(), name),
		})
	}
	return plans, nil
}
func (commandResourceProvider) IncludeInDefaultInstallAll() bool { return true }
func (commandResourceProvider) DetectDefaultTarget(packageDir, name string, _ *config.Config) (target string, ok bool, err error) {
	if strings.TrimSpace(name) != core.CommandsSubdir() {
		names, err := core.ListCursorCompatibleCommands(packageDir)
		if err != nil {
			return "", false, err
		}
		for _, candidate := range names {
			if candidate == strings.TrimSpace(name) {
				return "commands", true, nil
			}
		}
		return "", false, nil
	}
	names, err := core.ListCursorCompatibleCommands(packageDir)
	if err != nil {
		return "", false, err
	}
	return "commands", len(names) > 0, nil
}
func (commandResourceProvider) Remove(projectRoot, name string, cfg *config.Config, isUser bool) (bool, error) {
	installed, err := commandResourceProvider{}.ListInstalled(projectRoot, cfg, isUser)
	if err != nil {
		return false, err
	}
	return removeFromInstalledList(name, installed, func() error {
		skillsDir := config.EffectiveSkillsDir(projectRoot, isUser, cfg)
		commandsDir := config.EffectiveCommandsDir(projectRoot, isUser, cfg)
		if err := core.RemoveSkill(skillsDir, name); err != nil {
			return err
		}
		return core.RemoveCommand(commandsDir, name)
	})
}

type skillResourceProvider struct{}

func (skillResourceProvider) Kind() string      { return resourceKindSkill }
func (skillResourceProvider) Target() string    { return "skills" }
func (skillResourceProvider) OutputDir() string { return ".cursor/skills" }
func (skillResourceProvider) RequiresName() bool {
	return true
}
func (skillResourceProvider) ListAvailable(packageDir string, cfg *config.Config) ([]string, error) {
	return core.ListSkillDirs(packageDir, cfg.SkillsSubdir)
}
func (skillResourceProvider) ListInstalled(projectRoot string, cfg *config.Config, isUser bool) ([]string, error) {
	skillsDir := config.EffectiveSkillsDir(projectRoot, isUser, cfg)
	return core.ListSkillDirsFrom(skillsDir)
}
func (skillResourceProvider) Install(projectRoot, packageDir, name string, cfg *config.Config, opts nativeResourceInstallOptions) (core.InstallStrategy, error) {
	skillsDir := config.EffectiveSkillsDir(projectRoot, opts.IsUser, cfg)
	if strings.TrimSpace(name) == core.SkillsSubdir(cfg.SkillsSubdir) {
		return installAllFromProviderWithDir(skillResourceProvider{}, projectRoot, packageDir, cfg, opts.IsUser)
	}
	return core.InstallSkillToDir(skillsDir, packageDir, name, cfg.SkillsSubdir)
}
func (skillResourceProvider) PlanInstallAll(packageDir string, cfg *config.Config) ([]nativeResourceInstallAllPlan, error) {
	names, err := core.ListSkillDirs(packageDir, cfg.SkillsSubdir)
	if err != nil {
		return nil, err
	}
	plans := make([]nativeResourceInstallAllPlan, 0, len(names))
	for _, name := range names {
		plans = append(plans, nativeResourceInstallAllPlan{
			Name:  name,
			Label: filepath.Join(core.SkillsSubdir(cfg.SkillsSubdir), name),
		})
	}
	return plans, nil
}
func (skillResourceProvider) IncludeInDefaultInstallAll() bool { return true }
func (skillResourceProvider) DetectDefaultTarget(packageDir, name string, cfg *config.Config) (target string, ok bool, err error) {
	if strings.TrimSpace(name) != core.SkillsSubdir(cfg.SkillsSubdir) {
		return "", false, nil
	}
	names, err := core.ListSkillDirs(packageDir, cfg.SkillsSubdir)
	if err != nil {
		return "", false, err
	}
	return "skills", len(names) > 0, nil
}
func (skillResourceProvider) Remove(projectRoot, name string, cfg *config.Config, isUser bool) (bool, error) {
	installed, err := skillResourceProvider{}.ListInstalled(projectRoot, cfg, isUser)
	if err != nil {
		return false, err
	}
	return removeFromInstalledList(name, installed, func() error {
		return core.RemoveSkill(config.EffectiveSkillsDir(projectRoot, isUser, cfg), name)
	})
}

type agentResourceProvider struct{}

func (agentResourceProvider) Kind() string      { return resourceKindAgent }
func (agentResourceProvider) Target() string    { return "agents" }
func (agentResourceProvider) OutputDir() string { return ".cursor/agents" }
func (agentResourceProvider) RequiresName() bool {
	return true
}
func (agentResourceProvider) ListAvailable(packageDir string, cfg *config.Config) ([]string, error) {
	return core.ListAgentFiles(packageDir, cfg.AgentsSubdir)
}
func (agentResourceProvider) ListInstalled(projectRoot string, cfg *config.Config, isUser bool) ([]string, error) {
	agentsDir := config.EffectiveAgentsDir(projectRoot, isUser, cfg)
	return core.ListAgentFilesFrom(agentsDir)
}
func (agentResourceProvider) Install(projectRoot, packageDir, name string, cfg *config.Config, opts nativeResourceInstallOptions) (core.InstallStrategy, error) {
	agentsDir := config.EffectiveAgentsDir(projectRoot, opts.IsUser, cfg)
	if strings.TrimSpace(name) == core.AgentsSubdir(cfg.AgentsSubdir) {
		return installAllFromProviderWithDir(agentResourceProvider{}, projectRoot, packageDir, cfg, opts.IsUser)
	}
	subdir := core.ResolveAgentsSubdir(packageDir, cfg.AgentsSubdir)
	agentsRoot, err := security.SafeJoin(packageDir, subdir)
	if err != nil {
		return core.StrategyUnknown, err
	}
	return core.InstallAgentToDir(agentsDir, agentsRoot, name, ".md")
}
func (agentResourceProvider) PlanInstallAll(packageDir string, cfg *config.Config) ([]nativeResourceInstallAllPlan, error) {
	names, err := core.ListAgentFiles(packageDir, cfg.AgentsSubdir)
	if err != nil {
		return nil, err
	}
	plans := make([]nativeResourceInstallAllPlan, 0, len(names))
	for _, name := range names {
		plans = append(plans, nativeResourceInstallAllPlan{
			Name:  name,
			Label: filepath.Join(core.AgentsSubdir(cfg.AgentsSubdir), name),
		})
	}
	return plans, nil
}
func (agentResourceProvider) IncludeInDefaultInstallAll() bool { return true }
func (agentResourceProvider) DetectDefaultTarget(packageDir, name string, cfg *config.Config) (target string, ok bool, err error) {
	if strings.TrimSpace(name) != core.AgentsSubdir(cfg.AgentsSubdir) {
		return "", false, nil
	}
	names, err := core.ListAgentFiles(packageDir, cfg.AgentsSubdir)
	if err != nil {
		return "", false, err
	}
	return "agents", len(names) > 0, nil
}
func (agentResourceProvider) Remove(projectRoot, name string, cfg *config.Config, isUser bool) (bool, error) {
	installed, err := agentResourceProvider{}.ListInstalled(projectRoot, cfg, isUser)
	if err != nil {
		return false, err
	}
	return removeFromInstalledList(name, installed, func() error {
		return core.RemoveAgent(config.EffectiveAgentsDir(projectRoot, isUser, cfg), name)
	})
}

type hooksResourceProvider struct{}

func (hooksResourceProvider) Kind() string      { return resourceKindHooks }
func (hooksResourceProvider) Target() string    { return "hooks" }
func (hooksResourceProvider) OutputDir() string { return ".cursor/hooks" }
func (hooksResourceProvider) RequiresName() bool {
	return false
}
func (hooksResourceProvider) ListAvailable(packageDir string, cfg *config.Config) ([]string, error) {
	return core.ListHookPresets(packageDir, cfg.HooksSubdir)
}
func (hooksResourceProvider) ListInstalled(projectRoot string, cfg *config.Config, isUser bool) ([]string, error) {
	hooksDir := config.EffectiveHooksDir(projectRoot, isUser, cfg)
	jsonPath := config.EffectiveHooksJSON(projectRoot, isUser, cfg)
	return core.ListInstalledHooksFrom(hooksDir, jsonPath)
}
func (hooksResourceProvider) Install(projectRoot, packageDir, name string, cfg *config.Config, opts nativeResourceInstallOptions) (core.InstallStrategy, error) {
	hooksDir := config.EffectiveHooksDir(projectRoot, opts.IsUser, cfg)
	jsonPath := config.EffectiveHooksJSON(projectRoot, opts.IsUser, cfg)
	return core.InstallHookPresetToDirs(hooksDir, jsonPath, packageDir, name, cfg.HooksSubdir)
}
func (hooksResourceProvider) PlanInstallAll(packageDir string, cfg *config.Config) ([]nativeResourceInstallAllPlan, error) {
	names, err := core.ListHookPresets(packageDir, cfg.HooksSubdir)
	if err != nil {
		return nil, err
	}
	plans := make([]nativeResourceInstallAllPlan, 0, len(names))
	for _, name := range names {
		plans = append(plans, nativeResourceInstallAllPlan{
			Name:  name,
			Label: filepath.Join(core.HooksSubdir(cfg.HooksSubdir), name),
		})
	}
	return plans, nil
}
func (hooksResourceProvider) IncludeInDefaultInstallAll() bool { return true }
func (hooksResourceProvider) DetectDefaultTarget(_, _ string, _ *config.Config) (target string, ok bool, err error) {
	return "", false, nil
}
func (hooksResourceProvider) Remove(projectRoot, _ string, cfg *config.Config, isUser bool) (bool, error) {
	installed, err := hooksResourceProvider{}.ListInstalled(projectRoot, cfg, isUser)
	if err != nil {
		return false, err
	}
	if len(installed) == 0 {
		return false, nil
	}
	return true, core.RemoveHookPresetFromDirs(config.EffectiveHooksDir(projectRoot, isUser, cfg), config.EffectiveHooksJSON(projectRoot, isUser, cfg))
}

type rulesResourceProvider struct {
	target string
	tp     TransformerProvider
}

func (p rulesResourceProvider) Kind() string   { return resourceKindRule }
func (p rulesResourceProvider) Target() string { return p.target }
func (p rulesResourceProvider) OutputDir() string {
	switch p.target {
	case "cursor":
		return ".cursor/rules"
	case "copilot-instr":
		return ".github/instructions"
	case "copilot-prompt":
		return ".github/prompts"
	default:
		return ".cursor/rules"
	}
}
func (p rulesResourceProvider) RequiresName() bool { return true }

func (p rulesResourceProvider) ListAvailable(packageDir string, _ *config.Config) ([]string, error) {
	presets, err := core.ListPackagePresets(packageDir)
	if err != nil {
		return nil, err
	}
	pkgs, err := core.ListPackageDirs(packageDir)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool)
	var out []string
	for _, f := range presets {
		name := strings.TrimSuffix(f, ".mdc")
		if name != "" && !seen[name] {
			seen[name] = true
			out = append(out, name)
		}
	}
	for _, d := range pkgs {
		if !seen[d] {
			seen[d] = true
			out = append(out, d)
		}
	}
	return out, nil
}

func (p rulesResourceProvider) ListInstalled(projectRoot string, cfg *config.Config, isUser bool) ([]string, error) {
	if p.target != "cursor" {
		return nil, nil
	}
	rulesDir := config.EffectiveRulesDir(projectRoot, isUser, cfg)
	presets, err := core.ListProjectPresetsFrom(rulesDir)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(presets))
	for _, f := range presets {
		name := strings.TrimSuffix(f, ".mdc")
		if name != "" {
			out = append(out, name)
		}
	}
	return out, nil
}

func (p rulesResourceProvider) Install(projectRoot, packageDir, name string, cfg *config.Config, opts nativeResourceInstallOptions) (core.InstallStrategy, error) {
	rulesPackageDir := core.ResolveRulesPackageDir(packageDir)
	trans, err := p.tp.Transformer(p.target)
	if err != nil {
		return core.StrategyUnknown, err
	}
	pkgPath := filepath.Join(rulesPackageDir, name)
	info, statErr := os.Stat(pkgPath)
	isPackage := statErr == nil && info.IsDir()

	if p.target == "cursor" && opts.IsUser {
		rulesDir := config.EffectiveRulesDir(projectRoot, true, cfg)
		if isPackage {
			if core.UseSymlink() || core.WantGNUStow() {
				return core.InstallPackageToRulesDir(rulesDir, rulesPackageDir, name, opts.Excludes, opts.NoFlatten)
			}
			return installPackageWithTransformerToRulesDir(rulesDir, pkgPath, name, trans, opts.Excludes, opts.NoFlatten)
		}
		presetPath := filepath.Join(rulesPackageDir, name)
		if !strings.HasSuffix(presetPath, ".mdc") {
			presetPath += ".mdc"
		}
		return installPresetWithTransformerToRulesDir(rulesDir, presetPath, name, trans, rulesPackageDir)
	}

	if isPackage {
		if trans.Target() == "cursor" && (core.UseSymlink() || core.WantGNUStow()) {
			return core.InstallPackageFromPackageDir(projectRoot, rulesPackageDir, name, opts.Excludes, opts.NoFlatten)
		}
		return installPackageWithTransformer(projectRoot, pkgPath, name, trans, opts.Excludes, opts.NoFlatten)
	}
	presetPath := filepath.Join(rulesPackageDir, name)
	if !strings.HasSuffix(presetPath, ".mdc") {
		presetPath += ".mdc"
	}
	return installPresetWithTransformer(projectRoot, presetPath, name, trans, rulesPackageDir)
}

func (p rulesResourceProvider) PlanInstallAll(packageDir string, _ *config.Config) ([]nativeResourceInstallAllPlan, error) {
	pkgs, err := core.ListPackageDirs(packageDir)
	if err != nil {
		return nil, err
	}
	presets, err := core.ListPackagePresets(packageDir)
	if err != nil {
		return nil, err
	}
	plans := make([]nativeResourceInstallAllPlan, 0, len(pkgs)+len(presets))
	for _, d := range pkgs {
		plans = append(plans, nativeResourceInstallAllPlan{Name: d, Label: d})
	}
	for _, f := range presets {
		name := strings.TrimSuffix(f, ".mdc")
		if name != "" {
			plans = append(plans, nativeResourceInstallAllPlan{Name: name, Label: name})
		}
	}
	return plans, nil
}

func (p rulesResourceProvider) IncludeInDefaultInstallAll() bool {
	return p.target == "cursor"
}

func (p rulesResourceProvider) DetectDefaultTarget(packageDir, name string, _ *config.Config) (target string, ok bool, err error) {
	presets, err := core.ListPackagePresets(packageDir)
	if err != nil {
		return "", false, err
	}
	for _, f := range presets {
		if strings.TrimSuffix(f, ".mdc") == name {
			return p.target, true, nil
		}
	}
	pkgs, err := core.ListPackageDirs(packageDir)
	if err != nil {
		return "", false, err
	}
	for _, d := range pkgs {
		if d == name {
			return p.target, true, nil
		}
	}
	return "", false, nil
}

func (p rulesResourceProvider) Remove(projectRoot, name string, cfg *config.Config, isUser bool) (bool, error) {
	if p.target != "cursor" {
		return false, nil
	}
	installed, err := p.ListInstalled(projectRoot, cfg, isUser)
	if err != nil {
		return false, err
	}
	return removeFromInstalledList(name, installed, func() error {
		return core.RemovePreset(config.EffectiveRulesDir(projectRoot, isUser, cfg), name)
	})
}

func installAllFromProviderWithDir(provider nativeResourceProvider, projectRoot, packageDir string, cfg *config.Config, isUser bool) (core.InstallStrategy, error) {
	plans, err := provider.PlanInstallAll(packageDir, cfg)
	if err != nil {
		return core.StrategyUnknown, err
	}
	if len(plans) == 0 {
		return core.StrategyUnknown, errors.Newf(errors.CodeNotFound, "%s collection not found", provider.Kind())
	}

	usedStrategy := core.StrategyCopy
	for _, plan := range plans {
		strategy, err := provider.Install(projectRoot, packageDir, plan.Name, cfg, nativeResourceInstallOptions{IsUser: isUser})
		if err != nil {
			return core.StrategyUnknown, err
		}
		if strategy != core.StrategyCopy {
			usedStrategy = strategy
		}
	}
	return usedStrategy, nil
}
