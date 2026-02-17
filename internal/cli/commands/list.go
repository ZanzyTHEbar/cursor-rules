package commands

import (
	"github.com/ZanzyTHEbar/cursor-rules/internal/app"
	"github.com/ZanzyTHEbar/cursor-rules/internal/cli"
	"github.com/ZanzyTHEbar/cursor-rules/internal/cli/display"
	"github.com/spf13/cobra"
)

// NewListCmd returns the list command. Accepts AppContext for parity.
func NewListCmd(ctx *cli.AppContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List rules in the configured package directory",
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, isUser, err := cli.ResolveDestination(ctx.App(), cmd)
			if err != nil {
				return err
			}
			resp, err := ctx.App().ListRules(app.ListRequest{Global: isUser})
			if err != nil {
				return err
			}
			p := display.NewPrinter(ctx.Messenger(), cmd.OutOrStdout(), cmd.ErrOrStderr())
			display.RenderListResponse(p, resp)
			return nil
		},
	}
	return cmd
}
