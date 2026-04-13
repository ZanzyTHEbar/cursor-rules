package commands

import (
	"github.com/ZanzyTHEbar/cursor-rules/internal/app"
	"github.com/ZanzyTHEbar/cursor-rules/internal/cli"
	"github.com/ZanzyTHEbar/cursor-rules/internal/cli/display"
	"github.com/spf13/cobra"
)

// NewListCmd returns the list command. Accepts AppContext for parity.
func NewListCmd(ctx *cli.AppContext) *cobra.Command {
	var targetFlag string
	var kindFlag string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List shared content grouped by concrete target",
		Long: `List shared package content grouped by the concrete install targets it can feed.
By default this reads from the configured package directory. With --global (or --dir user),
	it lists installed user resources instead.`,
		Example: `  # Show all targets
  cursor-rules list

  # Show only one target
  cursor-rules list --target opencode-skills

  # Show all targets for one kind
  cursor-rules list --kind command

  # Show installed global Copilot prompts
  cursor-rules list --global --target copilot-prompt`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, isUser, err := cli.ResolveDestination(ctx.App(), cmd)
			if err != nil {
				return err
			}
			resp, err := ctx.App().ListRules(app.ListRequest{Global: isUser, Target: targetFlag, Kind: kindFlag})
			if err != nil {
				return err
			}
			p := display.NewPrinter(ctx.Messenger(), cmd.OutOrStdout(), cmd.ErrOrStderr())
			display.RenderListResponse(p, resp)
			return nil
		},
	}
	cmd.Flags().StringVar(&targetFlag, "target", "", "list only the specified concrete target")
	cmd.Flags().StringVar(&kindFlag, "kind", "", "filter targets by kind: rule|command|skill|agent|hooks")
	return cmd
}
