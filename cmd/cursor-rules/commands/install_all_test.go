package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ZanzyTHEbar/cursor-rules/cli"
	"github.com/spf13/viper"
)

func TestInstallAllInstallsAllPackages(t *testing.T) {
	shared := t.TempDir()
	os.Setenv("CURSOR_RULES_DIR", shared)
	defer os.Unsetenv("CURSOR_RULES_DIR")

	// pkgA
	if err := os.MkdirAll(filepath.Join(shared, "pkgA"), 0o755); err != nil {
		t.Fatalf("failed to create pkgA: %v", err)
	}
	if err := os.WriteFile(filepath.Join(shared, "pkgA", "a1.mdc"), []byte("---\ndescription: a1\n---\nbody"), 0o644); err != nil {
		t.Fatalf("failed to write pkgA/a1.mdc: %v", err)
	}

	// pkgB
	if err := os.MkdirAll(filepath.Join(shared, "pkgB"), 0o755); err != nil {
		t.Fatalf("failed to create pkgB: %v", err)
	}
	if err := os.WriteFile(filepath.Join(shared, "pkgB", "b1.mdc"), []byte("---\ndescription: b1\n---\nbody"), 0o644); err != nil {
		t.Fatalf("failed to write pkgB/b1.mdc: %v", err)
	}

	workdir := t.TempDir()
	v := viper.New()
	v.Set("workdir", workdir)
	ctx := cli.NewAppContext(v, nil)

	cmd := NewInstallCmd(ctx)
	cmd.SetArgs([]string{"all"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("install all failed: %v", err)
	}

	expectExists(t, filepath.Join(workdir, ".cursor", "rules", "a1.mdc"))
	expectExists(t, filepath.Join(workdir, ".cursor", "rules", "b1.mdc"))
}

func expectExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file to exist: %s (%v)", path, err)
	}
}
