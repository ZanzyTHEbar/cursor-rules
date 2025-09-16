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
		RunE: func(cmd *cobra.Command, args []string) error {
			var wd string
			if ctx != nil && ctx.Viper != nil {
				wd = ctx.Viper.GetString("workdir")
			}
			if wd == "" {
				w, _ := cmd.Root().Flags().GetString("workdir")
				if w == "" {
					w, _ = core.WorkingDir()
				}
				wd = w
			}
			presets, err := core.ListProjectPresets(wd)
			if err != nil {
				return err
			}
			for _, p := range presets {
				fmt.Println(p)
			}

			// Also list custom commands if present
			cmds, err := core.ListProjectCommands(wd)
			if err == nil {
				if len(cmds) > 0 {
					fmt.Println("\ncommands:")
					for _, c := range cmds {
						fmt.Println(c)
					}
				}
			}
			return nil
		},
	}
	return cmd
}
