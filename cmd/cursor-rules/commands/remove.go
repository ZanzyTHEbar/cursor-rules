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
					var absErr error
					w, absErr = filepath.Abs(".")
					if absErr != nil {
						return fmt.Errorf("failed to get absolute path: %w", absErr)
					}
				}
				wd = w
			}
			// Try removing preset first
			if err := core.RemovePreset(wd, preset); err == nil {
				fmt.Printf("Removed preset %q from %s/.cursor/rules/\n", preset, wd)
				return nil
			}
			// If not a preset, try removing a command
			if err := core.RemoveCommand(wd, preset); err == nil {
				fmt.Printf("Removed command %q from %s/.cursor/commands/\n", preset, wd)
				return nil
			}
			// Both operations returned nil (files don't exist), which is fine
			return nil
		},
	}
	return cmd
}
