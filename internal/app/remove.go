package app

import (
	"strings"

	"github.com/ZanzyTHEbar/cursor-rules/internal/config"
	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
	"github.com/ZanzyTHEbar/cursor-rules/internal/errors"
)

// RemoveRequest describes a remove request.
type RemoveRequest struct {
	Name    string
	Type    string // rule, command, skill, agent, hooks (empty = try preset then command)
	Workdir string
	Global  bool // if true, remove from user dirs (~/.cursor/...) instead of project
}

// RemoveResponse captures remove results.
type RemoveResponse struct {
	Name           string
	Workdir        string
	RemovedPreset  bool
	RemovedCommand bool
	RemovedSkill   bool
	RemovedAgent   bool
	RemovedHooks   bool
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

	switch req.Type {
	case "rule":
		if strings.TrimSpace(req.Name) == "" {
			return nil, errors.New(errors.CodeInvalidArgument, "name required for remove --type rule")
		}
		removed, err := removeRuleIfInstalled(wd, req.Name, req.Global, cfg)
		if err != nil {
			return nil, err
		}
		resp.RemovedPreset = removed
		return resp, nil
	}

	if provider, ok := a.resourceRegistry().providerForKind(req.Type); ok {
		if provider.RequiresName() && strings.TrimSpace(req.Name) == "" {
			return nil, errors.Newf(errors.CodeInvalidArgument, "name required for remove --type %s", req.Type)
		}
		removed, err := provider.Remove(wd, req.Name, cfg, req.Global)
		if err != nil {
			return nil, err
		}
		setRemoveFlag(resp, provider.Kind(), removed)
		return resp, nil
	}

	// Default: try preset then command (backward compatible)
	if removed, err := removeRuleIfInstalled(wd, req.Name, req.Global, cfg); err != nil {
		return nil, err
	} else if removed {
		resp.RemovedPreset = true
		return resp, nil
	}
	if provider, ok := a.resourceRegistry().providerForKind(resourceKindCommand); ok {
		removed, err := provider.Remove(wd, req.Name, cfg, req.Global)
		if err != nil {
			return nil, err
		}
		if removed {
			resp.RemovedCommand = true
			return resp, nil
		}
	}
	return resp, nil
}

func removeRuleIfInstalled(projectRoot, name string, isUser bool, cfg *config.Config) (bool, error) {
	rulesDir := config.EffectiveRulesDir(projectRoot, isUser, cfg)
	presets, err := core.ListProjectPresetsFrom(rulesDir)
	if err != nil {
		return false, err
	}
	if !containsInstalledResource(presets, name) {
		return false, nil
	}
	return true, core.RemovePreset(rulesDir, name)
}

func setRemoveFlag(resp *RemoveResponse, kind string, removed bool) {
	if resp == nil || !removed {
		return
	}
	switch kind {
	case resourceKindCommand:
		resp.RemovedCommand = true
	case resourceKindSkill:
		resp.RemovedSkill = true
	case resourceKindAgent:
		resp.RemovedAgent = true
	case resourceKindHooks:
		resp.RemovedHooks = true
	}
}
