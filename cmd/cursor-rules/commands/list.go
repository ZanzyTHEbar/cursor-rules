package commands

import (
	"fmt"
	"os"

	"github.com/ZanzyTHEbar/cursor-rules/cli"
	cfgpkg "github.com/ZanzyTHEbar/cursor-rules/internal/config"
	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
	"github.com/spf13/cobra"
)

// NewListCmd returns the list command. Accepts AppContext for parity.
func NewListCmd(ctx *cli.AppContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List rules in the configured shared directory",
		RunE: func(cmd *cobra.Command, _ []string) error {
			var ui cli.Messenger
			if ctx != nil {
				ui = ctx.Messenger()
			}
			out := cmd.OutOrStdout()
			info := func(format string, args ...interface{}) {
				if ui != nil {
					ui.Info(format, args...)
					return
				}
				fmt.Fprintf(out, format, args...)
			}

			cfgPath := ""
			if ctx != nil && ctx.Viper != nil {
				cfgPath = ctx.Viper.ConfigFileUsed()
			}
			cfg, _ := cfgpkg.LoadConfig(cfgPath)

			sharedDir := core.DefaultSharedDir()
			if os.Getenv("CURSOR_RULES_DIR") == "" && cfg != nil && cfg.SharedDir != "" {
				sharedDir = cfg.SharedDir
			}

			tree, err := core.BuildRulesTree(sharedDir)
			if err != nil {
				return err
			}

			info("%s\n", core.FormatRulesTree(tree))
			return nil
		},
	}
	return cmd
}
