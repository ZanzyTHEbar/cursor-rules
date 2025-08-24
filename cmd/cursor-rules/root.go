package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/ZanzyTHEbar/cursor-rules/cli"
	"github.com/ZanzyTHEbar/cursor-rules/internal/config"
	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
	gblogger "github.com/ZanzyTHEbar/go-basetools/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Version is set at build time via -ldflags. Defaults to "dev".
var Version = "dev"

// FIXME: this needs to be implemented
var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "cursor-rules",
	Short: "Manage shared Cursor .mdc presets across projects",
	Long:  "cursor-rules is a CLI to install/sync/manage shared Cursor rules in .cursor/rules/ stubs.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = Version
	rootCmd.PersistentFlags().StringP("workdir", "w", "", "workspace root (defaults to current directory)")

	ctx := cli.NewAppContext(nil, nil)
	// postInit loads config into the application and may start background services
	postInit := func(v *viper.Viper) error {
		// Initialize logger (defaults); can be made configurable later
		gblogger.InitLogger(&gblogger.Config{Logger: gblogger.Logger{Style: "text", Level: "info"}})
		// Load config and optionally start watcher
		cfg, err := config.LoadConfig(v.ConfigFileUsed())
		if err != nil {
			slog.Warn("failed to load config", "error", err)
			return nil
		}
		if cfg.Watch {
			ctxBG := context.Background()
			if err := core.StartWatcher(ctxBG, cfg.SharedDir, cfg.AutoApply); err != nil {
				slog.Warn("failed to start watcher", "error", err)
			} else {
				slog.Info("watching shared dir", "dir", cfg.SharedDir)
			}
		}
		return nil
	}
	cli.ConfigureRoot(rootCmd, ctx, postInit)
}
