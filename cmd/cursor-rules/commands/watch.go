package commands

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ZanzyTHEbar/cursor-rules/cli"
	cfgpkg "github.com/ZanzyTHEbar/cursor-rules/internal/config"
	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
	"github.com/spf13/cobra"
)

// NewWatchCmd returns the watch command. Accepts AppContext for parity.
func NewWatchCmd(ctx *cli.AppContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Start a long-running watcher that auto-applies presets based on mapping",
		RunE: func(cmd *cobra.Command, _ []string) error {
			var ui cli.Messenger
			if ctx != nil {
				ui = ctx.Messenger()
			}
			// prefer viper-configured path, fallback to flag
			var cfgFileFlag string
			if ctx != nil && ctx.Viper != nil {
				cfgFileFlag = ctx.Viper.GetString("config")
			}
			if cfgFileFlag == "" {
				var err error
				cfgFileFlag, err = cmd.Root().Flags().GetString("config")
				if err != nil {
					return fmt.Errorf("failed to get config flag: %w", err)
				}
			}
			cfg, err := cfgpkg.LoadConfig(cfgFileFlag)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			if cfg == nil {
				return fmt.Errorf("no config found; please provide --config or create $HOME/.cursor/rules/config.yaml")
			}
			if cfg.SharedDir == "" {
				cfg.SharedDir = core.DefaultSharedDir()
			}
			ctxBG, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			if err := core.StartWatcher(ctxBG, cfg.SharedDir, cfg.AutoApply); err != nil {
				return fmt.Errorf("failed to start watcher: %w", err)
			}
			<-ctxBG.Done()
			if ui != nil {
				ui.Info("watcher: shutting down\n")
			} else {
				fmt.Fprintln(os.Stderr, "watcher: shutting down")
			}
			return nil
		},
	}
	return cmd
}
