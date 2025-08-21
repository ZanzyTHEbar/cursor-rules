package main

import (
	"fmt"
	"path/filepath"

	cfgpkg "github.com/ZanzyTHEbar/cursor-rules/internal/config"
	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
	"github.com/spf13/cobra"
)

var applyFlag bool
var dryRunFlag bool

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync shared presets and optionally apply to a project",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		shared := core.DefaultSharedDir()
		// attempt git pull if repo
		if err := core.SyncSharedRepo(shared); err != nil {
			return fmt.Errorf("failed to sync shared repo: %w", err)
		}
		presets, err := core.ListSharedPresets(shared)
		if err != nil {
			return fmt.Errorf("failed to list presets: %w", err)
		}
		fmt.Printf("Shared dir: %s\n", shared)
		for _, p := range presets {
			fmt.Println("-", p)
		}

		// optional apply to project via --apply flag (policy-driven)
		wd, _ := rootCmd.Flags().GetString("workdir")
		if applyFlag && wd != "" {
			// load config to see if presets are specified
			cfg, _ := cfgpkg.LoadConfig("")
			var toApply []string
			if cfg != nil && len(cfg.Presets) > 0 {
				toApply = cfg.Presets
			} else {
				// fallback: apply everything listed
				for _, p := range presets {
					name := p[:len(p)-len(filepath.Ext(p))]
					toApply = append(toApply, name)
				}
			}
			for _, name := range toApply {
				if dryRunFlag {
					fmt.Printf("would apply %s -> %s/.cursor/rules/\n", name, wd)
					continue
				}
				if err := core.ApplyPresetToProject(wd, name, shared); err != nil {
					fmt.Printf("failed to apply %s: %v\n", name, err)
				} else {
					fmt.Printf("applied %s -> %s/.cursor/rules/\n", name, wd)
				}
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.Flags().BoolVar(&applyFlag, "apply", false, "apply presets to --workdir (uses config.presets if set; otherwise applies all shared presets)")
	syncCmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "print what would be applied without making changes")
}
