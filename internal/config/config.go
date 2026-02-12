package config

import (
	"os"
	"os/exec"
	"strings"

	"github.com/ZanzyTHEbar/cursor-rules/internal/errors"
	"github.com/spf13/viper"
)

type Config struct {
	PackageDir string
	Watch      bool
	AutoApply  bool
	EnableStow bool
	Presets    []string
	LogLevel   string
}

const defaultLogLevel = "info"

// LoadConfig reads config from provided file or default location
func LoadConfig(cfgFile string) (*Config, error) {
	v := viper.New()
	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		if cfgDir := DefaultConfigDir(); cfgDir != "" {
			v.AddConfigPath(cfgDir)
		}
		v.SetConfigName("config")
		v.SetConfigType("yaml")
	}

	// defaults
	v.SetDefault("packageDir", DefaultPackageDir())
	v.SetDefault("watch", false)
	v.SetDefault("autoApply", false)
	v.SetDefault("enableStow", false)
	v.SetDefault("presets", []string{})
	v.SetDefault("logLevel", defaultLogLevel)

	if err := v.ReadInConfig(); err != nil {
		// if not found, return defaults
		cfg := &Config{
			PackageDir: v.GetString("packageDir"),
			Watch:      v.GetBool("watch"),
			AutoApply:  v.GetBool("autoApply"),
			EnableStow: v.GetBool("enableStow"),
			Presets:    v.GetStringSlice("presets"),
			LogLevel:   NormalizeLogLevel(v.GetString("logLevel")),
		}
		enableStowIfRequested(cfg)
		return cfg, nil
	}

	if v.IsSet("sharedDir") {
		return nil, errors.ErrLegacyConfigKey("sharedDir", "packageDir")
	}

	cfg := &Config{
		PackageDir: v.GetString("packageDir"),
		Watch:      v.GetBool("watch"),
		AutoApply:  v.GetBool("autoApply"),
		EnableStow: v.GetBool("enableStow"),
		Presets:    v.GetStringSlice("presets"),
		LogLevel:   NormalizeLogLevel(v.GetString("logLevel")),
	}
	enableStowIfRequested(cfg)
	return cfg, nil
}

func enableStowIfRequested(cfg *Config) {
	if cfg == nil || !cfg.EnableStow {
		return
	}

	if _, err := exec.LookPath("stow"); err == nil {
		_ = os.Setenv("CURSOR_RULES_USE_GNUSTOW", "1")
	}
}

// NormalizeLogLevel coerces any provided level string into the supported set.
func NormalizeLogLevel(level string) string {
	switch strings.ToLower(level) {
	case "debug":
		return "debug"
	case "info", "":
		return defaultLogLevel
	case "warn", "warning":
		return "warn"
	case "error":
		return "error"
	default:
		return defaultLogLevel
	}
}
