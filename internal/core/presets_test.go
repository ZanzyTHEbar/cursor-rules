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

func TestInstallNestedPackage(t *testing.T) {
	// Setup temp shared dir with nested package structure
	sharedDir := t.TempDir()
	
	// Create nested package: frontend/react
	nestedPkg := filepath.Join(sharedDir, "frontend", "react")
	if err := os.MkdirAll(nestedPkg, 0o755); err != nil {
		t.Fatalf("mkdir nested package failed: %v", err)
	}
	
	// Create files in nested package
	if err := os.WriteFile(filepath.Join(nestedPkg, "hooks.mdc"), []byte("# React Hooks"), 0o644); err != nil {
		t.Fatalf("write hooks.mdc: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nestedPkg, "components.mdc"), []byte("# React Components"), 0o644); err != nil {
		t.Fatalf("write components.mdc: %v", err)
	}
	
	// Create subdirectory within the package
	subDir := filepath.Join(nestedPkg, "advanced")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("mkdir advanced subdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "patterns.mdc"), []byte("# Advanced Patterns"), 0o644); err != nil {
		t.Fatalf("write patterns.mdc: %v", err)
	}

	// override DefaultSharedDir via env
	os.Setenv("CURSOR_RULES_DIR", sharedDir)
	defer os.Unsetenv("CURSOR_RULES_DIR")

	// target project
	proj := t.TempDir()
	
	// Test installing nested package - should be auto-flattened
	if err := InstallPackage(proj, "frontend/react", nil, false); err != nil {
		t.Fatalf("InstallPackage nested failed: %v", err)
	}

	// First, let's see what files were actually created
	rulesDir := filepath.Join(proj, ".cursor", "rules")
	var createdFiles []string
	filepath.Walk(rulesDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			rel, _ := filepath.Rel(rulesDir, path)
			createdFiles = append(createdFiles, rel)
		}
		return nil
	})
	t.Logf("Created files: %v", createdFiles)
	
	// Verify files are flattened to rules root (nested packages should always flatten)
	if _, err := os.Stat(filepath.Join(proj, ".cursor", "rules", "hooks.mdc")); err != nil {
		t.Fatalf("expected hooks.mdc in rules root: %v", err)
	}
	if _, err := os.Stat(filepath.Join(proj, ".cursor", "rules", "components.mdc")); err != nil {
		t.Fatalf("expected components.mdc in rules root: %v", err)
	}
	if _, err := os.Stat(filepath.Join(proj, ".cursor", "rules", "patterns.mdc")); err != nil {
		t.Fatalf("expected patterns.mdc in rules root: %v", err)
	}
	
	// Verify that nested structure is NOT preserved (should be flattened)
	if _, err := os.Stat(filepath.Join(proj, ".cursor", "rules", "frontend", "react", "hooks.mdc")); err == nil {
		t.Fatalf("expected nested structure to be flattened, but found file at nested path")
	}
}

func TestInstallNestedPackageDeepNesting(t *testing.T) {
	// Setup temp shared dir with deeply nested package structure
	sharedDir := t.TempDir()
	
	// Create deeply nested package: backend/nodejs/express/middleware
	deepNestedPkg := filepath.Join(sharedDir, "backend", "nodejs", "express", "middleware")
	if err := os.MkdirAll(deepNestedPkg, 0o755); err != nil {
		t.Fatalf("mkdir deep nested package failed: %v", err)
	}
	
	// Create files in deeply nested package
	if err := os.WriteFile(filepath.Join(deepNestedPkg, "auth.mdc"), []byte("# Auth Middleware"), 0o644); err != nil {
		t.Fatalf("write auth.mdc: %v", err)
	}
	if err := os.WriteFile(filepath.Join(deepNestedPkg, "logging.mdc"), []byte("# Logging Middleware"), 0o644); err != nil {
		t.Fatalf("write logging.mdc: %v", err)
	}

	// override DefaultSharedDir via env
	os.Setenv("CURSOR_RULES_DIR", sharedDir)
	defer os.Unsetenv("CURSOR_RULES_DIR")

	// target project
	proj := t.TempDir()
	
	// Test installing deeply nested package - should be auto-flattened
	if err := InstallPackage(proj, "backend/nodejs/express/middleware", nil, false); err != nil {
		t.Fatalf("InstallPackage deeply nested failed: %v", err)
	}

	// Verify files are flattened to rules root
	if _, err := os.Stat(filepath.Join(proj, ".cursor", "rules", "auth.mdc")); err != nil {
		t.Fatalf("expected auth.mdc in rules root: %v", err)
	}
	if _, err := os.Stat(filepath.Join(proj, ".cursor", "rules", "logging.mdc")); err != nil {
		t.Fatalf("expected logging.mdc in rules root: %v", err)
	}
}

func TestInstallPackageRegularVsNested(t *testing.T) {
	// Setup temp shared dir with both regular and nested packages
	sharedDir := t.TempDir()
	
	// Create regular package: frontend
	regularPkg := filepath.Join(sharedDir, "frontend")
	if err := os.MkdirAll(regularPkg, 0o755); err != nil {
		t.Fatalf("mkdir regular package failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(regularPkg, "regular.mdc"), []byte("# Regular Package"), 0o644); err != nil {
		t.Fatalf("write regular.mdc: %v", err)
	}
	
	// Create nested package: frontend/react  
	nestedPkg := filepath.Join(sharedDir, "frontend", "react")
	if err := os.MkdirAll(nestedPkg, 0o755); err != nil {
		t.Fatalf("mkdir nested package failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nestedPkg, "nested.mdc"), []byte("# Nested Package"), 0o644); err != nil {
		t.Fatalf("write nested.mdc: %v", err)
	}

	// override DefaultSharedDir via env
	os.Setenv("CURSOR_RULES_DIR", sharedDir)
	defer os.Unsetenv("CURSOR_RULES_DIR")

	// target project
	proj := t.TempDir()
	
	// Test installing regular package WITHOUT flatten - should preserve structure
	if err := InstallPackage(proj, "frontend", nil, false); err != nil {
		t.Fatalf("InstallPackage regular failed: %v", err)
	}
	
	// Verify regular package preserves structure
	if _, err := os.Stat(filepath.Join(proj, ".cursor", "rules", "frontend", "regular.mdc")); err != nil {
		t.Fatalf("expected regular.mdc in frontend subdirectory: %v", err)
	}
	
	// Clean up for next test
	os.RemoveAll(filepath.Join(proj, ".cursor"))
	
	// Test installing nested package - should auto-flatten
	if err := InstallPackage(proj, "frontend/react", nil, false); err != nil {
		t.Fatalf("InstallPackage nested failed: %v", err)
	}
	
	// Verify nested package is flattened
	if _, err := os.Stat(filepath.Join(proj, ".cursor", "rules", "nested.mdc")); err != nil {
		t.Fatalf("expected nested.mdc in rules root: %v", err)
	}
	
	// Verify it's NOT in nested structure
	if _, err := os.Stat(filepath.Join(proj, ".cursor", "rules", "frontend", "react", "nested.mdc")); err == nil {
		t.Fatalf("expected nested package to be flattened, but found file at nested path")
	}
}
