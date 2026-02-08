package app

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ZanzyTHEbar/cursor-rules/internal/config"
	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
)

// ConfigInitRequest describes config init behavior.
type ConfigInitRequest struct {
	ConfigPath string
	PackageDir string
	Force      bool
}

// ConfigInitResponse captures config init output.
type ConfigInitResponse struct {
	ConfigPath string
	PackageDir string
	BackupPath string
	EnableStow bool
}

// InitConfig creates a config file with defaults.
func (a *App) InitConfig(req ConfigInitRequest) (*ConfigInitResponse, error) {
	cfgPath := a.ResolveConfigPath(req.ConfigPath)
	if cfgPath == "" {
		cfgPath = config.DefaultConfigPath()
	}
	configDir := filepath.Dir(cfgPath)

	packageDir := strings.TrimSpace(req.PackageDir)
	if packageDir == "" {
		packageDir = a.ResolvePackageDir(nil)
	}

	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to prepare config directory %s: %w", configDir, err)
	}
	if packageDir != "" {
		if err := os.MkdirAll(packageDir, 0o755); err != nil {
			return nil, fmt.Errorf("failed to prepare package directory %s: %w", packageDir, err)
		}
	}

	var backupPath string
	if _, statErr := os.Stat(cfgPath); statErr == nil {
		if !req.Force {
			return nil, fmt.Errorf("config already exists at %s (use --force to overwrite)", cfgPath)
		}
		backup, backupErr := backupConfig(cfgPath)
		if backupErr != nil {
			return nil, backupErr
		}
		backupPath = backup
	} else if !errors.Is(statErr, os.ErrNotExist) {
		return nil, fmt.Errorf("failed to inspect existing config: %w", statErr)
	}

	enableStow := core.HasStow()
	content := buildDefaultConfig(packageDir, enableStow)
	if err := core.AtomicWriteString(filepath.Dir(cfgPath), cfgPath, content, 0o644); err != nil {
		return nil, fmt.Errorf("failed to write config: %w", err)
	}

	return &ConfigInitResponse{
		ConfigPath: cfgPath,
		PackageDir: packageDir,
		BackupPath: backupPath,
		EnableStow: enableStow,
	}, nil
}

func backupConfig(path string) (string, error) {
	timestamp := time.Now().Format("20060102-150405")
	backupPath := fmt.Sprintf("%s.%s.bak", path, timestamp)
	if err := os.Rename(path, backupPath); err != nil {
		return "", fmt.Errorf("failed to backup existing config: %w", err)
	}
	return backupPath, nil
}

func buildDefaultConfig(packageDir string, enableStow bool) string {
	var buf bytes.Buffer
	buf.WriteString("# Cursor Rules Manager configuration\n")
	buf.WriteString(fmt.Sprintf("# Generated on %s\n\n", time.Now().Format(time.RFC3339)))
	buf.WriteString("packageDir: ")
	buf.WriteString(packageDir)
	buf.WriteByte('\n')
	buf.WriteString("watch: false\n")
	buf.WriteString("autoApply: false\n")
	fmt.Fprintf(&buf, "enableStow: %t\n", enableStow)
	buf.WriteString("presets: []\n")
	buf.WriteString("logLevel: info\n")
	return buf.String()
}
