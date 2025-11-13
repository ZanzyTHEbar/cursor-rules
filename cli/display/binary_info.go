package display

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

const (
	ansiReset = "\033[0m"
	ansiBold  = "\033[1m"
	ansiCyan  = "\033[36m"
)

// BinaryInfo represents high-level metadata about the CLI binary.
type BinaryInfo struct {
	Name           string
	Version        string
	SharedDir      string
	OverrideHint   string
	ConfigPath     string
	WatcherEnabled bool
	AutoApply      bool
	Tips           []string
}

// StyleOptions control how BinaryInfo is rendered.
type StyleOptions struct {
	EnableColor bool
}

// StyleForWriter infers reasonable style defaults for the supplied writer.
func StyleForWriter(w io.Writer) StyleOptions {
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return StyleOptions{EnableColor: false}
	}

	type fdWriter interface {
		Fd() uintptr
	}

	if f, ok := w.(fdWriter); ok {
		if term.IsTerminal(int(f.Fd())) {
			return StyleOptions{EnableColor: true}
		}
	}

	return StyleOptions{EnableColor: false}
}

// RenderBinaryInfo writes a friendly summary panel about the binary.
func RenderBinaryInfo(w io.Writer, info *BinaryInfo, opts StyleOptions) {
	builder := &strings.Builder{}

	title := info.Name
	if info.Version != "" {
		title = fmt.Sprintf("%s v%s", title, info.Version)
	}
	builder.WriteString(applyHeadingStyle(title, opts))
	builder.WriteString("\n")
	builder.WriteString(strings.Repeat("─", len(stripANSI(title))))
	builder.WriteString("\n\n")

	renderSection(builder, "Paths", []keyValue{
		{Label: "Shared dir", Value: zeroFallback(info.SharedDir, "not configured")},
		{Label: "Config file", Value: zeroFallback(info.ConfigPath, "not found")},
	}, opts)

	if info.OverrideHint != "" {
		builder.WriteString(indent(info.OverrideHint))
		builder.WriteString("\n\n")
	}

	renderSection(builder, "Status", []keyValue{
		{Label: "Watcher", Value: boolStatus(info.WatcherEnabled)},
		{Label: "Auto-apply", Value: boolStatus(info.AutoApply)},
	}, opts)

	if len(info.Tips) > 0 {
		builder.WriteString(applyHeadingStyle("Next steps", opts))
		builder.WriteString("\n")
		for _, tip := range info.Tips {
			builder.WriteString(indent("• " + tip))
			builder.WriteString("\n")
		}
		builder.WriteString("\n")
	}

	fmt.Fprint(w, builder.String())
}

// Heading returns a styled heading string respecting the provided options.
func Heading(text string, opts StyleOptions) string {
	return applyHeadingStyle(text, opts)
}

type keyValue struct {
	Label string
	Value string
}

func renderSection(builder *strings.Builder, title string, kvs []keyValue, opts StyleOptions) {
	builder.WriteString(applyHeadingStyle(title, opts))
	builder.WriteString("\n")

	width := 0
	for _, kv := range kvs {
		if len(kv.Label) > width {
			width = len(kv.Label)
		}
	}

	for _, kv := range kvs {
		builder.WriteString(indent(fmt.Sprintf("%-*s : %s", width, kv.Label, kv.Value)))
		builder.WriteString("\n")
	}
	builder.WriteString("\n")
}

func indent(s string) string {
	return "  " + s
}

func zeroFallback(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func boolStatus(enabled bool) string {
	if enabled {
		return "enabled"
	}
	return "disabled"
}

func applyHeadingStyle(text string, opts StyleOptions) string {
	if !opts.EnableColor {
		return text
	}
	return ansiCyan + ansiBold + text + ansiReset
}

func stripANSI(text string) string {
	var b strings.Builder
	b.Grow(len(text))
	inEscape := false
	for i := 0; i < len(text); i++ {
		ch := text[i]
		if ch == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') {
				inEscape = false
			}
			continue
		}
		b.WriteByte(ch)
	}
	return b.String()
}
