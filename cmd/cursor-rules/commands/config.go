package commands

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ZanzyTHEbar/cursor-rules/cli"
	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
	"github.com/ZanzyTHEbar/cursor-rules/internal/security"
	"github.com/spf13/cobra"
)

// NewConfigCmd groups configuration-related helpers under `cursor-rules config`.
func NewConfigCmd(ctx *cli.AppContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage cursor-rules configuration",
	}

	cmd.AddCommand(newConfigInitCmd(ctx))
	return cmd
}

func newConfigInitCmd(ctx *cli.AppContext) *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create a default config.yaml under the shared presets directory",
		RunE: func(cmd *cobra.Command, _ []string) error {
			sharedDir := core.DefaultSharedDir()
			if ctx != nil && ctx.Viper != nil {
				if v := ctx.Viper.GetString("sharedDir"); v != "" {
					sharedDir = v
				}
			}

			if err := os.MkdirAll(sharedDir, 0o755); err != nil {
				return fmt.Errorf("failed to prepare shared directory %s: %w", sharedDir, err)
			}

			cfgPath, err := security.SafeJoin(sharedDir, "config.yaml")
			if err != nil {
				return fmt.Errorf("invalid config path: %w", err)
			}

			if _, statErr := os.Stat(cfgPath); statErr == nil {
				if !force {
					return fmt.Errorf("config already exists at %s (use --force to overwrite)", cfgPath)
				}
				backupPath, backupErr := backupConfig(cfgPath)
				if backupErr != nil {
					return backupErr
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Existing config backed up to %s\n", backupPath)
			} else if !errors.Is(statErr, os.ErrNotExist) {
				return fmt.Errorf("failed to inspect existing config: %w", statErr)
			}

			enableStow := core.HasStow()
			content := buildDefaultConfig(sharedDir, enableStow)
			if err := core.AtomicWriteString(filepath.Dir(cfgPath), cfgPath, content, 0o644); err != nil {
				return fmt.Errorf("failed to write config: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Config written to %s (enableStow=%t)\n", cfgPath, enableStow)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "overwrite existing config.yaml (creates a backup)")
	return cmd
}

func backupConfig(path string) (string, error) {
	timestamp := time.Now().Format("20060102-150405")
	backupPath := fmt.Sprintf("%s.%s.bak", path, timestamp)
	if err := os.Rename(path, backupPath); err != nil {
		return "", fmt.Errorf("failed to backup existing config: %w", err)
	}
	return backupPath, nil
}

func buildDefaultConfig(sharedDir string, enableStow bool) string {
	var buf bytes.Buffer
	buf.WriteString("# Cursor Rules Manager configuration\n")
	buf.WriteString(fmt.Sprintf("# Generated on %s\n\n", time.Now().Format(time.RFC3339)))
	buf.WriteString("sharedDir: ")
	buf.WriteString(sharedDir)
	buf.WriteByte('\n')
	buf.WriteString("watch: false\n")
	buf.WriteString("autoApply: false\n")
	fmt.Fprintf(&buf, "enableStow: %t\n", enableStow)
	buf.WriteString("presets: []\n")
	return buf.String()
}
