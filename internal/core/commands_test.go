package core

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestInstallAndRemoveCommand(t *testing.T) {
	// create temp command dir
	sharedDir := t.TempDir()

	// create sample command file
	cmdName := "hello"
	cmdFile := filepath.Join(sharedDir, cmdName+".md")
	if err := os.WriteFile(cmdFile, []byte("# hello command\n"), 0o644); err != nil {
		t.Fatalf("failed to write command: %v", err)
	}

	// create temp project dir
	proj := t.TempDir()

	// override DefaultPackageDir via env so package dir used is our temp dir
	old := os.Getenv("CURSOR_RULES_PACKAGE_DIR")
	os.Setenv("CURSOR_RULES_PACKAGE_DIR", sharedDir)
	defer os.Setenv("CURSOR_RULES_PACKAGE_DIR", old)

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
	commandsDir := filepath.Join(proj, ".cursor", "commands")
	if err := RemoveCommand(commandsDir, cmdName); err != nil {
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
	if err := os.WriteFile(filepath.Join(pkg, ".cursor-commands-ignore"), []byte("templates/*\n"), 0o644); err != nil {
		t.Fatalf("write ignore: %v", err)
	}

	// override DefaultSharedCommandsDir via env
	old := os.Getenv("CURSOR_COMMANDS_DIR")
	os.Setenv("CURSOR_COMMANDS_DIR", sharedDir)
	defer os.Setenv("CURSOR_COMMANDS_DIR", old)

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

func TestListPackageCommandsIncludesCommandDirectories(t *testing.T) {
	packageDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(packageDir, "hello.md"), []byte("# hello"), 0o644); err != nil {
		t.Fatalf("write hello.md: %v", err)
	}

	commandDir := filepath.Join(packageDir, "commands", "release")
	if err := os.MkdirAll(commandDir, 0o755); err != nil {
		t.Fatalf("mkdir release command: %v", err)
	}
	if err := os.WriteFile(filepath.Join(commandDir, "run.md"), []byte("# run"), 0o644); err != nil {
		t.Fatalf("write run.md: %v", err)
	}

	names, err := ListPackageCommands(packageDir)
	if err != nil {
		t.Fatalf("ListPackageCommands: %v", err)
	}
	if !slices.Contains(names, "hello.md") {
		t.Fatalf("expected hello.md in command list: %v", names)
	}
	if !slices.Contains(names, "release") {
		t.Fatalf("expected release command directory in command list: %v", names)
	}
}

func TestInstallAndRemoveCommandDirectory(t *testing.T) {
	sharedDir := t.TempDir()
	commandDir := filepath.Join(sharedDir, "commands", "release")
	if err := os.MkdirAll(filepath.Join(commandDir, "partials"), 0o755); err != nil {
		t.Fatalf("mkdir partials: %v", err)
	}
	if err := os.WriteFile(filepath.Join(commandDir, "run.md"), []byte("# run"), 0o644); err != nil {
		t.Fatalf("write run.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(commandDir, "partials", "details.md"), []byte("# details"), 0o644); err != nil {
		t.Fatalf("write details.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(commandDir, "notes.txt"), []byte("ignore"), 0o644); err != nil {
		t.Fatalf("write notes.txt: %v", err)
	}
	if err := os.WriteFile(filepath.Join(commandDir, ".cursor-commands-ignore"), []byte("partials/*\n"), 0o644); err != nil {
		t.Fatalf("write ignore file: %v", err)
	}

	proj := t.TempDir()
	strategy, err := InstallCommandToProject(proj, sharedDir, "release", nil, false)
	if err != nil {
		t.Fatalf("InstallCommandToProject failed: %v", err)
	}
	if strategy != StrategyCopy && strategy != StrategySymlink && strategy != StrategyStow {
		t.Fatalf("unexpected install strategy: %s", strategy)
	}

	runPath := filepath.Join(proj, ".cursor", "commands", "release", "run.md")
	if _, err := os.Stat(runPath); err != nil {
		t.Fatalf("expected run.md at %s: %v", runPath, err)
	}
	if _, err := os.Stat(filepath.Join(proj, ".cursor", "commands", "release", "partials", "details.md")); err == nil {
		t.Fatalf("expected ignored markdown file to be skipped")
	}
	if _, err := os.Stat(filepath.Join(proj, ".cursor", "commands", "release", "notes.txt")); err == nil {
		t.Fatalf("expected non-markdown file to be skipped")
	}

	cmds, err := ListProjectCommands(proj)
	if err != nil {
		t.Fatalf("ListProjectCommands failed: %v", err)
	}
	if !slices.Contains(cmds, "release") {
		t.Fatalf("expected release directory in project command list: %v", cmds)
	}

	commandsDir := filepath.Join(proj, ".cursor", "commands")
	if err := RemoveCommand(commandsDir, "release"); err != nil {
		t.Fatalf("RemoveCommand failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(proj, ".cursor", "commands", "release")); !os.IsNotExist(err) {
		t.Fatalf("expected command directory removed, got: %v", err)
	}
}
