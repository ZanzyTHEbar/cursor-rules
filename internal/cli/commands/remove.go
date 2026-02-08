package commands

import (
	"github.com/ZanzyTHEbar/cursor-rules/internal/app"
	"github.com/ZanzyTHEbar/cursor-rules/internal/cli"
	"github.com/ZanzyTHEbar/cursor-rules/internal/cli/display"
	"github.com/spf13/cobra"
)

// NewRemoveCmd returns the remove command. Accepts AppContext for parity.
func NewRemoveCmd(ctx *cli.AppContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <preset>",
		Short: "Remove a preset stub from the current project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workdir := cli.GetOptionalFlag(cmd, "workdir")
			req := app.RemoveRequest{
				Name:    args[0],
				Workdir: workdir,
			}
			resp, err := ctx.App().Remove(req)
			if err != nil {
				return err
			}
			p := display.NewPrinter(ctx.Messenger(), cmd.OutOrStdout(), cmd.ErrOrStderr())
			display.RenderRemoveResponse(p, resp)
			return nil
		},
	}
	return cmd
}
