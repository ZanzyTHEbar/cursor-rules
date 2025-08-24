package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestPaletteRegisterAndFromCommand(t *testing.T) {
	p := Palette{}

	// create a simple command
	c := &cobra.Command{Use: "foo"}

	// wrap and register
	p.Register(FromCommand(c))

	ctx := NewAppContext(nil, nil)
	cmds := p.Commands(ctx)
	if len(cmds) != 1 {
		t.Fatalf("expected 1 command, got %d", len(cmds))
	}
	if cmds[0].Use != "foo" {
		t.Fatalf("expected use 'foo', got %q", cmds[0].Use)
	}
}

func TestNewRoot_ComposesCommands(t *testing.T) {
	p := Palette{}
	p.Register(func(_ *AppContext) *cobra.Command {
		return &cobra.Command{Use: "bar"}
	})

	ctx := NewAppContext(nil, nil)
	root := NewRoot(ctx, p)
	cmds := root.Commands()
	found := false
	for _, c := range cmds {
		if c.Use == "bar" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected root to contain command 'bar'")
	}
}
