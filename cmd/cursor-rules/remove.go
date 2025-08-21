package main

import (
	"fmt"
	"path/filepath"

	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <preset>",
	Short: "Remove a preset stub from the current project",
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
		if err := core.RemovePreset(wd, preset); err != nil {
			return fmt.Errorf("remove failed: %w", err)
		}
		fmt.Printf("Removed preset %q from %s/.cursor/rules/\n", preset, wd)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
