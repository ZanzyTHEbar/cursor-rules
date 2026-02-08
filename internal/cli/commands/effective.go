package commands

import (
	"github.com/ZanzyTHEbar/cursor-rules/internal/app"
	"github.com/ZanzyTHEbar/cursor-rules/internal/cli"
	"github.com/ZanzyTHEbar/cursor-rules/internal/cli/display"
	"github.com/spf13/cobra"
)

// NewEffectiveCmd returns the effective command with multi-target support.
func NewEffectiveCmd(ctx *cli.AppContext) *cobra.Command {
	var targetFlag string

	cmd := &cobra.Command{
		Use:   "effective",
		Short: "Show effective merged rules for current workspace",
		Long: `Display the merged rules that would be active in the current workspace.
For Copilot targets, simulates the non-deterministic merge order.

Examples:
  # Show Cursor rules
  cursor-rules effective
  
  # Show Copilot instructions
  cursor-rules effective --target copilot-instr`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			workdir := cli.GetOptionalFlag(cmd, "workdir")
			req := app.EffectiveRequest{
				Target:  targetFlag,
				Workdir: workdir,
			}
			resp, err := ctx.App().EffectiveRules(req)
			if err != nil {
				return err
			}
			p := display.NewPrinter(ctx.Messenger(), cmd.OutOrStdout(), cmd.ErrOrStderr())
			display.RenderEffectiveResponse(p, resp)
			return nil
		},
	}

	cmd.Flags().StringVar(&targetFlag, "target", "cursor", "target format to show: cursor|copilot-instr|copilot-prompt")

	return cmd
}
