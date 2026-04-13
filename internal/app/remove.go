package app

import (
	"strings"

	"github.com/ZanzyTHEbar/cursor-rules/internal/config"
	"github.com/ZanzyTHEbar/cursor-rules/internal/errors"
)

// RemoveRequest describes a remove request.
type RemoveRequest struct {
	Name    string
	Type    string // kind filter: rule, command, skill, agent, hooks
	Target  string
	Workdir string
	Global  bool // if true, remove from user dirs (~/.cursor/...) instead of project
}

// RemoveMatch captures a target-scoped remove result.
type RemoveMatch struct {
	Target  string
	Kind    string
	Name    string
	Path    string
	Removed bool
}

// RemoveResponse captures remove results.
type RemoveResponse struct {
	Name    string
	Workdir string
	Matches []RemoveMatch
}

// Remove removes a preset, command, skill, agent, or hooks from the project or user dirs when Global is true.
func (a *App) Remove(req RemoveRequest) (*RemoveResponse, error) {
	cfg, _, err := a.LoadConfig("")
	if err != nil {
		return nil, err
	}
	wd, err := a.ResolveWorkdir(req.Workdir, true)
	if err != nil {
		return nil, err
	}
	if req.Global {
		wd = config.GlobalProjectRoot(cfg)
	}

	resp := &RemoveResponse{
		Name:    req.Name,
		Workdir: wd,
	}

	if req.Type != "" && req.Target != "" {
		provider, ok := a.resourceRegistry().providerForTarget(req.Target)
		if !ok {
			return nil, errors.Newf(errors.CodeInvalidArgument, "unknown target: %s", req.Target)
		}
		if provider.Kind() != req.Type {
			return nil, errors.Newf(errors.CodeInvalidArgument, "target %s is not of type %s", req.Target, req.Type)
		}
	}

	providers, err := a.removeProviders(req)
	if err != nil {
		return nil, err
	}

	trimmedName := strings.TrimSpace(req.Name)
	for _, provider := range providers {
		if provider.RequiresName() && trimmedName == "" {
			return nil, errors.Newf(errors.CodeInvalidArgument, "name required for remove target %s", provider.Target())
		}
	}

	if req.Target != "" {
		provider := providers[0]
		removed, err := provider.Remove(wd, req.Name, cfg, req.Global)
		if err != nil {
			return nil, err
		}
		resp.Matches = append(resp.Matches, RemoveMatch{
			Target:  provider.Target(),
			Kind:    provider.Kind(),
			Name:    req.Name,
			Path:    provider.OutputDir(wd, cfg, req.Global),
			Removed: removed,
		})
		return resp, nil
	}

	matches, err := a.findInstalledRemoveMatches(providers, wd, trimmedName, cfg, req.Global)
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return resp, nil
	}
	if len(matches) > 1 {
		return nil, errors.Newf(errors.CodeFailedPrecondition, "remove %q is ambiguous across targets: %s", req.Name, strings.Join(removeTargets(matches), ", "))
	}

	match := matches[0]
	removed, err := match.provider.Remove(wd, req.Name, cfg, req.Global)
	if err != nil {
		return nil, err
	}
	resp.Matches = append(resp.Matches, RemoveMatch{
		Target:  match.provider.Target(),
		Kind:    match.provider.Kind(),
		Name:    req.Name,
		Path:    match.provider.OutputDir(wd, cfg, req.Global),
		Removed: removed,
	})
	return resp, nil
}

func (a *App) removeProviders(req RemoveRequest) ([]nativeResourceProvider, error) {
	registry := a.resourceRegistry()
	if target := strings.TrimSpace(req.Target); target != "" {
		provider, ok := registry.providerForTarget(target)
		if !ok {
			return nil, errors.Newf(errors.CodeInvalidArgument, "unknown target: %s", target)
		}
		return []nativeResourceProvider{provider}, nil
	}
	if kind := strings.TrimSpace(req.Type); kind != "" {
		providers := registry.providersForKind(kind)
		if len(providers) == 0 {
			return nil, errors.Newf(errors.CodeInvalidArgument, "unknown type: %s", kind)
		}
		return providers, nil
	}
	return registry.providers(), nil
}

type installedRemoveMatch struct {
	provider nativeResourceProvider
}

func (a *App) findInstalledRemoveMatches(providers []nativeResourceProvider, projectRoot, name string, cfg *config.Config, isUser bool) ([]installedRemoveMatch, error) {
	trimmedName := strings.TrimSpace(name)
	matches := make([]installedRemoveMatch, 0, len(providers))
	for _, provider := range providers {
		if provider.RequiresName() && trimmedName == "" {
			continue
		}
		if !provider.RequiresName() && trimmedName != "" {
			continue
		}
		installed, err := provider.ListInstalled(projectRoot, cfg, isUser)
		if err != nil {
			return nil, err
		}
		if !provider.RequiresName() {
			if len(installed) > 0 {
				matches = append(matches, installedRemoveMatch{provider: provider})
			}
			continue
		}
		if containsInstalledResource(installed, trimmedName) {
			matches = append(matches, installedRemoveMatch{provider: provider})
		}
	}
	return matches, nil
}

func removeTargets(matches []installedRemoveMatch) []string {
	targets := make([]string, 0, len(matches))
	for _, match := range matches {
		targets = append(targets, match.provider.Target())
	}
	return targets
}
