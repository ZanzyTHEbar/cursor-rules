package commands

import (
	"github.com/ZanzyTHEbar/cursor-rules/internal/app"
	"github.com/ZanzyTHEbar/cursor-rules/internal/cli"
	"github.com/ZanzyTHEbar/cursor-rules/internal/cli/display"
	"github.com/spf13/cobra"
)

// NewInstallCmd returns the install command with transformer support.
func NewInstallCmd(ctx *cli.AppContext) *cobra.Command {
	var excludeFlag []string
	var noFlattenFlag bool
	var targetFlag string
	var allTargetsFlag bool

	var excludeAllFlag []string
	var noFlattenAllFlag bool
	var targetAllFlag string
	var allTargetsAllFlag bool

	cmd := &cobra.Command{
		Use:   "install <preset|package>",
		Short: "Install a preset or package into the current project",
		Long: `Install a preset or package into .cursor/rules/ (default) or 
.github/instructions/ or .github/prompts/ depending on --target flag.

Examples:
  # Install to Cursor (default)
  cursor-rules install frontend
  
  # Install to Copilot Instructions
  cursor-rules install frontend --target copilot-instr
  
  # Install to Copilot Prompts
  cursor-rules install frontend --target copilot-prompt
  
  # Install to all targets defined in manifest
  cursor-rules install frontend --all-targets`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workdir, isUser, err := cli.ResolveDestination(ctx.App(), cmd)
			if err != nil {
				return err
			}
			req := &app.InstallRequest{
				Name:              args[0],
				Workdir:           workdir,
				Global:            isUser,
				Excludes:          excludeFlag,
				NoFlatten:         noFlattenFlag,
				Target:            targetFlag,
				AllTargets:        allTargetsFlag,
				ShowInstallMethod: true,
			}
			resp, err := ctx.App().Install(req)
			if err != nil {
				return err
			}
			p := display.NewPrinter(ctx.Messenger(), cmd.OutOrStdout(), cmd.ErrOrStderr())
			display.RenderInstallResponse(p, resp)
			return nil
		},
	}

	cmd.Flags().StringArrayVar(&excludeFlag, "exclude", []string{}, "patterns to exclude when installing a package (can be repeated)")
	cmd.Flags().BoolVarP(&noFlattenFlag, "no-flatten", "n", false, "preserve package directory structure")
	cmd.Flags().StringVar(&targetFlag, "target", "cursor", "output target: cursor|copilot-instr|copilot-prompt|cursor-commands|cursor-skills|cursor-agents|cursor-hooks")
	cmd.Flags().BoolVar(&allTargetsFlag, "all-targets", false, "install to all targets in manifest")

	allCmd := &cobra.Command{
		Use:   "all",
		Short: "Install all packages from the package directory",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			workdir, isUser, err := cli.ResolveDestination(ctx.App(), cmd)
			if err != nil {
				return err
			}
			req := &app.InstallAllRequest{
				Workdir:                workdir,
				Global:                 isUser,
				Excludes:               excludeAllFlag,
				NoFlatten:              noFlattenAllFlag,
				Target:                 targetAllFlag,
				AllTargets:             allTargetsAllFlag,
				ShowInstallMethodFirst: true,
			}
			resp, err := ctx.App().InstallAll(req)
			if err != nil {
				return err
			}
			p := display.NewPrinter(ctx.Messenger(), cmd.OutOrStdout(), cmd.ErrOrStderr())
			display.RenderInstallAllResponse(p, resp)
			return nil
		},
	}

	allCmd.Flags().StringArrayVar(&excludeAllFlag, "exclude", []string{}, "patterns to exclude when installing a package (can be repeated)")
	allCmd.Flags().BoolVarP(&noFlattenAllFlag, "no-flatten", "n", false, "preserve package directory structure")
	allCmd.Flags().StringVar(&targetAllFlag, "target", "cursor", "output target: cursor|copilot-instr|copilot-prompt|cursor-commands|cursor-skills|cursor-agents|cursor-hooks")
	allCmd.Flags().BoolVar(&allTargetsAllFlag, "all-targets", false, "install to all targets in manifest")

	cmd.AddCommand(allCmd)

	return cmd
}
