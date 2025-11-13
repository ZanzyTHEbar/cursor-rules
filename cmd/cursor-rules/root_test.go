package main

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestCategorizeCommands(t *testing.T) {
	cmds := []*cobra.Command{
		{Use: "install", Run: func(*cobra.Command, []string) {}},
		{Use: "sync", Run: func(*cobra.Command, []string) {}},
		{Use: "info", Run: func(*cobra.Command, []string) {}},
		{Use: "config", Run: func(*cobra.Command, []string) {}},
	}

	groups := categorizeCommands(cmds)
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
	if groups[0].Name != "Core commands" {
		t.Fatalf("expected first group to be Core commands, got %s", groups[0].Name)
	}
	if len(groups[0].Commands) != 2 {
		t.Fatalf("expected 2 core commands, got %d", len(groups[0].Commands))
	}
	if groups[1].Name != "Utilities" {
		t.Fatalf("expected second group Utilities, got %s", groups[1].Name)
	}
}
