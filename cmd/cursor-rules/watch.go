package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	cfgpkg "github.com/ZanzyTHEbar/cursor-rules/internal/config"
	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Start a long-running watcher that auto-applies presets based on mapping",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := cfgpkg.LoadConfig(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		if cfg == nil {
			return fmt.Errorf("no config found; please provide --config or create $HOME/.cursor-rules/config.yaml")
		}

		if cfg.SharedDir == "" {
			cfg.SharedDir = core.DefaultSharedDir()
		}

		if err := core.StartWatcher(cfg.SharedDir, cfg.AutoApply); err != nil {
			return fmt.Errorf("failed to start watcher: %w", err)
		}
		// wait for termination signal
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		<-ctx.Done()
		fmt.Fprintln(os.Stderr, "watcher: shutting down")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(watchCmd)
}
