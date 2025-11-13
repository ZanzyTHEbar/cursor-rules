package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	SharedDir  string
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
		home, err := os.UserHomeDir()
		if err == nil && home != "" {
			v.AddConfigPath(filepath.Join(home, ".cursor", "rules"))
		}
		v.SetConfigName("config")
		v.SetConfigType("yaml")
	}

	// defaults
	v.SetDefault("sharedDir", filepath.Join(os.Getenv("HOME"), ".cursor", "rules"))
	v.SetDefault("watch", false)
	v.SetDefault("autoApply", false)
	v.SetDefault("enableStow", false)
	v.SetDefault("presets", []string{})
	v.SetDefault("logLevel", defaultLogLevel)

	if err := v.ReadInConfig(); err != nil {
		// if not found, return defaults
		cfg := &Config{
			SharedDir:  v.GetString("sharedDir"),
			Watch:      v.GetBool("watch"),
			AutoApply:  v.GetBool("autoApply"),
			EnableStow: v.GetBool("enableStow"),
			Presets:    v.GetStringSlice("presets"),
			LogLevel:   NormalizeLogLevel(v.GetString("logLevel")),
		}
		enableStowIfRequested(cfg)
		return cfg, nil
	}

	cfg := &Config{
		SharedDir:  v.GetString("sharedDir"),
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

// DefaultConfigPath returns the standard config file path in the user's home directory.
func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ""
	}
	return filepath.Join(home, ".cursor", "rules", "config.yaml")
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
