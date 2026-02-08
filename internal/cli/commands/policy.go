package commands

import (
	"github.com/ZanzyTHEbar/cursor-rules/internal/cli"
	"github.com/ZanzyTHEbar/cursor-rules/internal/cli/display"
	"github.com/spf13/cobra"
)

// NewPolicyCmd returns the policy command
func NewPolicyCmd(ctx *cli.AppContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy",
		Short: "Manage application policy for presets",
		RunE: func(cmd *cobra.Command, _ []string) error {
			resp := ctx.App().Policy()
			p := display.NewPrinter(ctx.Messenger(), cmd.OutOrStdout(), cmd.ErrOrStderr())
			p.Info("%s\n", resp.Message)
			return nil
		},
	}
	return cmd
}
