package main

import (
	"fmt"
	"os"

	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the shared rules directory (~/.cursor-rules)",
	RunE: func(cmd *cobra.Command, args []string) error {
		shared := core.DefaultSharedDir()
		if err := os.MkdirAll(shared, 0o755); err != nil {
			return fmt.Errorf("failed to create shared directory %s: %w", shared, err)
		}
		fmt.Printf("Initialized shared directory: %s\n", shared)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
