package main

import (
	"fmt"
	"os"

	"github.com/ZanzyTHEbar/cursor-rules/internal/config"
	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
	"github.com/spf13/cobra"
)

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
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cursor-rules/config.yaml)")
	rootCmd.PersistentFlags().StringP("workdir", "w", "", "workspace root (defaults to current directory)")
}

func initConfig() {
	// Load config and optionally start watcher
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to load config: %v\n", err)
		return
	}
	if cfg.Watch {
		if err := core.StartWatcher(cfg.SharedDir, cfg.AutoApply); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to start watcher: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "watching shared dir: %s\n", cfg.SharedDir)
		}
	}
}
