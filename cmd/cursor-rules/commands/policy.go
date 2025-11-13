package commands

import (
	"github.com/ZanzyTHEbar/cursor-rules/cli"
	"github.com/spf13/cobra"
)

// NewPolicyCmd returns the policy command
func NewPolicyCmd(ctx *cli.AppContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy",
		Short: "Manage application policy for presets",
		RunE: func(cmd *cobra.Command, _ []string) error {
			var ui cli.Messenger
			if ctx != nil {
				ui = ctx.Messenger()
			}
			if ui != nil {
				ui.Info("policy command not yet implemented\n")
			} else {
				cmd.Println("policy command not yet implemented")
			}
			return nil
		},
	}
	return cmd
}
