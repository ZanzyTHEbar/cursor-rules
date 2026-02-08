package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ZanzyTHEbar/cursor-rules/internal/cli"
	"github.com/ZanzyTHEbar/cursor-rules/internal/cli/commands"
)

func executeWithCapturedOutput(root *cobra.Command, ctx *cli.AppContext, args ...string) (string, error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	if ctx != nil {
		ctx.SetMessenger(cli.NewMessenger(buf, buf, "info"))
	}
	_, err := root.ExecuteC()
	return buf.String(), err
}

func TestInstallCommandCreatesStub(t *testing.T) {
	// prepare package dir with a preset file
	packageDir := t.TempDir()
	presetName := "example"
	presetFile := filepath.Join(packageDir, presetName+".mdc")
	presetContent := `---
description: "Example preset"
apply_to: "**/*.ts"
---
@file /dev/null
`
	if err := os.WriteFile(presetFile, []byte(presetContent), 0o644); err != nil {
		t.Fatalf("write preset: %v", err)
	}
	// set env so core.DefaultPackageDir() will pick this up
	if err := os.Setenv("CURSOR_RULES_PACKAGE_DIR", packageDir); err != nil {
		t.Fatalf("setenv: %v", err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("CURSOR_RULES_PACKAGE_DIR") })

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
	out, err := executeWithCapturedOutput(root, ctx, "install", presetName)
	if err != nil {
		t.Fatalf("install failed: %v; out=%s", err, out)
	}

	// verify stub created
	want := filepath.Join(project, ".cursor", "rules", presetName+".mdc")
	if _, err := os.Stat(want); err != nil {
		t.Fatalf("expected stub at %s: %v; out=%s", want, err, out)
	}
}

func TestInstallCommandCopyReplacesSymlink(t *testing.T) {
	packageDir := t.TempDir()
	presetName := "example"
	presetFile := filepath.Join(packageDir, presetName+".mdc")
	presetContent := `---
description: "Example preset"
apply_to: "**/*.ts"
---
hello
`
	if err := os.WriteFile(presetFile, []byte(presetContent), 0o644); err != nil {
		t.Fatalf("write preset: %v", err)
	}

	t.Setenv("CURSOR_RULES_PACKAGE_DIR", packageDir)
	t.Setenv("CURSOR_RULES_USE_GNUSTOW", "")

	project := t.TempDir()

	v := viper.New()
	v.Set("workdir", project)
	ctx := cli.NewAppContext(v, nil)

	p := cli.Palette{commands.NewInstallCmd}
	root := cli.NewRoot(ctx, p)

	// 1) Install using symlinks
	t.Setenv("CURSOR_RULES_SYMLINK", "1")
	out1, err := executeWithCapturedOutput(root, ctx, "install", presetName)
	if err != nil {
		t.Fatalf("install (symlink) failed: %v; out=%s", err, out1)
	}

	dest := filepath.Join(project, ".cursor", "rules", presetName+".mdc")
	info, err := os.Lstat(dest)
	if err != nil {
		t.Fatalf("expected symlink at %s: %v; out=%s", dest, err, out1)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("expected symlink install, got regular file")
	}

	// 2) Re-run install in copy mode and ensure the symlink is replaced
	t.Setenv("CURSOR_RULES_SYMLINK", "")
	out2, err := executeWithCapturedOutput(root, ctx, "install", presetName)
	if err != nil {
		t.Fatalf("install (copy) failed: %v; out=%s", err, out2)
	}
	if !strings.Contains(out2, "Install method: copy") {
		t.Fatalf("expected install to report copy method; out=%s", out2)
	}

	info2, err := os.Lstat(dest)
	if err != nil {
		t.Fatalf("expected file at %s after copy install: %v; out=%s", dest, err, out2)
	}
	if info2.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("expected copy install to replace symlink, but destination is still a symlink")
	}

	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("read copied file: %v", err)
	}
	gotStr := string(got)
	if strings.Contains(gotStr, "@file ") {
		t.Fatalf("expected copy install to write full content (not a stub), but file still contains an @file reference: %q", gotStr)
	}
	if !strings.Contains(gotStr, "hello") {
		t.Fatalf("expected installed file to contain preset body content; got=%q", gotStr)
	}
}
