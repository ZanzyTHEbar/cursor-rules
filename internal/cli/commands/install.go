package commands

import (
	"github.com/ZanzyTHEbar/cursor-rules/internal/app"
	"github.com/ZanzyTHEbar/cursor-rules/internal/cli"
	"github.com/ZanzyTHEbar/cursor-rules/internal/cli/display"
	"github.com/spf13/cobra"
)

// NewInstallCmd returns the install command with subcommands for rules, commands, skills, agents, hooks.
func NewInstallCmd(ctx *cli.AppContext) *cobra.Command {
	var excludeFlag []string
	var noFlattenFlag bool
	var targetFlag string
	var allTargetsFlag bool

	cmd := &cobra.Command{
		Use:   "install [name]",
		Short: "Install rules, commands, skills, agents, or hooks into the current project",
		Long: `Install rules (default), or use subcommands for commands, skills, agents, hooks.

Examples:
  # Install rules preset (default)
  cursor-rules install frontend
  cursor-rules install frontend --target copilot-instr

  # Install via subcommands (no --target needed)
  cursor-rules install commands my-cmd
  cursor-rules install commands all
  cursor-rules install skills deploy
  cursor-rules install skills all
  cursor-rules install agents code-reviewer
  cursor-rules install hooks my-hooks
  cursor-rules install all`,
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if cli.ShowHelpIfReservedArg(cmd, args) {
				return nil
			}
			if len(args) == 0 {
				return cmd.Help()
			}
			return runInstallRules(ctx, cmd, args[0], excludeFlag, noFlattenFlag, targetFlag, allTargetsFlag)
		},
	}

	cmd.Flags().StringArrayVar(&excludeFlag, "exclude", []string{}, "patterns to exclude when installing a package (can be repeated)")
	cmd.Flags().BoolVarP(&noFlattenFlag, "no-flatten", "n", false, "preserve package directory structure")
	cmd.Flags().StringVar(&targetFlag, "target", "cursor", "rules output target: cursor|copilot-instr|copilot-prompt")
	cmd.Flags().BoolVar(&allTargetsFlag, "all-targets", false, "install to all targets in manifest")

	cmd.AddCommand(newInstallRulesCmd(ctx))
	cmd.AddCommand(newInstallCommandsCmd(ctx))
	cmd.AddCommand(newInstallSkillsCmd(ctx))
	cmd.AddCommand(newInstallAgentsCmd(ctx))
	cmd.AddCommand(newInstallHooksCmd(ctx))
	cmd.AddCommand(newInstallAllCmd(ctx))

	return cmd
}

func runInstallRules(ctx *cli.AppContext, cmd *cobra.Command, name string, exclude []string, noFlatten bool, target string, allTargets bool) error {
	workdir, isUser, err := cli.ResolveDestination(ctx.App(), cmd)
	if err != nil {
		return err
	}
	req := &app.InstallRequest{
		Name:              name,
		Workdir:           workdir,
		Global:            isUser,
		Excludes:          exclude,
		NoFlatten:         noFlatten,
		Target:            target,
		AllTargets:        allTargets,
		ShowInstallMethod: true,
	}
	resp, err := ctx.App().Install(req)
	if err != nil {
		return err
	}
	p := display.NewPrinter(ctx.Messenger(), cmd.OutOrStdout(), cmd.ErrOrStderr())
	display.RenderInstallResponse(p, resp)
	return nil
}

func newInstallRulesCmd(ctx *cli.AppContext) *cobra.Command {
	var excludeFlag []string
	var noFlattenFlag bool
	var targetFlag string
	var allTargetsFlag bool

	c := &cobra.Command{
		Use:   "rules [name|all]",
		Short: "Install a rules preset, package, or all rules",
		Long:  `Install a rules preset or package to .cursor/rules/ or Copilot targets. With no name, installs all rules.`,
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if cli.ShowHelpIfReservedArg(cmd, args) {
				return nil
			}
			if len(args) == 0 || args[0] == "all" {
				workdir, isUser, err := cli.ResolveDestination(ctx.App(), cmd)
				if err != nil {
					return err
				}
				req := &app.InstallAllRequest{
					Workdir:                workdir,
					Global:                 isUser,
					Excludes:               excludeFlag,
					NoFlatten:              noFlattenFlag,
					Target:                 targetFlag,
					AllTargets:             allTargetsFlag,
					ShowInstallMethodFirst: true,
				}
				resp, err := ctx.App().InstallAll(req)
				if err != nil {
					return err
				}
				p := display.NewPrinter(ctx.Messenger(), cmd.OutOrStdout(), cmd.ErrOrStderr())
				display.RenderInstallAllResponse(p, resp)
				return nil
			}
			return runInstallRules(ctx, cmd, args[0], excludeFlag, noFlattenFlag, targetFlag, allTargetsFlag)
		},
	}
	c.Flags().StringArrayVar(&excludeFlag, "exclude", []string{}, "patterns to exclude")
	c.Flags().BoolVarP(&noFlattenFlag, "no-flatten", "n", false, "preserve package structure")
	c.Flags().StringVar(&targetFlag, "target", "cursor", "output target: cursor|copilot-instr|copilot-prompt")
	c.Flags().BoolVar(&allTargetsFlag, "all-targets", false, "install to all targets in manifest")
	return c
}

func newInstallCommandsCmd(ctx *cli.AppContext) *cobra.Command {
	var excludeFlag []string
	var noFlattenFlag bool

	cmd := &cobra.Command{
		Use:   "commands [name|all]",
		Short: "Install a command or all commands as Cursor skills",
		Long:  `Install a command from the package dir into .cursor/skills/ as a Cursor-compatible skill. Use "all" to install the entire commands collection.`,
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if cli.ShowHelpIfReservedArg(cmd, args) {
				return nil
			}
			workdir, isUser, err := cli.ResolveDestination(ctx.App(), cmd)
			if err != nil {
				return err
			}
			name := ""
			if len(args) > 0 {
				name = args[0]
			}
			target := "commands"
			if name == "" || name == "all" {
				req := &app.InstallAllRequest{
					Workdir:                workdir,
					Global:                 isUser,
					Excludes:               excludeFlag,
					NoFlatten:              noFlattenFlag,
					Target:                 target,
					ShowInstallMethodFirst: true,
				}
				resp, err := ctx.App().InstallAll(req)
				if err != nil {
					return err
				}
				p := display.NewPrinter(ctx.Messenger(), cmd.OutOrStdout(), cmd.ErrOrStderr())
				display.RenderInstallAllResponse(p, resp)
				return nil
			}
			req := &app.InstallRequest{
				Name:              name,
				Workdir:           workdir,
				Global:            isUser,
				Excludes:          excludeFlag,
				NoFlatten:         noFlattenFlag,
				Target:            target,
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
	cmd.Flags().StringArrayVar(&excludeFlag, "exclude", []string{}, "patterns to exclude")
	cmd.Flags().BoolVarP(&noFlattenFlag, "no-flatten", "n", false, "preserve package structure")
	return cmd
}

func newInstallSkillsCmd(ctx *cli.AppContext) *cobra.Command {
	var excludeFlag []string

	cmd := &cobra.Command{
		Use:   "skills [name|all]",
		Short: "Install a skill or all skills",
		Long:  `Install a skill from the package dir into .cursor/skills/<name>/. With no name, installs all skills.`,
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if cli.ShowHelpIfReservedArg(cmd, args) {
				return nil
			}
			workdir, isUser, err := cli.ResolveDestination(ctx.App(), cmd)
			if err != nil {
				return err
			}
			name := ""
			if len(args) > 0 {
				name = args[0]
			}
			target := "skills"
			if name == "" || name == "all" {
				req := &app.InstallAllRequest{
					Workdir:                workdir,
					Global:                 isUser,
					Excludes:               excludeFlag,
					Target:                 target,
					ShowInstallMethodFirst: true,
				}
				resp, err := ctx.App().InstallAll(req)
				if err != nil {
					return err
				}
				p := display.NewPrinter(ctx.Messenger(), cmd.OutOrStdout(), cmd.ErrOrStderr())
				display.RenderInstallAllResponse(p, resp)
				return nil
			}
			req := &app.InstallRequest{
				Name:              name,
				Workdir:           workdir,
				Global:            isUser,
				Excludes:          excludeFlag,
				Target:            target,
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
	cmd.Flags().StringArrayVar(&excludeFlag, "exclude", []string{}, "patterns to exclude")
	return cmd
}

func newInstallAgentsCmd(ctx *cli.AppContext) *cobra.Command {
	var excludeFlag []string

	cmd := &cobra.Command{
		Use:   "agents [name|all]",
		Short: "Install an agent or all agents",
		Long:  `Install an agent from the package dir into .cursor/agents/<name>.md. With no name, installs all agents.`,
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if cli.ShowHelpIfReservedArg(cmd, args) {
				return nil
			}
			workdir, isUser, err := cli.ResolveDestination(ctx.App(), cmd)
			if err != nil {
				return err
			}
			name := ""
			if len(args) > 0 {
				name = args[0]
			}
			target := "agents"
			if name == "" || name == "all" {
				req := &app.InstallAllRequest{
					Workdir:                workdir,
					Global:                 isUser,
					Excludes:               excludeFlag,
					Target:                 target,
					ShowInstallMethodFirst: true,
				}
				resp, err := ctx.App().InstallAll(req)
				if err != nil {
					return err
				}
				p := display.NewPrinter(ctx.Messenger(), cmd.OutOrStdout(), cmd.ErrOrStderr())
				display.RenderInstallAllResponse(p, resp)
				return nil
			}
			req := &app.InstallRequest{
				Name:              name,
				Workdir:           workdir,
				Global:            isUser,
				Excludes:          excludeFlag,
				Target:            target,
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
	cmd.Flags().StringArrayVar(&excludeFlag, "exclude", []string{}, "patterns to exclude")
	return cmd
}

func newInstallHooksCmd(ctx *cli.AppContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hooks [preset|all]",
		Short: "Install a hook preset or all hook presets",
		Long:  `Install a hook preset from the package dir into .cursor/hooks.json and .cursor/hooks/. With no preset, installs all hook presets.`,
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if cli.ShowHelpIfReservedArg(cmd, args) {
				return nil
			}
			workdir, isUser, err := cli.ResolveDestination(ctx.App(), cmd)
			if err != nil {
				return err
			}
			if len(args) == 0 || args[0] == "all" {
				req := &app.InstallAllRequest{
					Workdir:                workdir,
					Global:                 isUser,
					Target:                 "hooks",
					ShowInstallMethodFirst: true,
				}
				resp, err := ctx.App().InstallAll(req)
				if err != nil {
					return err
				}
				p := display.NewPrinter(ctx.Messenger(), cmd.OutOrStdout(), cmd.ErrOrStderr())
				display.RenderInstallAllResponse(p, resp)
				return nil
			}
			req := &app.InstallRequest{
				Name:              args[0],
				Workdir:           workdir,
				Global:            isUser,
				Target:            "hooks",
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
	return cmd
}

func newInstallAllCmd(ctx *cli.AppContext) *cobra.Command {
	var excludeFlag []string
	var noFlattenFlag bool
	var targetFlag string
	var allTargetsFlag bool

	cmd := &cobra.Command{
		Use:   "all",
		Short: "Install all packages from the package directory",
		Long:  `Install all installable resources from the package directory. By default this includes rules, commands, skills, agents, and hooks. Use --target to limit to a specific type.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			workdir, isUser, err := cli.ResolveDestination(ctx.App(), cmd)
			if err != nil {
				return err
			}
			req := &app.InstallAllRequest{
				Workdir:                workdir,
				Global:                 isUser,
				Excludes:               excludeFlag,
				NoFlatten:              noFlattenFlag,
				Target:                 targetFlag,
				AllTargets:             allTargetsFlag,
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
	cmd.Flags().StringArrayVar(&excludeFlag, "exclude", []string{}, "patterns to exclude")
	cmd.Flags().BoolVarP(&noFlattenFlag, "no-flatten", "n", false, "preserve package structure")
	cmd.Flags().StringVar(&targetFlag, "target", "cursor", "output target: cursor|copilot-instr|copilot-prompt|commands|skills|agents|hooks")
	cmd.Flags().BoolVar(&allTargetsFlag, "all-targets", false, "install to all targets in manifest")
	return cmd
}
