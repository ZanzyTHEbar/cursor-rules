package commands

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/ZanzyTHEbar/cursor-rules/cli"
	"github.com/ZanzyTHEbar/cursor-rules/cli/display"
	"github.com/ZanzyTHEbar/cursor-rules/internal/config"
	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
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

			cfgPath, cfg := loadConfig(ctx)
			sharedDir := core.DefaultSharedDir()
			if cfg != nil && strings.TrimSpace(cfg.SharedDir) != "" {
				sharedDir = cfg.SharedDir
			}
			if cfgPath == "" {
				if candidate := config.DefaultConfigPath(); candidate != "" {
					if _, err := os.Stat(candidate); err == nil {
						cfgPath = candidate
					}
				}
			}

			info := display.BinaryInfo{
				Name:           "cursor-rules",
				Version:        cmd.Root().Version,
				SharedDir:      sharedDir,
				ConfigPath:     cfgPath,
				WatcherEnabled: cfg != nil && cfg.Watch,
				AutoApply:      cfg != nil && cfg.AutoApply,
				Tips:           nil,
			}
			display.RenderBinaryInfo(out, &info, style)

			fmt.Fprintln(out, display.Heading("Configuration", style))
			writeKeyValues(out, []kv{
				{"Config file", fallbackPath(cfgPath)},
				{"Shared dir", sharedDir},
				{"Watch", boolWord(cfg != nil && cfg.Watch)},
				{"Auto-apply", boolWord(cfg != nil && cfg.AutoApply)},
				{"GNU Stow", boolWord(cfg != nil && cfg.EnableStow || core.WantGNUStow())},
			})

			fmt.Fprintln(out, display.Heading("Environment overrides", style))
			if overrides := envOverrides(); len(overrides) > 0 {
				for _, line := range overrides {
					fmt.Fprintf(out, "  %s\n", line)
				}
			} else {
				fmt.Fprintln(out, "  (none)")
			}

			wd, err := resolveWorkdir(ctx, cmd)
			if err != nil {
				return err
			}

			presets, err := core.ListProjectPresets(wd)
			if err != nil {
				return fmt.Errorf("listing presets: %w", err)
			}
			sort.Strings(presets)

			customCmds, err := core.ListProjectCommands(wd)
			if err != nil {
				return fmt.Errorf("listing project commands: %w", err)
			}
			sort.Strings(customCmds)

			fmt.Fprintln(out, display.Heading("Workspace", style))
			writeKeyValues(out, []kv{
				{"Workdir", wd},
				{"Presets", summarizeList(presets)},
				{"Commands", summarizeList(customCmds)},
			})
			if len(presets) > 0 {
				fmt.Fprintln(out, "  preset files:")
				for _, p := range presets {
					fmt.Fprintf(out, "    • %s\n", p)
				}
			}
			if len(customCmds) > 0 {
				fmt.Fprintln(out, "  custom commands:")
				for _, c := range customCmds {
					fmt.Fprintf(out, "    • %s\n", c)
				}
			}

			return nil
		},
	}
	return cmd
}

func loadConfig(ctx *cli.AppContext) (string, *config.Config) {
	var cfgPath string
	if ctx != nil && ctx.Viper != nil {
		cfgPath = ctx.Viper.ConfigFileUsed()
	}
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return cfgPath, nil
	}
	return cfgPath, cfg
}

func envOverrides() []string {
	keys := []string{
		"CURSOR_RULES_DIR",
		"CURSOR_RULES_USE_GNUSTOW",
		"CURSOR_RULES_SYMLINK",
	}
	var out []string
	for _, k := range keys {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" {
			out = append(out, fmt.Sprintf("%s=%s", k, v))
		}
	}
	return out
}

func summarizeList(items []string) string {
	if len(items) == 0 {
		return "(none)"
	}
	return fmt.Sprintf("%d", len(items))
}

func boolWord(v bool) string {
	if v {
		return "enabled"
	}
	return "disabled"
}

func fallbackPath(p string) string {
	if strings.TrimSpace(p) == "" {
		return "not found"
	}
	return p
}

type kv struct {
	Label string
	Value string
}

func writeKeyValues(out io.Writer, entries []kv) {
	width := 0
	for _, entry := range entries {
		if len(entry.Label) > width {
			width = len(entry.Label)
		}
	}
	for _, entry := range entries {
		fmt.Fprintf(out, "  %-*s : %s\n", width, entry.Label, entry.Value)
	}
	fmt.Fprintln(out)
}

func resolveWorkdir(ctx *cli.AppContext, cmd *cobra.Command) (string, error) {
	if ctx != nil && ctx.Viper != nil {
		if wd := ctx.Viper.GetString("workdir"); wd != "" {
			return wd, nil
		}
	}
	if cmd != nil {
		if flag := cmd.Flags().Lookup("workdir"); flag != nil && flag.Value.String() != "" {
			return flag.Value.String(), nil
		}
		if flag := cmd.InheritedFlags().Lookup("workdir"); flag != nil && flag.Value.String() != "" {
			return flag.Value.String(), nil
		}
		if cmd.Root() != nil {
			if flag := cmd.Root().PersistentFlags().Lookup("workdir"); flag != nil && flag.Value.String() != "" {
				return flag.Value.String(), nil
			}
		}
	}
	return core.WorkingDir()
}
