package commands

import (
	"github.com/ZanzyTHEbar/cursor-rules/internal/app"
	"github.com/ZanzyTHEbar/cursor-rules/internal/cli"
	"github.com/ZanzyTHEbar/cursor-rules/internal/cli/display"
	"github.com/spf13/cobra"
)

// NewSyncCmd returns the sync command. Accepts AppContext for parity.
func NewSyncCmd(ctx *cli.AppContext) *cobra.Command {
	var applyFlag bool
	var dryRunFlag bool

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync shared presets and optionally apply to a project",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			workdir := cli.GetOptionalFlag(cmd, "workdir")
			cfgPath := cli.GetOptionalFlag(cmd, "config")
			req := app.SyncRequest{
				Apply:      applyFlag,
				DryRun:     dryRunFlag,
				Workdir:    workdir,
				ConfigPath: cfgPath,
			}
			resp, err := ctx.App().Sync(req)
			if err != nil {
				return err
			}
			p := display.NewPrinter(ctx.Messenger(), cmd.OutOrStdout(), cmd.ErrOrStderr())
			display.RenderSyncResponse(p, resp)
			return nil
		},
	}
	cmd.Flags().BoolVar(&applyFlag, "apply", false, "apply presets to --workdir (uses config.presets if set; otherwise applies all shared presets)")
	cmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "print what would be applied without making changes")
	return cmd
}
