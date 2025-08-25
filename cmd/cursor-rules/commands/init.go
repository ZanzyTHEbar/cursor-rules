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
		RunE: func(cmd *cobra.Command, args []string) error {
			// prefer workdir from AppContext.Viper, fallback to flag
			var wd string
			if ctx != nil && ctx.Viper != nil {
				wd = ctx.Viper.GetString("workdir")
			}
			if wd == "" {
				w, _ := cmd.Root().Flags().GetString("workdir")
				if w == "" {
					w, _ = filepath.Abs(".")
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
