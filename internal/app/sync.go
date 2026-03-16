package app

import (
	"path/filepath"

	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
	"github.com/ZanzyTHEbar/cursor-rules/internal/errors"
)

// SyncRequest describes a sync request.
type SyncRequest struct {
	Apply      bool
	DryRun     bool
	Workdir    string
	ConfigPath string
}

// SyncApplyResult captures a single apply result.
type SyncApplyResult struct {
	Name     string
	Workdir  string
	DryRun   bool
	Strategy core.InstallStrategy
	Error    string
}

// SyncResponse captures sync output.
type SyncResponse struct {
	PackageDir        string
	Presets           []string
	Commands          []string
	Skills            []string
	Agents            []string
	Hooks             []string
	Applied           []SyncApplyResult
	ApplySkipped      bool
	Workdir           string
	UsedConfigPresets bool
}

// Sync synchronizes the package repo and optionally applies presets.
func (a *App) Sync(req SyncRequest) (*SyncResponse, error) {
	cfg, _, err := a.LoadConfig(req.ConfigPath)
	if err != nil {
		return nil, errors.Wrapf(err, errors.CodeInternal, "load config")
	}
	packageDir := a.ResolvePackageDir(cfg)

	if err := core.SyncPackageRepo(packageDir); err != nil {
		return nil, err
	}
	presets, err := core.ListPackagePresets(packageDir)
	if err != nil {
		return nil, err
	}
	resp := &SyncResponse{
		PackageDir: packageDir,
		Presets:    presets,
	}
	for _, provider := range a.resourceRegistry().providers() {
		items, listErr := provider.ListAvailable(packageDir, cfg)
		if listErr != nil {
			if provider.Kind() == resourceKindCommand {
				return nil, errors.Wrapf(listErr, errors.CodeInternal, "list shared commands")
			}
			continue
		}
		assignSyncItems(resp, provider.Kind(), items)
	}

	if !req.Apply {
		return resp, nil
	}

	if req.Workdir == "" {
		resp.ApplySkipped = true
		return resp, nil
	}

	resp.Workdir = req.Workdir
	var toApply []string
	if cfg != nil && len(cfg.Presets) > 0 {
		toApply = append(toApply, cfg.Presets...)
		resp.UsedConfigPresets = true
	} else {
		for _, p := range presets {
			name := p[:len(p)-len(filepath.Ext(p))]
			toApply = append(toApply, name)
		}
	}

	for _, name := range toApply {
		if req.DryRun {
			resp.Applied = append(resp.Applied, SyncApplyResult{
				Name:    name,
				Workdir: req.Workdir,
				DryRun:  true,
			})
			continue
		}
		strategy, err := core.ApplyPresetToProject(req.Workdir, name, packageDir)
		if err != nil {
			resp.Applied = append(resp.Applied, SyncApplyResult{
				Name:    name,
				Workdir: req.Workdir,
				Error:   err.Error(),
			})
			continue
		}
		resp.Applied = append(resp.Applied, SyncApplyResult{
			Name:     name,
			Workdir:  req.Workdir,
			Strategy: strategy,
		})
	}

	return resp, nil
}

func assignSyncItems(resp *SyncResponse, kind string, items []string) {
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
