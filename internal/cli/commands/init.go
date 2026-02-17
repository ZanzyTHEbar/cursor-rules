package commands

import (
	"github.com/ZanzyTHEbar/cursor-rules/internal/app"
	"github.com/ZanzyTHEbar/cursor-rules/internal/cli"
	"github.com/ZanzyTHEbar/cursor-rules/internal/cli/display"
	"github.com/spf13/cobra"
)

// NewInitCmd returns the init command
func NewInitCmd(ctx *cli.AppContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize project with .cursor/rules, commands, skills, agents, and hooks",
		RunE: func(cmd *cobra.Command, _ []string) error {
			workdir := cli.GetOptionalFlag(cmd, "workdir")
			resp, err := ctx.App().InitProject(app.InitRequest{Workdir: workdir})
			if err != nil {
				return err
			}
			p := display.NewPrinter(ctx.Messenger(), cmd.OutOrStdout(), cmd.ErrOrStderr())
			display.RenderInitResponse(p, resp)
			return nil
		},
	}
	return cmd
}
