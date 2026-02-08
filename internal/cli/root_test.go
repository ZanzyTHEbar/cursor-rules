package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestPaletteCommandsAndNewRoot(t *testing.T) {
	ctx := NewAppContext(nil, nil)

	// register a single command factory for testing (no import cycle)
	p := Palette{}
	p.Register(func(_ *AppContext) *cobra.Command {
		return &cobra.Command{
			Use: "policy",
			RunE: func(cmd *cobra.Command, _ []string) error {
				cmd.Println("policy command not yet implemented")
				return nil
			},
		}
	})

	cmds := p.Commands(ctx)
	if len(cmds) != 1 {
		t.Fatalf("expected 1 command, got %d", len(cmds))
	}
	if cmds[0].Use != "policy" {
		t.Fatalf("expected command use 'policy', got %q", cmds[0].Use)
	}

	root := NewRoot(ctx, p)
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"policy"})
	_, err := root.ExecuteC()
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}
	if !strings.Contains(buf.String(), "policy command") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}
