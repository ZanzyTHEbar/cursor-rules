package commands

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ZanzyTHEbar/cursor-rules/internal/app"
	"github.com/ZanzyTHEbar/cursor-rules/internal/cli"
	"github.com/ZanzyTHEbar/cursor-rules/internal/cli/display"
	"github.com/spf13/cobra"
)

// NewWatchCmd returns the watch command. Accepts AppContext for parity.
func NewWatchCmd(ctx *cli.AppContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Start a long-running watcher that auto-applies presets based on mapping",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfgPath := cli.GetOptionalFlag(cmd, "config")
			ctxBG, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			_, err := ctx.App().StartWatcher(ctxBG, app.WatchRequest{ConfigPath: cfgPath})
			if err != nil {
				return fmt.Errorf("failed to start watcher: %w", err)
			}

			<-ctxBG.Done()
			p := display.NewPrinter(ctx.Messenger(), cmd.OutOrStdout(), cmd.ErrOrStderr())
			p.Info("watcher: shutting down\n")
			return nil
		},
	}
	return cmd
}
