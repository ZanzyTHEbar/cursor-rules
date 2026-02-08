package commands

import (
	"github.com/ZanzyTHEbar/cursor-rules/internal/app"
	"github.com/ZanzyTHEbar/cursor-rules/internal/cli"
	"github.com/ZanzyTHEbar/cursor-rules/internal/cli/display"
	"github.com/spf13/cobra"
)

// NewInfoCmd returns the info command which prints detailed diagnostics.
func NewInfoCmd(ctx *cli.AppContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show detailed information about the CLI binary and workspace",
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			style := display.StyleForWriter(out)
			cfgPath := cli.GetOptionalFlag(cmd, "config")
			resp, err := ctx.App().Info(app.InfoRequest{ConfigPath: cfgPath})
			if err != nil {
				return err
			}

			view := display.InfoView{
				Binary: display.BinaryInfo{
					Name:           "cursor-rules",
					Version:        cmd.Root().Version,
					PackageDir:     resp.PackageDir,
					ConfigPath:     resp.ConfigPath,
					WatcherEnabled: resp.Watch,
					AutoApply:      resp.AutoApply,
					Tips:           nil,
				},
				ConfigPath:   resp.ConfigPath,
				ConfigDir:    resp.ConfigDir,
				PackageDir:   resp.PackageDir,
				Watch:        resp.Watch,
				AutoApply:    resp.AutoApply,
				EnableStow:   resp.EnableStow,
				EnvOverrides: resp.EnvOverrides,
				Workdir:      resp.Workdir,
				Presets:      resp.Presets,
				Commands:     resp.Commands,
			}
			display.RenderInfo(out, style, &view)
			return nil
		},
	}
	return cmd
}
