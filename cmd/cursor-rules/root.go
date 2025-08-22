package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/ZanzyTHEbar/cursor-rules/internal/config"
	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
	gblogger "github.com/ZanzyTHEbar/go-basetools/logger"
	"github.com/spf13/cobra"
)

var cfgFile string

// Version is set at build time via -ldflags. Defaults to "dev".
var Version = "dev"

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
	cobra.OnInitialize(initConfig)
	rootCmd.Version = Version
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cursor-rules/config.yaml)")
	rootCmd.PersistentFlags().StringP("workdir", "w", "", "workspace root (defaults to current directory)")
}

func initConfig() {
	// Initialize logger (defaults); can be made configurable later
	gblogger.InitLogger(&gblogger.Config{Logger: gblogger.Logger{Style: "text", Level: "info"}})
	// Load config and optionally start watcher
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		slog.Warn("failed to load config", "error", err)
		return
	}
	if cfg.Watch {
		ctx := context.Background()
		if err := core.StartWatcher(ctx, cfg.SharedDir, cfg.AutoApply); err != nil {
			slog.Warn("failed to start watcher", "error", err)
		} else {
			slog.Info("watching shared dir", "dir", cfg.SharedDir)
		}
	}
}
