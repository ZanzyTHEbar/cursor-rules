package commands

import (
	"fmt"
	"path/filepath"

	"github.com/ZanzyTHEbar/cursor-rules/cli"
	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
	"github.com/spf13/cobra"
)

// NewInitCmd returns the init command
func NewInitCmd(ctx *cli.AppContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a project with .cursor/rules/ directory",
		RunE: func(cmd *cobra.Command, _ []string) error {
			// prefer workdir from AppContext.Viper, fallback to flag
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
					var absErr error
					w, absErr = filepath.Abs(".")
					if absErr != nil {
						return fmt.Errorf("failed to get absolute path: %w", absErr)
					}
				}
				wd = w
			}
			if err := core.InitProject(wd); err != nil {
				return fmt.Errorf("init failed: %w", err)
			}
			fmt.Printf("Initialized project at %s/.cursor/rules/\n", wd)
			return nil
		},
	}
	return cmd
}
