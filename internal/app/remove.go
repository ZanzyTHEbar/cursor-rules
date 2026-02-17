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
	wd, err := a.ResolveWorkdir(req.Workdir, true)
	if err != nil {
		return nil, err
	}
	if req.Global {
		wd = config.GlobalProjectRoot()
	}

	resp := &RemoveResponse{
		Name:    req.Name,
		Workdir: wd,
	}

	switch req.Type {
	case "skill":
		if strings.TrimSpace(req.Name) == "" {
			return nil, errors.New(errors.CodeInvalidArgument, "name required for remove --type skill")
		}
		if err := core.RemoveSkill(wd, req.Name); err != nil {
			return nil, err
		}
		resp.RemovedSkill = true
		return resp, nil
	case "agent":
		if strings.TrimSpace(req.Name) == "" {
			return nil, errors.New(errors.CodeInvalidArgument, "name required for remove --type agent")
		}
		if err := core.RemoveAgent(wd, req.Name); err != nil {
			return nil, err
		}
		resp.RemovedAgent = true
		return resp, nil
	case "hooks":
		if err := core.RemoveHookPresetFromProject(wd); err != nil {
			return nil, err
		}
		resp.RemovedHooks = true
		return resp, nil
	case "command":
		if strings.TrimSpace(req.Name) == "" {
			return nil, errors.New(errors.CodeInvalidArgument, "name required for remove --type command")
		}
		if err := core.RemoveCommand(wd, req.Name); err != nil {
			return nil, err
		}
		resp.RemovedCommand = true
		return resp, nil
	case "rule":
		if strings.TrimSpace(req.Name) == "" {
			return nil, errors.New(errors.CodeInvalidArgument, "name required for remove --type rule")
		}
		if err := core.RemovePreset(wd, req.Name); err != nil {
			return nil, err
		}
		resp.RemovedPreset = true
		return resp, nil
	}

	// Default: try preset then command (backward compatible)
	if err := core.RemovePreset(wd, req.Name); err == nil {
		resp.RemovedPreset = true
		return resp, nil
	}
	if err := core.RemoveCommand(wd, req.Name); err == nil {
		resp.RemovedCommand = true
		return resp, nil
	}
	return resp, nil
}
