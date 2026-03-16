package app

import (
	"github.com/ZanzyTHEbar/cursor-rules/internal/config"
	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
	"github.com/ZanzyTHEbar/cursor-rules/internal/errors"
)

// ListRequest describes a rules listing request.
type ListRequest struct {
	ConfigPath string
	Global     bool // if true, list from user dirs (~/.cursor/...) instead of package dir
}

// ListResponse contains rules tree data plus commands, skills, agents, and hooks from the package dir.
type ListResponse struct {
	PackageDir string
	Tree       *core.RulesTree
	Commands   []string
	Skills     []string
	Agents     []string
	Hooks      []string
	// Errors holds partial failures from providers (e.g. permissions, missing dirs).
	// When using structured/JSON output, include this field so callers can surface warnings.
	Errors []string
}

// ListRules returns the rules tree and available commands, skills, agents, and hooks for the configured package directory or user dirs when Global is true.
func (a *App) ListRules(req ListRequest) (*ListResponse, error) {
	cfg, _, err := a.LoadConfig(req.ConfigPath)
	if err != nil {
		return nil, errors.Wrapf(err, errors.CodeInternal, "load config")
	}
	var packageDir string
	if req.Global {
		packageDir = a.ResolvePackageDir(cfg)
		tree, err := core.BuildRulesTree(config.EffectiveUserRulesDir(cfg))
		if err != nil {
			return nil, err
		}
		resp := &ListResponse{
			PackageDir: packageDir,
			Tree:       tree,
		}
		projectRoot := config.GlobalProjectRoot(cfg)
		for _, provider := range a.resourceRegistry().providers() {
			items, err := provider.ListInstalled(projectRoot, cfg, true)
			if err != nil {
				resp.Errors = append(resp.Errors, provider.Kind()+": "+err.Error())
				continue
			}
			assignProviderItems(resp, provider.Kind(), items)
		}
		return resp, nil
	}
	packageDir = a.ResolvePackageDir(cfg)
	tree, err := core.BuildRulesTree(packageDir)
	if err != nil {
		return nil, err
	}
	resp := &ListResponse{
		PackageDir: packageDir,
		Tree:       tree,
	}
	for _, provider := range a.resourceRegistry().providers() {
		items, err := provider.ListAvailable(packageDir, cfg)
		if err != nil {
			resp.Errors = append(resp.Errors, provider.Kind()+": "+err.Error())
			continue
		}
		assignProviderItems(resp, provider.Kind(), items)
	}
	return resp, nil
}

func assignProviderItems(resp *ListResponse, kind string, items []string) {
	if resp == nil {
		return
	}
	switch kind {
	case resourceKindCommand:
		resp.Commands = items
	case resourceKindSkill:
		resp.Skills = items
	case resourceKindAgent:
		resp.Agents = items
	case resourceKindHooks:
		resp.Hooks = items
	}
}
