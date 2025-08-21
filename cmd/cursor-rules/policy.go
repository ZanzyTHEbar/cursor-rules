package main

import (
	"fmt"

	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
	"github.com/spf13/cobra"
)

var policyCmd = &cobra.Command{
	Use:   "policy",
	Short: "Policy utilities",
}

var policyGenCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a starter config.yaml with presets",
	RunE: func(cmd *cobra.Command, args []string) error {
		shared := core.DefaultSharedDir()
		presets, err := core.ListSharedPresets(shared)
		if err != nil {
			return err
		}
		fmt.Println("# ~/.cursor-rules/config.yaml")
		fmt.Printf("sharedDir: %q\n", shared)
		fmt.Println("watch: false")
		fmt.Println("autoApply: false")
		fmt.Println("presets:")
		for _, p := range presets {
			fmt.Printf("  - %q\n", p[:len(p)-4])
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(policyCmd)
	policyCmd.AddCommand(policyGenCmd)
}
