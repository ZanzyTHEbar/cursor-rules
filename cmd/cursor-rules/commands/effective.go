package commands

import (
	"fmt"
	"path/filepath"

	"github.com/ZanzyTHEbar/cursor-rules/cli"
	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
	"github.com/spf13/cobra"
)

// NewEffectiveCmd returns the effective command. Accepts AppContext for parity.
func NewEffectiveCmd(_ *cli.AppContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "effective",
		Short: "Show effective presets applied to a project",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			wd, _ := cmd.Root().Flags().GetString("workdir")
			if wd == "" {
				wd, _ = filepath.Abs(".")
			}
			out, err := core.EffectiveRules(wd)
			if err != nil {
				return err
			}
			fmt.Println(out)
			return nil
		},
	}
	return cmd
}
