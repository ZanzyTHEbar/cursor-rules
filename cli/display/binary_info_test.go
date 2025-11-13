package display

import (
	"bytes"
	"strings"
	"testing"
)

func TestRenderBinaryInfo_NoColor(t *testing.T) {
	var buf bytes.Buffer
	info := BinaryInfo{
		Name:           "cursor-rules",
		Version:        "1.2.3",
		SharedDir:      "/tmp/shared",
		ConfigPath:     "/home/user/.cursor/rules/config.yaml",
		OverrideHint:   "Override via env var.",
		WatcherEnabled: true,
		AutoApply:      false,
		Tips: []string{
			"Run `cursor-rules sync` to update presets",
			"Use `cursor-rules install <name>` to add a preset",
		},
	}

	RenderBinaryInfo(&buf, &info, StyleOptions{EnableColor: false})
	out := buf.String()

	for _, needle := range []string{
		"cursor-rules v1.2.3",
		"Shared dir  : /tmp/shared",
		"Config file : /home/user/.cursor/rules/config.yaml",
		"Watcher    : enabled",
		"Auto-apply : disabled",
		"Next steps",
		"• Run `cursor-rules sync` to update presets",
	} {
		if !strings.Contains(out, needle) {
			t.Fatalf("expected output to contain %q\n%s", needle, out)
		}
	}
	if strings.Contains(out, "\x1b[") {
		t.Fatalf("expected no ANSI sequences, found output: %q", out)
	}
}

func TestRenderBinaryInfo_Color(t *testing.T) {
	var buf bytes.Buffer
	info := BinaryInfo{Name: "cursor-rules", Version: "dev"}

	RenderBinaryInfo(&buf, &info, StyleOptions{EnableColor: true})
	out := buf.String()
	if !strings.Contains(out, ansiCyan) {
		t.Fatalf("expected ANSI cyan code, got %q", out)
	}
	if !strings.Contains(out, ansiBold) {
		t.Fatalf("expected ANSI bold code, got %q", out)
	}
}
