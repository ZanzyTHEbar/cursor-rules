package commands

import (
	"github.com/ZanzyTHEbar/cursor-rules/internal/app"
	"github.com/ZanzyTHEbar/cursor-rules/internal/cli"
	"github.com/ZanzyTHEbar/cursor-rules/internal/cli/display"
	"github.com/spf13/cobra"
)

// NewConfigCmd groups configuration-related helpers under `cursor-rules config`.
func NewConfigCmd(ctx *cli.AppContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage cursor-rules configuration",
	}

	cmd.AddCommand(newConfigInitCmd(ctx))
	cmd.AddCommand(newConfigLinkCmd(ctx))
	return cmd
}

func newConfigLinkCmd(ctx *cli.AppContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "link",
		Short: "Create symlinks from ~/.cursor to CURSOR_*_DIR custom dirs",
		Long:  `When CURSOR_RULES_DIR, CURSOR_COMMANDS_DIR, etc. are set, creates symlinks at ~/.cursor/rules, ~/.cursor/commands, etc. so Cursor sees your custom dirs as user globals.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			resp, err := ctx.App().LinkGlobal(app.LinkGlobalRequest{})
			if err != nil {
				return err
			}
			p := display.NewPrinter(ctx.Messenger(), cmd.OutOrStdout(), cmd.ErrOrStderr())
			display.RenderLinkGlobalResponse(p, resp)
			return nil
		},
	}
	return cmd
}

func newConfigInitCmd(ctx *cli.AppContext) *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create a default config.yaml under the config directory",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfgPath := cli.GetOptionalFlag(cmd, "config")
			req := app.ConfigInitRequest{
				ConfigPath: cfgPath,
				Force:      force,
			}
			resp, err := ctx.App().InitConfig(req)
			if err != nil {
				return err
			}
			p := display.NewPrinter(ctx.Messenger(), cmd.OutOrStdout(), cmd.ErrOrStderr())
			display.RenderConfigInitResponse(p, resp)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "overwrite existing config.yaml (creates a backup)")
	return cmd
}
