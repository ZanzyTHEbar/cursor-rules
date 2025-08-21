package main

import (
	"fmt"
	"path/filepath"

	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install <preset>",
	Short: "Install a preset into the current project (.cursor/rules/)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		preset := args[0]
		wd, err := rootCmd.Flags().GetString("workdir")
		if err != nil {
			return err
		}
		if wd == "" {
			wd, _ = filepath.Abs(".")
		}
		if err := core.InstallPreset(wd, preset); err != nil {
			return fmt.Errorf("install failed: %w", err)
		}
		fmt.Printf("Installed preset %q into %s/.cursor/rules/\n", preset, wd)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
