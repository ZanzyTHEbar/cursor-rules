package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallAndRemoveCommand(t *testing.T) {
	// create temp shared dir
	sharedDir := t.TempDir()

	// create sample command file
	cmdName := "hello"
	cmdFile := filepath.Join(sharedDir, cmdName+".md")
	if err := os.WriteFile(cmdFile, []byte("# hello command\n"), 0o644); err != nil {
		t.Fatalf("failed to write command: %v", err)
	}

	// create temp project dir
	proj := t.TempDir()

	// override DefaultSharedDir via env so shared dir used is our temp sharedDir
	old := os.Getenv("CURSOR_RULES_DIR")
	os.Setenv("CURSOR_RULES_DIR", sharedDir)
	defer os.Setenv("CURSOR_RULES_DIR", old)

	// Test InstallCommand
	if err := InstallCommand(proj, cmdName); err != nil {
		t.Fatalf("InstallCommand failed: %v", err)
	}

	// verify stub exists
	stub := filepath.Join(proj, ".cursor", "commands", cmdName+".md")
	if _, err := os.Stat(stub); err != nil {
		t.Fatalf("expected stub file at %s, err: %v", stub, err)
	}

	// Test ListProjectCommands
	cmds, err := ListProjectCommands(proj)
	if err != nil {
		t.Fatalf("ListProjectCommands failed: %v", err)
	}
	found := false
	for _, c := range cmds {
		if c == cmdName+".md" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected command %s in list output", cmdName)
	}

	// Test RemoveCommand
	if err := RemoveCommand(proj, cmdName); err != nil {
		t.Fatalf("RemoveCommand failed: %v", err)
	}
	if _, err := os.Stat(stub); !os.IsNotExist(err) {
		t.Fatalf("expected stub removed, but still exists or different error: %v", err)
	}
}

func TestInstallCommandPackageFlattenIgnore(t *testing.T) {
	sharedDir := t.TempDir()
	pkg := filepath.Join(sharedDir, "pkg")
	if err := os.MkdirAll(pkg, 0o755); err != nil {
		t.Fatalf("mkdir pkg failed: %v", err)
	}
	// create files
	if err := os.WriteFile(filepath.Join(pkg, "a.md"), []byte("a"), 0o644); err != nil {
		t.Fatalf("write a.md: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(pkg, "templates"), 0o755); err != nil {
		t.Fatalf("mkdir templates: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pkg, "templates", "t.md"), []byte("t"), 0o644); err != nil {
		t.Fatalf("write t.md: %v", err)
	}
	// write ignore file to skip templates/*
	if err := os.WriteFile(filepath.Join(pkg, ".cursor-rules-ignore"), []byte("templates/*\n"), 0o644); err != nil {
		t.Fatalf("write ignore: %v", err)
	}

	// override DefaultSharedDir via env so shared dir used is our temp sharedDir
	old := os.Getenv("CURSOR_RULES_DIR")
	os.Setenv("CURSOR_RULES_DIR", sharedDir)
	defer os.Setenv("CURSOR_RULES_DIR", old)

	// target project
	proj := t.TempDir()
	// run install package with noFlatten=false (default flattening behavior)
	if err := InstallCommandPackage(proj, "pkg", nil, false); err != nil {
		t.Fatalf("InstallCommandPackage failed: %v", err)
	}

	// verify a.md exists at commands root
	if _, err := os.Stat(filepath.Join(proj, ".cursor", "commands", "a.md")); err != nil {
		t.Fatalf("expected a.md in commands root: %v", err)
	}
	// verify templates/t.md NOT copied
	if _, err := os.Stat(filepath.Join(proj, ".cursor", "commands", "t.md")); err == nil {
		t.Fatalf("expected templates t.md to be ignored when flattened")
	}
}
