package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/ZanzyTHEbar/cursor-rules/internal/cli/display"
	"github.com/ZanzyTHEbar/cursor-rules/internal/config"
	gblogger "github.com/ZanzyTHEbar/go-basetools/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Version is set at build time via -ldflags. Defaults to "dev".
var Version = "dev"

var rootCmd *cobra.Command

// Execute builds and runs the root command.
func Execute() {
	if rootCmd == nil {
		rootCmd = BuildRoot(NewAppContext(nil, nil))
	}
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// BuildRoot constructs the root command and wires configuration and templates.
func BuildRoot(ctx *AppContext) *cobra.Command {
	if ctx == nil {
		ctx = NewAppContext(nil, nil)
	}

	root := NewRoot(ctx, DefaultPalette)
	root.Version = Version
	root.PersistentFlags().StringP("workdir", "w", "", "workspace root (defaults to current directory)")
	root.PersistentFlags().String("dir", "", "destination: path or 'user' (shorthand: -w/--workdir for path, --global for user)")
	root.PersistentFlags().Bool("global", false, "use user dirs (~/.cursor/...) as destination (same as --dir user)")

	// postInit loads config into the application and may start background services
	postInit := func(v *viper.Viper) error {
		ctx.Viper = v
		// Initialize logger with configured level (defaults to info)
		level := config.NormalizeLogLevel(v.GetString("logLevel"))
		gblogger.InitLogger(&gblogger.Config{Logger: gblogger.Logger{Style: "text", Level: level}})
		// Load config and optionally start watcher
		ctxBG := context.Background()
		started, resp, err := ctx.App().AutoStartWatcher(ctxBG)
		if err != nil {
			slog.Warn("failed to start watcher", "error", err)
			return nil
		}
		if started && resp != nil {
			slog.Info("watching package dir", "dir", resp.PackageDir)
		}
		return nil
	}
	ConfigureRoot(root, ctx, postInit)

	cobra.AddTemplateFunc("categorizeCommands", CategorizeCommands)
	cobra.AddTemplateFunc("trimTrailingWhitespaces", TrimTrailingWhitespaces)
	cobra.AddTemplateFunc("rpad", Rpad)

	root.SetHelpTemplate(strings.TrimSpace(helpTemplate))
	root.SetUsageTemplate(strings.TrimSpace(usageTemplate))

	root.RunE = func(cmd *cobra.Command, _ []string) error {
		appCtx := ctx.App()
		cfg, cfgPath, err := appCtx.LoadConfig("")
		if err != nil {
			// Config load errors are non-fatal for root command display
			cfg = nil
			cfgPath = ""
		}
		packageDir := appCtx.ResolvePackageDir(cfg)

		if cfgPath == "" {
			if candidate := appCtx.ResolveConfigPath(""); candidate != "" {
				if _, err := os.Stat(candidate); err == nil {
					cfgPath = candidate
				}
			}
		}

		out := cmd.OutOrStdout()
		info := display.BinaryInfo{
			Name:           root.Name(),
			Version:        Version,
			PackageDir:     packageDir,
			ConfigPath:     cfgPath,
			OverrideHint:   "Override packageDir via CURSOR_RULES_PACKAGE_DIR or config.yaml; override config dir via CURSOR_RULES_CONFIG_DIR.",
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

	return root
}

// NewRoot constructs the root command and composes the provided palette using
// the given AppContext.
func NewRoot(ctx *AppContext, p Palette) *cobra.Command {
	if ctx == nil {
		ctx = NewAppContext(nil, nil)
	}
	if p == nil {
		p = Palette{}
	}

	root := &cobra.Command{
		Use:   "cursor-rules",
		Short: "Manage shared Cursor .mdc presets across projects",
		Long:  "cursor-rules is a CLI to install/sync/manage shared Cursor rules in .cursor/rules/ stubs.",
	}

	// compose commands from palette
	root.AddCommand(p.Commands(ctx)...)
	return root
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
