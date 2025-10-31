package commands

import (
	"github.com/ZanzyTHEbar/cursor-rules/cli"
	"github.com/spf13/cobra"
)

// NewPolicyCmd returns the policy command
func NewPolicyCmd(_ *cli.AppContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy",
		Short: "Manage application policy for presets",
		RunE: func(cmd *cobra.Command, _ []string) error {
			// placeholder: implement policy listing/management
			cmd.Println("policy command not yet implemented")
			return nil
		},
	}
	return cmd
}
