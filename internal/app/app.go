package app

import (
	"path/filepath"
	"strings"

	"github.com/ZanzyTHEbar/cursor-rules/internal/config"
	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
	"github.com/ZanzyTHEbar/cursor-rules/internal/transform"
	"github.com/spf13/viper"
)

// TransformerProvider supplies transformers for targets.
type TransformerProvider interface {
	Transformer(target string) (transform.Transformer, error)
	AvailableTargets() []string
}

// App owns CLI-facing use-case logic.
type App struct {
	Viper        *viper.Viper
	Transformers TransformerProvider
}

// New returns an App instance configured with Viper and transformers.
func New(v *viper.Viper, t TransformerProvider) *App {
	return &App{Viper: v, Transformers: t}
}

// ResolveConfigPath returns the explicit config path or the discovered default.
func (a *App) ResolveConfigPath(explicit string) string {
	if v := strings.TrimSpace(explicit); v != "" {
		return v
	}
	if a != nil && a.Viper != nil {
		if used := strings.TrimSpace(a.Viper.ConfigFileUsed()); used != "" {
			return used
		}
	}
	return config.DefaultConfigPath()
}

// ResolveConfigDir returns the config directory for the resolved config path.
func (a *App) ResolveConfigDir(explicit string) string {
	cfgPath := a.ResolveConfigPath(explicit)
	if cfgPath == "" {
		return ""
	}
	return filepath.Dir(cfgPath)
}

// LoadConfig loads configuration using the resolved config path.
func (a *App) LoadConfig(explicit string) (*config.Config, string, error) {
	cfgPath := a.ResolveConfigPath(explicit)
	cfg, err := config.LoadConfig(cfgPath)
	return cfg, cfgPath, err
}

// ResolvePackageDir returns the effective package directory.
func (a *App) ResolvePackageDir(cfg *config.Config) string {
	if a != nil && a.Viper != nil {
		if v := strings.TrimSpace(a.Viper.GetString("packageDir")); v != "" {
			return v
		}
	}
	return config.ResolvePackageDir(cfg)
}

// ResolveWorkdir resolves the workdir. When allowDefault is false, it returns
// an empty string if no explicit workdir is provided.
func (a *App) ResolveWorkdir(explicit string, allowDefault bool) (string, error) {
	if v := strings.TrimSpace(explicit); v != "" {
		return v, nil
	}
	if a != nil && a.Viper != nil {
		if wd := strings.TrimSpace(a.Viper.GetString("workdir")); wd != "" {
			return wd, nil
		}
	}
	if !allowDefault {
		return "", nil
	}
	return core.WorkingDir()
}
