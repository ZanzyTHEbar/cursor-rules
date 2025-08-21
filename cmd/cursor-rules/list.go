package main

import (
	"fmt"

	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available shared presets",
	RunE: func(cmd *cobra.Command, args []string) error {
		shared := core.DefaultSharedDir()
		presets, err := core.ListSharedPresets(shared)
		if err != nil {
			return err
		}
		for _, p := range presets {
			fmt.Println(p)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
