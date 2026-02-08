package config

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	EnvPackageDir = "CURSOR_RULES_PACKAGE_DIR"
	EnvConfigDir  = "CURSOR_RULES_CONFIG_DIR"
)

// DefaultPackageDir returns the default package directory (~/.cursor/rules) using
// standard library helpers for cross-OS compatibility.
func DefaultPackageDir() string {
	home, err := os.UserHomeDir()
	if err == nil && home != "" {
		return filepath.Join(home, ".cursor", "rules")
	}
	if env := os.Getenv("HOME"); env != "" {
		return filepath.Join(env, ".cursor", "rules")
	}
	if cwd, cwdErr := os.Getwd(); cwdErr == nil && cwd != "" {
		return filepath.Join(cwd, ".cursor", "rules")
	}
	return filepath.Join(".cursor", "rules")
}

// DefaultConfigDir returns the config directory, honoring the env override.
func DefaultConfigDir() string {
	if v := strings.TrimSpace(os.Getenv(EnvConfigDir)); v != "" {
		return v
	}
	return DefaultPackageDir()
}

// DefaultConfigPath returns the standard config file path, honoring config dir override.
func DefaultConfigPath() string {
	return filepath.Join(DefaultConfigDir(), "config.yaml")
}

// ResolvePackageDir returns the effective package directory (env > config > default).
func ResolvePackageDir(cfg *Config) string {
	if v := strings.TrimSpace(os.Getenv(EnvPackageDir)); v != "" {
		return v
	}
	if cfg != nil {
		if v := strings.TrimSpace(cfg.PackageDir); v != "" {
			return v
		}
	}
	return DefaultPackageDir()
}

// ResolveConfigPath returns an explicit config file path if provided, otherwise default.
func ResolveConfigPath(cfgFile string) string {
	if v := strings.TrimSpace(cfgFile); v != "" {
		return v
	}
	return DefaultConfigPath()
}
