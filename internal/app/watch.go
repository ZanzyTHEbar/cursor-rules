package app

import (
	"context"
	"fmt"

	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
)

// WatchRequest describes a watcher start request.
type WatchRequest struct {
	ConfigPath string
}

// WatchResponse captures watcher configuration.
type WatchResponse struct {
	PackageDir string
	AutoApply  bool
}

// StartWatcher starts a watcher based on config.
func (a *App) StartWatcher(ctx context.Context, req WatchRequest) (*WatchResponse, error) {
	cfg, _, err := a.LoadConfig(req.ConfigPath)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, fmt.Errorf("no config found")
	}
	cfg.PackageDir = a.ResolvePackageDir(cfg)
	if err := core.StartWatcher(ctx, cfg.PackageDir, cfg.AutoApply); err != nil {
		return nil, err
	}
	return &WatchResponse{
		PackageDir: cfg.PackageDir,
		AutoApply:  cfg.AutoApply,
	}, nil
}

// AutoStartWatcher starts the watcher when enabled in config.
func (a *App) AutoStartWatcher(ctx context.Context) (bool, *WatchResponse, error) {
	cfg, _, err := a.LoadConfig("")
	if err != nil {
		return false, nil, err
	}
	if cfg == nil || !cfg.Watch {
		return false, nil, nil
	}
	cfg.PackageDir = a.ResolvePackageDir(cfg)
	if err := core.StartWatcher(ctx, cfg.PackageDir, cfg.AutoApply); err != nil {
		return false, nil, err
	}
	return true, &WatchResponse{
		PackageDir: cfg.PackageDir,
		AutoApply:  cfg.AutoApply,
	}, nil
}
