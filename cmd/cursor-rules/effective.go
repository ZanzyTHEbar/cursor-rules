package main

import (
	"fmt"
	"path/filepath"

	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
	"github.com/spf13/cobra"
)

var effectiveCmd = &cobra.Command{
	Use:   "effective",
	Short: "Show merged/effective rules for the project",
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, _ := rootCmd.Flags().GetString("workdir")
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

func init() {
	rootCmd.AddCommand(effectiveCmd)
}
