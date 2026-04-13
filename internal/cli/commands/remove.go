package commands

import (
	"errors"

	"github.com/ZanzyTHEbar/cursor-rules/internal/app"
	"github.com/ZanzyTHEbar/cursor-rules/internal/cli"
	"github.com/ZanzyTHEbar/cursor-rules/internal/cli/display"
	"github.com/spf13/cobra"
)

var errNameRequired = errors.New("name required (or use --type hooks to remove project hooks)")

// NewRemoveCmd returns the remove command. Accepts AppContext for parity.
func NewRemoveCmd(ctx *cli.AppContext) *cobra.Command {
	var typeFlag string
	var targetFlag string
	cmd := &cobra.Command{
		Use:   "remove [name]",
		Short: "Remove a preset, command, skill, agent, or hooks from the current project",
		Long: `Remove installed content from the current project or user destination.

Use --target to remove from one concrete target only. Without --target, removal succeeds
only when exactly one installed target matches the given name; if multiple targets match,
	the command errors and asks for disambiguation.`,
		Example: `  # Remove a Cursor rule install
  cursor-rules remove frontend --target cursor

  # Remove an OpenCode command install
  cursor-rules remove review --target opencode-commands

  # Remove configured hooks
  cursor-rules remove --type hooks

  # Remove a global OpenCode skill install
  cursor-rules remove deploy --target opencode-skills --global`,
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if cli.ShowHelpIfReservedArg(cmd, args) {
				return nil
			}
			name := ""
			if len(args) > 0 {
				name = args[0]
			}
			if targetFlag == "" && typeFlag != "hooks" && name == "" {
				return errNameRequired
			}
			workdir, isUser, err := cli.ResolveDestination(ctx.App(), cmd)
			if err != nil {
				return err
			}
			req := app.RemoveRequest{
				Name:    name,
				Type:    typeFlag,
				Target:  targetFlag,
				Workdir: workdir,
				Global:  isUser,
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
	cmd.Flags().StringVar(&typeFlag, "type", "", "type to remove: rule|command|skill|agent|hooks")
	cmd.Flags().StringVar(&targetFlag, "target", "", "remove only from the specified concrete target")
	return cmd
}
