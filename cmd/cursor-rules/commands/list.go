package commands

import (
	"fmt"

	"github.com/ZanzyTHEbar/cursor-rules/cli"
	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
	"github.com/spf13/cobra"
)

// NewListCmd returns the list command. Accepts AppContext for parity.
func NewListCmd(ctx *cli.AppContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List installed presets in current project",
		RunE: func(cmd *cobra.Command, _ []string) error {
			var ui cli.Messenger
			if ctx != nil {
				ui = ctx.Messenger()
			}
			out := cmd.OutOrStdout()
			info := func(format string, args ...interface{}) {
				if ui != nil {
					ui.Info(format, args...)
					return
				}
				fmt.Fprintf(out, format, args...)
			}

			var wd string
			if ctx != nil && ctx.Viper != nil {
				wd = ctx.Viper.GetString("workdir")
			}
			if wd == "" {
				w, err := cmd.Root().Flags().GetString("workdir")
				if err != nil {
					return fmt.Errorf("failed to get workdir flag: %w", err)
				}
				if w == "" {
					var err error
					w, err = core.WorkingDir()
					if err != nil {
						return fmt.Errorf("failed to get working directory: %w", err)
					}
				}
				wd = w
			}
			presets, err := core.ListProjectPresets(wd)
			if err != nil {
				return err
			}
			for _, p := range presets {
				info("%s\n", p)
			}

			// Also list custom commands if present
			cmds, err := core.ListProjectCommands(wd)
			if err == nil {
				if len(cmds) > 0 {
					info("\ncommands:\n")
					for _, c := range cmds {
						info("%s\n", c)
					}
				}
			}
			return nil
		},
	}
	return cmd
}
