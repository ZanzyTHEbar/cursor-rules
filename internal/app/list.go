package app

import (
	"strings"

	"github.com/ZanzyTHEbar/cursor-rules/internal/config"
	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
	"github.com/ZanzyTHEbar/cursor-rules/internal/errors"
)

// ListRequest describes a rules listing request.
type ListRequest struct {
	ConfigPath string
	Target     string
	Kind       string
	Global     bool // if true, list from user dirs (~/.cursor/...) instead of package dir
}

// ListTargetEntry contains items for one concrete target.
type ListTargetEntry struct {
	Target string
	Kind   string
	Items  []string
}

// ListResponse contains rules tree data plus target-scoped entries.
type ListResponse struct {
	PackageDir string
	Tree       *core.RulesTree
	Targets    []ListTargetEntry
	// Errors holds partial failures from providers (e.g. permissions, missing dirs).
	// When using structured/JSON output, include this field so callers can surface warnings.
	Errors []string
}

func (r *ListResponse) IncludesRules() bool {
	if r == nil {
		return false
	}
	for _, entry := range r.Targets {
		if entry.Kind == resourceKindRule {
			return true
		}
	}
	return false
}

// ListRules returns the rules tree and available commands, skills, agents, and hooks for the configured package directory or user dirs when Global is true.
func (a *App) ListRules(req ListRequest) (*ListResponse, error) {
	cfg, _, err := a.LoadConfig(req.ConfigPath)
	if err != nil {
		return nil, errors.Wrapf(err, errors.CodeInternal, "load config")
	}
	var packageDir string
	providers, err := a.listProviders(req)
	if err != nil {
		return nil, err
	}
	showRulesTree := includesRuleProvider(providers) || (strings.TrimSpace(req.Target) == "" && strings.TrimSpace(req.Kind) == "")
	if req.Global {
		packageDir = a.ResolvePackageDir(cfg)
		resp := &ListResponse{PackageDir: packageDir}
		if showRulesTree {
			tree, err := core.BuildRulesTree(config.EffectiveUserRulesDir(cfg))
			if err != nil {
				return nil, err
			}
			resp.Tree = tree
		}
		projectRoot := config.GlobalProjectRoot(cfg)
		for _, provider := range providers {
			items, err := provider.ListInstalled(projectRoot, cfg, true)
			if err != nil {
				resp.Errors = append(resp.Errors, provider.Target()+": "+err.Error())
				continue
			}
			resp.Targets = append(resp.Targets, ListTargetEntry{
				Target: provider.Target(),
				Kind:   provider.Kind(),
				Items:  items,
			})
		}
		return resp, nil
	}
	packageDir = a.ResolvePackageDir(cfg)
	resp := &ListResponse{PackageDir: packageDir}
	if showRulesTree {
		tree, err := core.BuildRulesTree(packageDir)
		if err != nil {
			return nil, err
		}
		resp.Tree = tree
	}
	for _, provider := range providers {
		items, err := provider.ListAvailable(packageDir, cfg)
		if err != nil {
			resp.Errors = append(resp.Errors, provider.Target()+": "+err.Error())
			continue
		}
		resp.Targets = append(resp.Targets, ListTargetEntry{
			Target: provider.Target(),
			Kind:   provider.Kind(),
			Items:  items,
		})
	}
	return resp, nil
}

func (a *App) listProviders(req ListRequest) ([]nativeResourceProvider, error) {
	registry := a.resourceRegistry()
	if target := strings.TrimSpace(req.Target); target != "" {
		provider, ok := registry.providerForTarget(target)
		if !ok {
			return nil, errors.Newf(errors.CodeInvalidArgument, "unknown target: %s", target)
		}
		if kind := strings.TrimSpace(req.Kind); kind != "" && provider.Kind() != kind {
			return nil, errors.Newf(errors.CodeInvalidArgument, "target %s is not of kind %s", target, kind)
		}
		return []nativeResourceProvider{provider}, nil
	}
	if kind := strings.TrimSpace(req.Kind); kind != "" {
		providers := registry.providersForKind(kind)
		if len(providers) == 0 {
			return nil, errors.Newf(errors.CodeInvalidArgument, "unknown kind: %s", kind)
		}
		return providers, nil
	}
	return registry.providers(), nil
}

func includesRuleProvider(providers []nativeResourceProvider) bool {
	for _, provider := range providers {
		if provider.Kind() == resourceKindRule {
			return true
		}
	}
	return false
}
