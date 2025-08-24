package cli

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestPaletteCommandsAndBuildRoot(t *testing.T) {
	ctx := NewAppContext(nil, nil)

	// register a single command factory for testing (no import cycle)
	p := Palette{}
	p.Register(func(_ *AppContext) *cobra.Command {
		return &cobra.Command{
			Use: "policy",
			RunE: func(cmd *cobra.Command, args []string) error {
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
	_, out, err := ExecuteC(root, "policy")
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}
	if !strings.Contains(out, "policy command") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestNewRootHasConfigFlag(t *testing.T) {
	ctx := NewAppContext(nil, nil)
	root := NewRoot(ctx, Palette{})
	if root.PersistentFlags().Lookup("config") == nil {
		t.Fatalf("root missing 'config' persistent flag")
	}
}
