package commands

import (
	"github.com/ZanzyTHEbar/cursor-rules/internal/app"
	"github.com/ZanzyTHEbar/cursor-rules/internal/cli"
	"github.com/ZanzyTHEbar/cursor-rules/internal/cli/display"
	"github.com/spf13/cobra"
)

// NewTransformCmd returns a command for previewing transformations.
func NewTransformCmd(ctx *cli.AppContext) *cobra.Command {
	var targetFlag string

	cmd := &cobra.Command{
		Use:   "transform <preset>",
		Short: "Preview frontmatter transformation for a preset",
		Long: `Dry-run transformation to see how Cursor rules will be converted
to Copilot format without writing files.

Example:
  cursor-rules transform frontend --target copilot-instr`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			req := app.TransformRequest{
				Name:   args[0],
				Target: targetFlag,
			}
			resp, err := ctx.App().TransformPreview(req)
			if err != nil {
				return err
			}
			p := display.NewPrinter(ctx.Messenger(), cmd.OutOrStdout(), cmd.ErrOrStderr())
			display.RenderTransformResponse(p, resp)
			return nil
		},
	}

	cmd.Flags().StringVar(&targetFlag, "target", "copilot-instr", "target format")

	return cmd
}
