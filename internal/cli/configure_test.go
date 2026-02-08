package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

type recordingLogger struct {
	msgs []string
}

func (l *recordingLogger) Printf(format string, v ...interface{}) {
	l.msgs = append(l.msgs, fmt.Sprintf(format, v...))
}

func TestConfigureRootLoadsDefaultConfig(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("CURSOR_RULES_CONFIG_DIR", configDir)
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("packageDir: /tmp\n"), 0o644); err != nil {
		t.Fatalf("failed to seed config: %v", err)
	}

	logger := &recordingLogger{}
	ctx := NewAppContext(nil, logger)
	root := &cobra.Command{Use: "test"}

	ConfigureRoot(root, ctx, nil)
	if err := root.PersistentPreRunE(root, []string{}); err != nil {
		t.Fatalf("pre-run failed: %v", err)
	}

	if got := ctx.Viper.ConfigFileUsed(); got != configPath {
		t.Fatalf("expected config path %s, got %s", configPath, got)
	}
	if len(logger.msgs) != 0 {
		t.Fatalf("expected no logger output, got: %#v", logger.msgs)
	}
}

func TestConfigureRootLogsNoneWhenMissing(t *testing.T) {
	t.Setenv("CURSOR_RULES_CONFIG_DIR", t.TempDir())
	logger := &recordingLogger{}
	ctx := NewAppContext(nil, logger)
	root := &cobra.Command{Use: "test"}

	ConfigureRoot(root, ctx, nil)
	if err := root.PersistentPreRunE(root, []string{}); err != nil {
		t.Fatalf("pre-run failed: %v", err)
	}

	if len(logger.msgs) != 0 {
		t.Fatalf("expected no logger output, got: %#v", logger.msgs)
	}
}

func TestConfigureRootEmitsDebugWhenEnabled(t *testing.T) {
	t.Setenv("CURSOR_RULES_CONFIG_DIR", t.TempDir())

	logger := &recordingLogger{}
	ctx := NewAppContext(nil, logger)
	root := &cobra.Command{Use: "test"}

	var buf bytes.Buffer
	root.SetOut(&buf)

	ConfigureRoot(root, ctx, nil)
	if err := root.PersistentFlags().Set("log-level", "debug"); err != nil {
		t.Fatalf("failed to set log-level flag: %v", err)
	}
	if err := root.PersistentPreRunE(root, []string{}); err != nil {
		t.Fatalf("pre-run failed: %v", err)
	}

	if !strings.Contains(buf.String(), "Using config file") {
		t.Fatalf("expected debug output to contain config usage, got %q", buf.String())
	}
}

func TestConfigureRootEnablesStowWhenConfigured(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("CURSOR_RULES_CONFIG_DIR", configDir)

	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("enableStow: true\n"), 0o644); err != nil {
		t.Fatalf("failed to seed config: %v", err)
	}

	binDir := t.TempDir()
	stowPath := filepath.Join(binDir, "stow")
	if err := os.WriteFile(stowPath, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("failed to create fake stow: %v", err)
	}
	t.Setenv("PATH", binDir)

	ctx := NewAppContext(nil, nil)
	root := &cobra.Command{Use: "test"}
	ConfigureRoot(root, ctx, nil)

	if err := root.PersistentPreRunE(root, []string{}); err != nil {
		t.Fatalf("pre-run failed: %v", err)
	}

	if os.Getenv("CURSOR_RULES_USE_GNUSTOW") != "1" {
		t.Fatalf("expected CURSOR_RULES_USE_GNUSTOW=1 when enableStow is true")
	}
}
