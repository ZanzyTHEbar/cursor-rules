package display

import (
	"fmt"
	"io"
	"strings"
)

// InfoView captures structured info output.
type InfoView struct {
	Binary       BinaryInfo
	ConfigPath   string
	ConfigDir    string
	PackageDir   string
	Watch        bool
	AutoApply    bool
	EnableStow   bool
	EnvOverrides []string
	Workdir      string
	Presets      []string
	Commands     []string
}

// RenderInfo renders the info command output.
func RenderInfo(out io.Writer, style StyleOptions, view *InfoView) {
	RenderBinaryInfo(out, &view.Binary, style)

	fmt.Fprintln(out, Heading("Configuration", style))
	writeKeyValues(out, []kv{
		{"Config file", fallbackPath(view.ConfigPath)},
		{"Config dir", fallbackPath(view.ConfigDir)},
		{"Package dir", view.PackageDir},
		{"Watch", boolWord(view.Watch)},
		{"Auto-apply", boolWord(view.AutoApply)},
		{"GNU Stow", boolWord(view.EnableStow)},
	})

	fmt.Fprintln(out, Heading("Environment overrides", style))
	if len(view.EnvOverrides) > 0 {
		for _, line := range view.EnvOverrides {
			fmt.Fprintf(out, "  %s\n", line)
		}
	} else {
		fmt.Fprintln(out, "  (none)")
	}

	fmt.Fprintln(out, Heading("Workspace", style))
	writeKeyValues(out, []kv{
		{"Workdir", view.Workdir},
		{"Presets", summarizeList(view.Presets)},
		{"Commands", summarizeList(view.Commands)},
	})
	if len(view.Presets) > 0 {
		fmt.Fprintln(out, "  preset files:")
		for _, p := range view.Presets {
			fmt.Fprintf(out, "    • %s\n", p)
		}
	}
	if len(view.Commands) > 0 {
		fmt.Fprintln(out, "  custom commands:")
		for _, c := range view.Commands {
			fmt.Fprintf(out, "    • %s\n", c)
		}
	}
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
