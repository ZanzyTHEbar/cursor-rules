package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"

	"github.com/ZanzyTHEbar/cursor-rules/cli"
	"github.com/ZanzyTHEbar/cursor-rules/cmd/cursor-rules/commands"
)

func TestInstallCommandCreatesStub(t *testing.T) {
	// prepare shared dir with a preset file
	shared := t.TempDir()
	presetName := "example"
	presetFile := filepath.Join(shared, presetName+".mdc")
	presetContent := `---
description: "Example preset"
apply_to: "**/*.ts"
---
@file /dev/null
`
	if err := os.WriteFile(presetFile, []byte(presetContent), 0o644); err != nil {
		t.Fatalf("write preset: %v", err)
	}
	// set env so core.DefaultSharedDir() will pick this up
	if err := os.Setenv("CURSOR_RULES_DIR", shared); err != nil {
		t.Fatalf("setenv: %v", err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("CURSOR_RULES_DIR") })

	// prepare project workdir
	project := t.TempDir()

	// create AppContext with workdir in viper
	v := viper.New()
	v.Set("workdir", project)
	ctx := cli.NewAppContext(v, nil)

	// build root with only install command
	p := cli.Palette{commands.NewInstallCmd}
	root := cli.NewRoot(ctx, p)

	// execute install
	_, out, err := cli.ExecuteC(root, "install", presetName)
	if err != nil {
		t.Fatalf("install failed: %v; out=%s", err, out)
	}

	// verify stub created
	want := filepath.Join(project, ".cursor", "rules", presetName+".mdc")
	if _, err := os.Stat(want); err != nil {
		t.Fatalf("expected stub at %s: %v; out=%s", want, err, out)
	}
}
