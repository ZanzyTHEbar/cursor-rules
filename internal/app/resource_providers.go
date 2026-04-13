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
	OutputDir(projectRoot string, cfg *config.Config, isUser bool) string
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
	ordered     []nativeResourceProvider
	kindOrdered []string
	byTarget    map[string]nativeResourceProvider
	byKind      map[string][]nativeResourceProvider
}

func newNativeResourceRegistry(transformerProvider TransformerProvider) *nativeResourceRegistry {
	providers := []nativeResourceProvider{
		commandResourceProvider{target: "commands"},
		commandResourceProvider{target: "opencode-commands", opencode: true},
		skillResourceProvider{target: "skills"},
		skillResourceProvider{target: "opencode-skills", opencode: true},
		agentResourceProvider{target: "agents"},
		agentResourceProvider{target: "opencode-agents", opencode: true},
		hooksResourceProvider{},
	}
	if transformerProvider != nil {
		for _, target := range orderedRuleTargets(transformerProvider.AvailableTargets()) {
			providers = append(providers, rulesResourceProvider{target: target, tp: transformerProvider})
		}
	}
	registry := &nativeResourceRegistry{
		ordered:  providers,
		byTarget: make(map[string]nativeResourceProvider, len(providers)),
		byKind:   make(map[string][]nativeResourceProvider, len(providers)),
	}
	for _, provider := range providers {
		registry.byTarget[provider.Target()] = provider
		kind := provider.Kind()
		if _, exists := registry.byKind[kind]; !exists {
			registry.kindOrdered = append(registry.kindOrdered, kind)
		}
		registry.byKind[kind] = append(registry.byKind[kind], provider)
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

func (r *nativeResourceRegistry) providersForKind(kind string) []nativeResourceProvider {
	if r == nil {
		return nil
	}
	providers, ok := r.byKind[strings.TrimSpace(kind)]
	if !ok {
		return nil
	}
	return slices.Clone(providers)
}

func (r *nativeResourceRegistry) providers() []nativeResourceProvider {
	if r == nil {
		return nil
	}
	return r.ordered
}

func (r *nativeResourceRegistry) uniqueKindProviders() []nativeResourceProvider {
	if r == nil {
		return nil
	}
	providers := make([]nativeResourceProvider, 0, len(r.kindOrdered))
	for _, kind := range r.kindOrdered {
		kindProviders := r.byKind[kind]
		if len(kindProviders) == 0 {
			continue
		}
		providers = append(providers, kindProviders[0])
	}
	return providers
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

func orderedRuleTargets(targets []string) []string {
	seen := make(map[string]struct{}, len(targets))
	for _, target := range targets {
		target = strings.TrimSpace(target)
		if target == "" {
			continue
		}
		seen[target] = struct{}{}
	}

	ordered := make([]string, 0, len(seen))
	for _, target := range []string{"cursor", "copilot-instr", "copilot-prompt", "opencode-rules"} {
		if _, ok := seen[target]; !ok {
			continue
		}
		ordered = append(ordered, target)
		delete(seen, target)
	}

	extra := make([]string, 0, len(seen))
	for target := range seen {
		extra = append(extra, target)
	}
	sort.Strings(extra)
	return append(ordered, extra...)
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

type commandResourceProvider struct {
	target   string
	opencode bool
}

func (p commandResourceProvider) Kind() string   { return resourceKindCommand }
func (p commandResourceProvider) Target() string { return p.target }
func (p commandResourceProvider) OutputDir(projectRoot string, cfg *config.Config, isUser bool) string {
	if p.opencode {
		return config.EffectiveOpenCodeCommandsDir(projectRoot, isUser)
	}
	return config.EffectiveSkillsDir(projectRoot, isUser, cfg)
}
func (commandResourceProvider) RequiresName() bool {
	return true
}
func (commandResourceProvider) ListAvailable(packageDir string, _ *config.Config) ([]string, error) {
	return core.ListCursorCompatibleCommands(packageDir)
}

func (p commandResourceProvider) ListInstalled(projectRoot string, cfg *config.Config, isUser bool) ([]string, error) {
	if p.opencode {
		commandsDir := config.EffectiveOpenCodeCommandsDir(projectRoot, isUser)
		return core.ListInstalledCommands(commandsDir)
	}
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

func (p commandResourceProvider) Install(projectRoot, packageDir, name string, cfg *config.Config, opts nativeResourceInstallOptions) (core.InstallStrategy, error) {
	if p.opencode {
		commandsDir := config.EffectiveOpenCodeCommandsDir(projectRoot, opts.IsUser)
		if strings.TrimSpace(name) == core.CommandsSubdir() {
			return core.InstallOpenCodeCommandCollectionToDir(commandsDir, packageDir, opts.Excludes)
		}
		return core.InstallOpenCodeCommandToDir(commandsDir, packageDir, name, opts.Excludes)
	}
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
func (p commandResourceProvider) IncludeInDefaultInstallAll() bool { return !p.opencode }
func (p commandResourceProvider) DetectDefaultTarget(packageDir, name string, _ *config.Config) (target string, ok bool, err error) {
	if p.opencode {
		return "", false, nil
	}
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
func (p commandResourceProvider) Remove(projectRoot, name string, cfg *config.Config, isUser bool) (bool, error) {
	installed, err := p.ListInstalled(projectRoot, cfg, isUser)
	if err != nil {
		return false, err
	}
	return removeFromInstalledList(name, installed, func() error {
		if p.opencode {
			return core.RemoveCommand(config.EffectiveOpenCodeCommandsDir(projectRoot, isUser), name)
		}
		skillsDir := config.EffectiveSkillsDir(projectRoot, isUser, cfg)
		commandsDir := config.EffectiveCommandsDir(projectRoot, isUser, cfg)
		if err := core.RemoveSkill(skillsDir, name); err != nil {
			return err
		}
		return core.RemoveCommand(commandsDir, name)
	})
}

type skillResourceProvider struct {
	target   string
	opencode bool
}

func (p skillResourceProvider) Kind() string   { return resourceKindSkill }
func (p skillResourceProvider) Target() string { return p.target }
func (p skillResourceProvider) OutputDir(projectRoot string, cfg *config.Config, isUser bool) string {
	if p.opencode {
		return config.EffectiveOpenCodeSkillsDir(projectRoot, isUser)
	}
	return config.EffectiveSkillsDir(projectRoot, isUser, cfg)
}
func (skillResourceProvider) RequiresName() bool {
	return true
}
func (skillResourceProvider) ListAvailable(packageDir string, cfg *config.Config) ([]string, error) {
	return core.ListSkillDirs(packageDir, cfg.SkillsSubdir)
}
func (p skillResourceProvider) ListInstalled(projectRoot string, cfg *config.Config, isUser bool) ([]string, error) {
	skillsDir := config.EffectiveSkillsDir(projectRoot, isUser, cfg)
	if p.opencode {
		skillsDir = config.EffectiveOpenCodeSkillsDir(projectRoot, isUser)
	}
	return core.ListSkillDirsFrom(skillsDir)
}
func (p skillResourceProvider) Install(projectRoot, packageDir, name string, cfg *config.Config, opts nativeResourceInstallOptions) (core.InstallStrategy, error) {
	skillsDir := config.EffectiveSkillsDir(projectRoot, opts.IsUser, cfg)
	if p.opencode {
		skillsDir = config.EffectiveOpenCodeSkillsDir(projectRoot, opts.IsUser)
	}
	if strings.TrimSpace(name) == core.SkillsSubdir(cfg.SkillsSubdir) {
		return installAllFromProviderWithDir(p, projectRoot, packageDir, cfg, opts.IsUser)
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
func (p skillResourceProvider) IncludeInDefaultInstallAll() bool { return !p.opencode }
func (p skillResourceProvider) DetectDefaultTarget(packageDir, name string, cfg *config.Config) (target string, ok bool, err error) {
	if p.opencode {
		return "", false, nil
	}
	if strings.TrimSpace(name) != core.SkillsSubdir(cfg.SkillsSubdir) {
		return "", false, nil
	}
	names, err := core.ListSkillDirs(packageDir, cfg.SkillsSubdir)
	if err != nil {
		return "", false, err
	}
	return "skills", len(names) > 0, nil
}
func (p skillResourceProvider) Remove(projectRoot, name string, cfg *config.Config, isUser bool) (bool, error) {
	installed, err := p.ListInstalled(projectRoot, cfg, isUser)
	if err != nil {
		return false, err
	}
	return removeFromInstalledList(name, installed, func() error {
		skillsDir := config.EffectiveSkillsDir(projectRoot, isUser, cfg)
		if p.opencode {
			skillsDir = config.EffectiveOpenCodeSkillsDir(projectRoot, isUser)
		}
		return core.RemoveSkill(skillsDir, name)
	})
}

type agentResourceProvider struct {
	target   string
	opencode bool
}

func (p agentResourceProvider) Kind() string   { return resourceKindAgent }
func (p agentResourceProvider) Target() string { return p.target }
func (p agentResourceProvider) OutputDir(projectRoot string, cfg *config.Config, isUser bool) string {
	if p.opencode {
		return config.EffectiveOpenCodeAgentsDir(projectRoot, isUser)
	}
	return config.EffectiveAgentsDir(projectRoot, isUser, cfg)
}
func (agentResourceProvider) RequiresName() bool {
	return true
}
func (agentResourceProvider) ListAvailable(packageDir string, cfg *config.Config) ([]string, error) {
	return core.ListAgentFiles(packageDir, cfg.AgentsSubdir)
}
func (p agentResourceProvider) ListInstalled(projectRoot string, cfg *config.Config, isUser bool) ([]string, error) {
	agentsDir := config.EffectiveAgentsDir(projectRoot, isUser, cfg)
	if p.opencode {
		agentsDir = config.EffectiveOpenCodeAgentsDir(projectRoot, isUser)
	}
	return core.ListAgentFilesFrom(agentsDir)
}
func (p agentResourceProvider) Install(projectRoot, packageDir, name string, cfg *config.Config, opts nativeResourceInstallOptions) (core.InstallStrategy, error) {
	agentsDir := config.EffectiveAgentsDir(projectRoot, opts.IsUser, cfg)
	if p.opencode {
		agentsDir = config.EffectiveOpenCodeAgentsDir(projectRoot, opts.IsUser)
	}
	if strings.TrimSpace(name) == core.AgentsSubdir(cfg.AgentsSubdir) {
		return installAllFromProviderWithDir(p, projectRoot, packageDir, cfg, opts.IsUser)
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
func (p agentResourceProvider) IncludeInDefaultInstallAll() bool { return !p.opencode }
func (p agentResourceProvider) DetectDefaultTarget(packageDir, name string, cfg *config.Config) (target string, ok bool, err error) {
	if p.opencode {
		return "", false, nil
	}
	if strings.TrimSpace(name) != core.AgentsSubdir(cfg.AgentsSubdir) {
		return "", false, nil
	}
	names, err := core.ListAgentFiles(packageDir, cfg.AgentsSubdir)
	if err != nil {
		return "", false, err
	}
	return "agents", len(names) > 0, nil
}
func (p agentResourceProvider) Remove(projectRoot, name string, cfg *config.Config, isUser bool) (bool, error) {
	installed, err := p.ListInstalled(projectRoot, cfg, isUser)
	if err != nil {
		return false, err
	}
	return removeFromInstalledList(name, installed, func() error {
		agentsDir := config.EffectiveAgentsDir(projectRoot, isUser, cfg)
		if p.opencode {
			agentsDir = config.EffectiveOpenCodeAgentsDir(projectRoot, isUser)
		}
		return core.RemoveAgent(agentsDir, name)
	})
}

type hooksResourceProvider struct{}

func (hooksResourceProvider) Kind() string   { return resourceKindHooks }
func (hooksResourceProvider) Target() string { return "hooks" }
func (hooksResourceProvider) OutputDir(projectRoot string, cfg *config.Config, isUser bool) string {
	return config.EffectiveHooksDir(projectRoot, isUser, cfg)
}
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
func (p rulesResourceProvider) OutputDir(projectRoot string, cfg *config.Config, isUser bool) string {
	switch p.target {
	case "cursor":
		return config.EffectiveRulesDir(projectRoot, isUser, cfg)
	case "opencode-rules":
		return config.EffectiveOpenCodeRulesDir(projectRoot, isUser)
	default:
		if p.tp != nil {
			if transformer, err := p.tp.Transformer(p.target); err == nil {
				return filepath.Join(projectRoot, transformer.OutputDir())
			}
		}
		return config.EffectiveRulesDir(projectRoot, isUser, cfg)
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
	if p.tp == nil {
		return nil, nil
	}
	transformer, err := p.tp.Transformer(p.target)
	if err != nil {
		return nil, err
	}
	return listInstalledRuleFiles(p.OutputDir(projectRoot, cfg, isUser), transformer.Extension())
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

	if p.target == "opencode-rules" && opts.IsUser {
		rulesDir := config.EffectiveOpenCodeRulesDir(projectRoot, true)
		if isPackage {
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
	if p.tp == nil {
		return false, nil
	}
	installed, err := p.ListInstalled(projectRoot, cfg, isUser)
	if err != nil {
		return false, err
	}
	transformer, err := p.tp.Transformer(p.target)
	if err != nil {
		return false, err
	}
	return removeFromInstalledList(name, installed, func() error {
		return removeInstalledRuleFile(p.OutputDir(projectRoot, cfg, isUser), name, transformer.Extension())
	})
}

func listInstalledRuleFiles(rulesDir, ext string) ([]string, error) {
	entries, err := os.ReadDir(rulesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	out := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ext) {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ext)
		if err := security.ValidatePackageName(name); err != nil {
			continue
		}
		out = append(out, name)
	}
	sort.Strings(out)
	return out, nil
}

func removeInstalledRuleFile(rulesDir, name, ext string) error {
	if err := security.ValidatePackageName(name); err != nil {
		return errors.Wrapf(err, errors.CodeInvalidArgument, "invalid resource name")
	}
	target, err := security.SafeJoin(rulesDir, name+ext)
	if err != nil {
		return errors.Wrapf(err, errors.CodeInvalidArgument, "invalid installed rule file path")
	}
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	return os.Remove(target)
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
