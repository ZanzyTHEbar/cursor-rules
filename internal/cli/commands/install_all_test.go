package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ZanzyTHEbar/cursor-rules/internal/cli"
	"github.com/spf13/viper"
)

func TestInstallAllInstallsAllPackages(t *testing.T) {
	packageDir := t.TempDir()
	os.Setenv("CURSOR_RULES_PACKAGE_DIR", packageDir)
	defer os.Unsetenv("CURSOR_RULES_PACKAGE_DIR")

	// pkgA
	if err := os.MkdirAll(filepath.Join(packageDir, "pkgA"), 0o755); err != nil {
		t.Fatalf("failed to create pkgA: %v", err)
	}
	if err := os.WriteFile(filepath.Join(packageDir, "pkgA", "a1.mdc"), []byte("---\ndescription: a1\n---\nbody"), 0o644); err != nil {
		t.Fatalf("failed to write pkgA/a1.mdc: %v", err)
	}

	// pkgB
	if err := os.MkdirAll(filepath.Join(packageDir, "pkgB"), 0o755); err != nil {
		t.Fatalf("failed to create pkgB: %v", err)
	}
	if err := os.WriteFile(filepath.Join(packageDir, "pkgB", "b1.mdc"), []byte("---\ndescription: b1\n---\nbody"), 0o644); err != nil {
		t.Fatalf("failed to write pkgB/b1.mdc: %v", err)
	}

	// dot directory should be ignored
	if err := os.MkdirAll(filepath.Join(packageDir, ".git"), 0o755); err != nil {
		t.Fatalf("failed to create .git: %v", err)
	}
	if err := os.WriteFile(filepath.Join(packageDir, ".git", "config"), []byte("noop"), 0o644); err != nil {
		t.Fatalf("failed to write .git/config: %v", err)
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

	// ensure no stray files from ignored dirs
	files, err := os.ReadDir(filepath.Join(workdir, ".cursor", "rules"))
	if err != nil {
		t.Fatalf("read rules dir: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected only 2 files installed, got %d", len(files))
	}
}

func TestInstallAllRespectsViperPackageDirWhenEnvUnset(t *testing.T) {
	packageDir := t.TempDir()
	os.Unsetenv("CURSOR_RULES_PACKAGE_DIR")

	// pkgA
	if err := os.MkdirAll(filepath.Join(packageDir, "pkgA"), 0o755); err != nil {
		t.Fatalf("failed to create pkgA: %v", err)
	}
	if err := os.WriteFile(filepath.Join(packageDir, "pkgA", "a1.mdc"), []byte("---\ndescription: a1\n---\nbody"), 0o644); err != nil {
		t.Fatalf("failed to write pkgA/a1.mdc: %v", err)
	}

	workdir := t.TempDir()
	v := viper.New()
	v.Set("workdir", workdir)
	// Simulate config.yaml being loaded into viper by PersistentPreRunE.
	v.Set("packageDir", packageDir)
	ctx := cli.NewAppContext(v, nil)

	cmd := NewInstallCmd(ctx)
	cmd.SetArgs([]string{"all"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("install all failed: %v", err)
	}

	expectExists(t, filepath.Join(workdir, ".cursor", "rules", "a1.mdc"))
}

func expectExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file to exist: %s (%v)", path, err)
	}
}
