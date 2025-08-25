package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ZanzyTHEbar/cursor-rules/cli"
	cfgpkg "github.com/ZanzyTHEbar/cursor-rules/internal/config"
	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
	"github.com/spf13/cobra"
)

// NewSyncCmd returns the sync command. Accepts AppContext for parity.
func NewSyncCmd(_ *cli.AppContext) *cobra.Command {
	var applyFlag bool
	var dryRunFlag bool

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync shared presets and optionally apply to a project",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load config to honor configured sharedDir when env override is not set
			cfg, _ := cfgpkg.LoadConfig("")

			shared := core.DefaultSharedDir()
			if os.Getenv("CURSOR_RULES_DIR") == "" && cfg != nil && cfg.SharedDir != "" {
				shared = cfg.SharedDir
			}
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

			wd, _ := cmd.Root().Flags().GetString("workdir")
			if applyFlag && wd != "" {
				var toApply []string
				if cfg != nil && len(cfg.Presets) > 0 {
					toApply = cfg.Presets
				} else {
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
	cmd.Flags().BoolVar(&applyFlag, "apply", false, "apply presets to --workdir (uses config.presets if set; otherwise applies all shared presets)")
	cmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "print what would be applied without making changes")
	return cmd
}
