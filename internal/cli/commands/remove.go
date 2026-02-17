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
	cmd := &cobra.Command{
		Use:   "remove [name]",
		Short: "Remove a preset, command, skill, agent, or hooks from the current project",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) > 0 {
				name = args[0]
			}
			if typeFlag != "hooks" && name == "" {
				return errNameRequired
			}
			workdir, isUser, err := cli.ResolveDestination(ctx.App(), cmd)
			if err != nil {
				return err
			}
			req := app.RemoveRequest{
				Name:    name,
				Type:    typeFlag,
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
	cmd.Flags().StringVar(&typeFlag, "type", "", "type to remove: rule|command|skill|agent|hooks (default: try rule then command)")
	return cmd
}
