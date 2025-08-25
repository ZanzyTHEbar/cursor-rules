package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApplyRemoveListInstall(t *testing.T) {
	// create temp shared dir
	sharedDir, err := os.MkdirTemp("", "shared-")
	if err != nil {
		t.Fatalf("failed to create shared dir: %v", err)
	}
	defer os.RemoveAll(sharedDir)

	// create a sample preset file
	presetName := "frontend"
	presetFile := filepath.Join(sharedDir, presetName+".mdc")
	if err := os.WriteFile(presetFile, []byte("# sample preset"), 0o644); err != nil {
		t.Fatalf("failed to write preset: %v", err)
	}

	// create temp project dir
	projectDir, err := os.MkdirTemp("", "project-")
	if err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}
	defer os.RemoveAll(projectDir)

	// set env override so DefaultSharedDir uses our temp sharedDir
	old := os.Getenv("CURSOR_RULES_DIR")
	os.Setenv("CURSOR_RULES_DIR", sharedDir)
	defer os.Setenv("CURSOR_RULES_DIR", old)

	// Test ApplyPresetToProject idempotency
	if err := ApplyPresetToProject(projectDir, presetName, sharedDir); err != nil {
		t.Fatalf("ApplyPresetToProject failed: %v", err)
	}
	// apply again - should be idempotent
	if err := ApplyPresetToProject(projectDir, presetName, sharedDir); err != nil {
		t.Fatalf("ApplyPresetToProject idempotent failed: %v", err)
	}

	// verify stub exists
	stub := filepath.Join(projectDir, ".cursor", "rules", presetName+".mdc")
	if _, err := os.Stat(stub); err != nil {
		t.Fatalf("expected stub file at %s, err: %v", stub, err)
	}

	// Test ListSharedPresets
	presets, err := ListSharedPresets(sharedDir)
	if err != nil {
		t.Fatalf("ListSharedPresets failed: %v", err)
	}
	found := false
	for _, p := range presets {
		if p == presetName+".mdc" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("preset %s not found in ListSharedPresets output", presetName)
	}

	// Test InstallPreset (uses DefaultSharedDir -> CURSOR_RULES_DIR)
	if err := InstallPreset(projectDir, presetName); err != nil {
		t.Fatalf("InstallPreset failed: %v", err)
	}
	// Install should have created a stub as well
	if _, err := os.Stat(stub); err != nil {
		t.Fatalf("expected stub after InstallPreset at %s, err: %v", stub, err)
	}

	// Test RemovePreset
	if err := RemovePreset(projectDir, presetName); err != nil {
		t.Fatalf("RemovePreset failed: %v", err)
	}
	if _, err := os.Stat(stub); !os.IsNotExist(err) {
		t.Fatalf("expected stub removed, but still exists or different error: %v", err)
	}
}

func TestApplyWithSymlinkPreference(t *testing.T) {
	// create temp shared dir
	sharedDir, err := os.MkdirTemp("", "shared-")
	if err != nil {
		t.Fatalf("failed to create shared dir: %v", err)
	}
	defer os.RemoveAll(sharedDir)

	// create a sample preset file
	presetName := "frontend"
	presetFile := filepath.Join(sharedDir, presetName+".mdc")
	if err := os.WriteFile(presetFile, []byte("# sample preset"), 0o644); err != nil {
		t.Fatalf("failed to write preset: %v", err)
	}

	// create temp project dir
	projectDir, err := os.MkdirTemp("", "project-")
	if err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}
	defer os.RemoveAll(projectDir)

	// set env override so DefaultSharedDir uses our temp sharedDir
	old := os.Getenv("CURSOR_RULES_DIR")
	os.Setenv("CURSOR_RULES_DIR", sharedDir)
	defer os.Setenv("CURSOR_RULES_DIR", old)

	// request symlink behavior
	oldSymlink := os.Getenv("CURSOR_RULES_SYMLINK")
	os.Setenv("CURSOR_RULES_SYMLINK", "1")
	defer os.Setenv("CURSOR_RULES_SYMLINK", oldSymlink)

	// Apply preset
	if err := ApplyPresetToProject(projectDir, presetName, sharedDir); err != nil {
		t.Fatalf("ApplyPresetToProject with symlink failed: %v", err)
	}

	// verify symlink exists
	stub := filepath.Join(projectDir, ".cursor", "rules", presetName+".mdc")
	info, err := os.Lstat(stub)
	if err != nil {
		t.Fatalf("expected symlink at %s, err: %v", stub, err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("expected symlink, but file is not a symlink: %s", stub)
	}

	// verify symlink target
	target, err := os.Readlink(stub)
	if err != nil {
		t.Fatalf("failed to read symlink: %v", err)
	}
	if target != presetFile {
		t.Fatalf("symlink target mismatch: got %s want %s", target, presetFile)
	}
}

func TestInstallPackageWithIgnoreAndFlatten(t *testing.T) {
	// Setup temp shared dir
	sharedDir := t.TempDir()
	// create package dir
	pkg := filepath.Join(sharedDir, "pkg")
	if err := os.MkdirAll(pkg, 0o755); err != nil {
		t.Fatalf("mkdir pkg failed: %v", err)
	}
	// create files
	if err := os.WriteFile(filepath.Join(pkg, "a.mdc"), []byte("a"), 0o644); err != nil {
		t.Fatalf("write a.mdc: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(pkg, "templates"), 0o755); err != nil {
		t.Fatalf("mkdir templates: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pkg, "templates", "t.mdc"), []byte("t"), 0o644); err != nil {
		t.Fatalf("write t.mdc: %v", err)
	}
	// write ignore file to skip templates/*
	if err := os.WriteFile(filepath.Join(pkg, ".cursor-rules-ignore"), []byte("templates/*\n"), 0o644); err != nil {
		t.Fatalf("write ignore: %v", err)
	}

	// override DefaultSharedDir via env
	os.Setenv("CURSOR_RULES_DIR", sharedDir)
	defer os.Unsetenv("CURSOR_RULES_DIR")

	// target project
	proj := t.TempDir()
	// run install package with flatten=true
	if err := InstallPackage(proj, "pkg", nil, true); err != nil {
		t.Fatalf("InstallPackage failed: %v", err)
	}

	// verify a.mdc exists at rules root
	if _, err := os.Stat(filepath.Join(proj, ".cursor", "rules", "a.mdc")); err != nil {
		t.Fatalf("expected a.mdc in rules root: %v", err)
	}
	// verify templates/t.mdc NOT copied
	if _, err := os.Stat(filepath.Join(proj, ".cursor", "rules", "t.mdc")); err == nil {
		t.Fatalf("expected templates t.mdc to be ignored when flattened")
	}
}
