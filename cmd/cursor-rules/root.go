package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"

	"github.com/ZanzyTHEbar/cursor-rules/cli"
	"github.com/ZanzyTHEbar/cursor-rules/cli/display"
	"github.com/ZanzyTHEbar/cursor-rules/cmd/cursor-rules/commands"
	"github.com/ZanzyTHEbar/cursor-rules/internal/config"
	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
	gblogger "github.com/ZanzyTHEbar/go-basetools/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Version is set at build time via -ldflags. Defaults to "dev".
var Version = "dev"

// cfgFile is intentionally omitted here; configuration is wired by
// `cli.ConfigureRoot` which defines and manages the `--config` flag.

var rootCmd = &cobra.Command{
	Use:   "cursor-rules",
	Short: "Manage shared Cursor .mdc presets across projects",
	Long:  "cursor-rules is a CLI to install/sync/manage shared Cursor rules in .cursor/rules/ stubs.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = Version
	rootCmd.PersistentFlags().StringP("workdir", "w", "", "workspace root (defaults to current directory)")

	ctx := cli.NewAppContext(nil, nil)
	// postInit loads config into the application and may start background services
	postInit := func(v *viper.Viper) error {
		// Initialize logger with configured level (defaults to info)
		level := config.NormalizeLogLevel(v.GetString("logLevel"))
		gblogger.InitLogger(&gblogger.Config{Logger: gblogger.Logger{Style: "text", Level: level}})
		// Load config and optionally start watcher
		cfg, err := config.LoadConfig(v.ConfigFileUsed())
		if err != nil {
			slog.Warn("failed to load config", "error", err)
			return nil
		}
		if cfg.Watch {
			ctxBG := context.Background()
			if err := core.StartWatcher(ctxBG, cfg.SharedDir, cfg.AutoApply); err != nil {
				slog.Warn("failed to start watcher", "error", err)
			} else {
				slog.Info("watching shared dir", "dir", cfg.SharedDir)
			}
		}
		return nil
	}
	cli.ConfigureRoot(rootCmd, ctx, postInit)

	// register all command factories into the global palette
	cli.Register(
		commands.NewInstallCmd,
		commands.NewRemoveCmd,
		commands.NewSyncCmd,
		commands.NewWatchCmd,
		commands.NewListCmd,
		commands.NewEffectiveCmd,
		commands.NewPolicyCmd,
		commands.NewInitCmd,
		commands.NewTransformCmd,
		commands.NewConfigCmd,
		commands.NewInfoCmd,
	)

	// Add commands registered into the global CLI palette (registered in
	// cmd/cursor-rules/init.go). ConfigureRoot only wires config/flags; we
	// still need to attach the concrete subcommands from the global palette.
	rootCmd.AddCommand(cli.DefaultPalette.Commands(ctx)...)

	cobra.AddTemplateFunc("categorizeCommands", categorizeCommands)
	cobra.AddTemplateFunc("trimTrailingWhitespaces", trimTrailingWhitespaces)
	cobra.AddTemplateFunc("rpad", rpad)

	rootCmd.SetHelpTemplate(strings.TrimSpace(helpTemplate))
	rootCmd.SetUsageTemplate(strings.TrimSpace(usageTemplate))

	rootCmd.RunE = func(cmd *cobra.Command, _ []string) error {
		sharedDir := core.DefaultSharedDir()
		var cfg *config.Config
		var err error
		var cfgPath string
		if ctx != nil && ctx.Viper != nil {
			cfgPath = ctx.Viper.ConfigFileUsed()
			cfg, err = config.LoadConfig(cfgPath)
			if err == nil && cfg != nil && cfg.SharedDir != "" {
				sharedDir = cfg.SharedDir
			}
		} else {
			cfg, err = config.LoadConfig("")
			if err == nil && cfg != nil && cfg.SharedDir != "" {
				sharedDir = cfg.SharedDir
			}
		}

		if cfgPath == "" {
			if candidate := config.DefaultConfigPath(); candidate != "" {
				if _, err := os.Stat(candidate); err == nil {
					cfgPath = candidate
				}
			}
		}

		out := cmd.OutOrStdout()
		info := display.BinaryInfo{
			Name:           rootCmd.Name(),
			Version:        Version,
			SharedDir:      sharedDir,
			ConfigPath:     cfgPath,
			OverrideHint:   "Override via CURSOR_RULES_DIR or sharedDir in $HOME/.cursor/rules/config.yaml.",
			WatcherEnabled: cfg != nil && cfg.Watch,
			AutoApply:      cfg != nil && cfg.AutoApply,
			Tips: []string{
				"Run `cursor-rules sync` to refresh shared presets",
				"Use `cursor-rules install <preset>` to add a preset to this workspace",
			},
		}

		display.RenderBinaryInfo(out, &info, display.StyleForWriter(out))
		fmt.Fprintln(out)
		return cmd.Help()
	}
}

type commandGroup struct {
	Name     string
	Commands []*cobra.Command
}

func categorizeCommands(cmds []*cobra.Command) []commandGroup {
	coreNames := map[string]struct{}{
		"install":   {},
		"remove":    {},
		"sync":      {},
		"list":      {},
		"effective": {},
	}

	var coreCmds []*cobra.Command
	var utilityCmds []*cobra.Command
	for _, cmd := range cmds {
		if !cmd.IsAvailableCommand() || cmd.Name() == "help" {
			continue
		}
		if _, ok := coreNames[cmd.Name()]; ok {
			coreCmds = append(coreCmds, cmd)
		} else {
			utilityCmds = append(utilityCmds, cmd)
		}
	}

	sort.Slice(coreCmds, func(i, j int) bool {
		return coreCmds[i].Name() < coreCmds[j].Name()
	})
	sort.Slice(utilityCmds, func(i, j int) bool {
		return utilityCmds[i].Name() < utilityCmds[j].Name()
	})

	var groups []commandGroup
	if len(coreCmds) > 0 {
		groups = append(groups, commandGroup{Name: "Core commands", Commands: coreCmds})
	}
	if len(utilityCmds) > 0 {
		groups = append(groups, commandGroup{Name: "Utilities", Commands: utilityCmds})
	}
	return groups
}

func rpad(s string, padding int) string {
	format := fmt.Sprintf("%%-%ds", padding)
	return fmt.Sprintf(format, s)
}

func trimTrailingWhitespaces(s string) string {
	return strings.TrimRight(s, " \t\r\n")
}

const usageTemplate = `
Usage: {{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Commands:
{{- range $group := categorizeCommands .Commands }}
  {{$group.Name}}:
{{- range $group.Commands }}
    {{rpad .Name .NamePadding }} {{.Short}}
{{- end}}
{{- end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:
{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information.
{{end}}
`

const helpTemplate = `
{{if or .Runnable .HasSubCommands}}
{{.UsageString}}{{end}}
Documentation: https://github.com/ZanzyTHEbar/cursor-rules#readme
`
