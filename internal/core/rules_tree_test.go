package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildRulesTreeAndFormat(t *testing.T) {
	shared := t.TempDir()

	writeFile(t, filepath.Join(shared, "root.mdc"), "content")
	writeFile(t, filepath.Join(shared, "readme.md"), "readme")

	pkgA := filepath.Join(shared, "pkgA")
	writeFile(t, filepath.Join(pkgA, "file1.mdc"), "a1")
	writeFile(t, filepath.Join(pkgA, "nested", "inner.mdc"), "inner")
	writeFile(t, filepath.Join(pkgA, "skip.mdc"), "skip me")
	writeFile(t, filepath.Join(pkgA, ".cursor-rules-ignore"), "skip.mdc\n")

	pkgB := filepath.Join(shared, "pkgB")
	if err := os.MkdirAll(pkgB, 0o755); err != nil {
		t.Fatalf("failed to create pkgB: %v", err)
	}

	tree, err := BuildRulesTree(shared)
	if err != nil {
		t.Fatalf("BuildRulesTree failed: %v", err)
	}

	if tree.SharedDir != shared {
		t.Fatalf("shared dir mismatch: got %s want %s", tree.SharedDir, shared)
	}

	if len(tree.Presets) != 2 || tree.Presets[0] != "readme.md" || tree.Presets[1] != "root.mdc" {
		t.Fatalf("unexpected presets: %+v", tree.Presets)
	}

	if len(tree.Packages) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(tree.Packages))
	}

	if tree.Packages[0].Name != "pkgA" {
		t.Fatalf("expected first package pkgA, got %s", tree.Packages[0].Name)
	}
	if len(tree.Packages[0].Files) != 2 {
		t.Fatalf("expected 2 files in pkgA, got %d (%+v)", len(tree.Packages[0].Files), tree.Packages[0].Files)
	}
	if tree.Packages[0].Files[0] != "file1.mdc" || tree.Packages[0].Files[1] != filepath.Join("nested", "inner.mdc") {
		t.Fatalf("unexpected files in pkgA: %+v", tree.Packages[0].Files)
	}

	if tree.Packages[1].Name != "pkgB" || len(tree.Packages[1].Files) != 0 {
		t.Fatalf("unexpected pkgB contents: %+v", tree.Packages[1])
	}

	got := FormatRulesTree(tree)
	expected := "shared dir: " + shared + `
presets:
  ├─ readme.md
  └─ root.mdc
packages:
  ├─ pkgA/
  │  ├─ file1.mdc
  │  └─ nested/inner.mdc
  └─ pkgB/
      (no rule files)`

	if got != expected {
		t.Fatalf("formatted tree mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, expected)
	}
}

func TestBuildRulesTreeMissingDir(t *testing.T) {
	tree, err := BuildRulesTree(filepath.Join(t.TempDir(), "missing"))
	if err != nil {
		t.Fatalf("expected graceful handling of missing dir, got err: %v", err)
	}
	if tree == nil {
		t.Fatal("expected tree, got nil")
	}
	if len(tree.Presets) != 0 || len(tree.Packages) != 0 {
		t.Fatalf("expected empty tree, got presets=%d packages=%d", len(tree.Presets), len(tree.Packages))
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("failed to create dir for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write file %s: %v", path, err)
	}
}
