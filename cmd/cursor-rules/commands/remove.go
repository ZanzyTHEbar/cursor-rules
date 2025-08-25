package commands

import (
	"fmt"
	"path/filepath"

	"github.com/ZanzyTHEbar/cursor-rules/cli"
	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
	"github.com/spf13/cobra"
)

// NewRemoveCmd returns the remove command. Accepts AppContext for parity.
func NewRemoveCmd(ctx *cli.AppContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <preset>",
		Short: "Remove a preset stub from the current project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			preset := args[0]
			var wd string
			if ctx != nil && ctx.Viper != nil {
				wd = ctx.Viper.GetString("workdir")
			}
			if wd == "" {
				w, err := cmd.Root().Flags().GetString("workdir")
				if err != nil {
					return err
				}
				if w == "" {
					w, _ = filepath.Abs(".")
				}
				wd = w
			}
			if err := core.RemovePreset(wd, preset); err != nil {
				return fmt.Errorf("remove failed: %w", err)
			}
			fmt.Printf("Removed preset %q from %s/.cursor/rules/\n", preset, wd)
			return nil
		},
	}
	return cmd
}
