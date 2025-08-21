package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	SharedDir string
	Watch     bool
	AutoApply bool
	Presets   []string
}

// LoadConfig reads config from provided file or default location
func LoadConfig(cfgFile string) (*Config, error) {
	v := viper.New()
	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		home, _ := os.UserHomeDir()
		v.AddConfigPath(filepath.Join(home, ".cursor-rules"))
		v.SetConfigName("config")
		v.SetConfigType("yaml")
	}

	// defaults
	v.SetDefault("sharedDir", filepath.Join(os.Getenv("HOME"), ".cursor-rules"))
	v.SetDefault("watch", false)
	v.SetDefault("autoApply", false)
	v.SetDefault("presets", []string{})

	if err := v.ReadInConfig(); err != nil {
		// if not found, return defaults
		return &Config{
			SharedDir: v.GetString("sharedDir"),
			Watch:     v.GetBool("watch"),
			AutoApply: v.GetBool("autoApply"),
			Presets:   v.GetStringSlice("presets"),
		}, nil
	}

	cfg := &Config{
		SharedDir: v.GetString("sharedDir"),
		Watch:     v.GetBool("watch"),
		AutoApply: v.GetBool("autoApply"),
		Presets:   v.GetStringSlice("presets"),
	}
	return cfg, nil
}
