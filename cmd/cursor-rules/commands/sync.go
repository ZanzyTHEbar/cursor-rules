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
func NewSyncCmd(ctx *cli.AppContext) *cobra.Command {
	var applyFlag bool
	var dryRunFlag bool

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync shared presets and optionally apply to a project",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			var ui cli.Messenger
			if ctx != nil {
				ui = ctx.Messenger()
			}
			stdout := cmd.OutOrStdout()
			stderr := cmd.ErrOrStderr()
			info := func(format string, args ...interface{}) {
				if ui != nil {
					ui.Info(format, args...)
					return
				}
				fmt.Fprintf(stdout, format, args...)
			}
			success := func(format string, args ...interface{}) {
				if ui != nil {
					ui.Success(format, args...)
					return
				}
				fmt.Fprintf(stdout, format, args...)
			}
			errMsg := func(format string, args ...interface{}) {
				if ui != nil {
					ui.Error(format, args...)
					return
				}
				fmt.Fprintf(stderr, format, args...)
			}

			// Load config to honor configured sharedDir when env override is not set
			cfg, err := cfgpkg.LoadConfig("")
			if err != nil {
				// Config load errors are non-fatal; we can proceed with defaults
				cfg = nil
			}

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
			success("Shared dir: %s\n", shared)
			for _, p := range presets {
				info("- %s\n", p)
			}

			// list shared commands too
			commands, err := core.ListSharedCommands(shared)
			if err == nil && len(commands) > 0 {
				info("commands in shared dir:\n")
				for _, c := range commands {
					info("- %s\n", c)
				}
			}

			wd, err := cmd.Root().Flags().GetString("workdir")
			if err != nil {
				return fmt.Errorf("failed to get workdir flag: %w", err)
			}
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
						info("would apply %s -> %s/.cursor/rules/\n", name, wd)
						continue
					}
					strategy, err := core.ApplyPresetToProject(wd, name, shared)
					if err != nil {
						errMsg("failed to apply %s: %v\n", name, err)
					} else {
						success("applied %s -> %s/.cursor/rules/ (method: %s)\n", name, wd, strategy)
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
