package app

import (
	"os"
	"sort"
	"strings"

	"github.com/ZanzyTHEbar/cursor-rules/internal/config"
	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
)

// InfoRequest describes an info request.
type InfoRequest struct {
	ConfigPath string
	Workdir    string
}

// InfoResponse captures info data for rendering.
type InfoResponse struct {
	ConfigPath   string
	ConfigDir    string
	PackageDir   string
	Watch        bool
	AutoApply    bool
	EnableStow   bool
	EnvOverrides []string
	Workdir      string
	Presets      []string
	Commands     []string
}

// Info returns diagnostics for configuration and workspace.
func (a *App) Info(req InfoRequest) (*InfoResponse, error) {
	cfgPath := ""
	if a != nil && a.Viper != nil {
		cfgPath = a.Viper.ConfigFileUsed()
	}
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		cfg = nil
	}
	if cfgPath == "" {
		if candidate := a.ResolveConfigPath(req.ConfigPath); candidate != "" {
			if _, statErr := os.Stat(candidate); statErr == nil {
				cfgPath = candidate
			}
		}
	}

	packageDir := a.ResolvePackageDir(cfg)
	wd, err := a.ResolveWorkdir(req.Workdir, true)
	if err != nil {
		return nil, err
	}

	presets, err := core.ListProjectPresets(wd)
	if err != nil {
		return nil, err
	}
	sort.Strings(presets)

	customCmds, err := core.ListProjectCommands(wd)
	if err != nil {
		return nil, err
	}
	sort.Strings(customCmds)

	return &InfoResponse{
		ConfigPath:   cfgPath,
		ConfigDir:    a.ResolveConfigDir(req.ConfigPath),
		PackageDir:   packageDir,
		Watch:        cfg != nil && cfg.Watch,
		AutoApply:    cfg != nil && cfg.AutoApply,
		EnableStow:   cfg != nil && cfg.EnableStow || core.WantGNUStow(),
		EnvOverrides: envOverrides(),
		Workdir:      wd,
		Presets:      presets,
		Commands:     customCmds,
	}, nil
}

func envOverrides() []string {
	keys := []string{
		"CURSOR_RULES_CONFIG_DIR",
		"CURSOR_RULES_PACKAGE_DIR",
		"CURSOR_RULES_USE_GNUSTOW",
		"CURSOR_RULES_SYMLINK",
	}
	var out []string
	for _, k := range keys {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" {
			out = append(out, k+"="+v)
		}
	}
	return out
}
