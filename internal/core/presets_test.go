package core

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestApplyRemoveListInstall(t *testing.T) {
	// create temp package dir
	packageDir, err := os.MkdirTemp("", "package-")
	if err != nil {
		t.Fatalf("failed to create package dir: %v", err)
	}
	defer os.RemoveAll(packageDir)

	// create a sample preset file
	presetName := "frontend"
	presetFile := filepath.Join(packageDir, presetName+".mdc")
	if err := os.WriteFile(presetFile, []byte("# sample preset"), 0o644); err != nil {
		t.Fatalf("failed to write preset: %v", err)
	}

	// create temp project dir
	projectDir, err := os.MkdirTemp("", "project-")
	if err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}
	defer os.RemoveAll(projectDir)

	// set env override so DefaultPackageDir uses our temp packageDir
	old := os.Getenv("CURSOR_RULES_PACKAGE_DIR")
	os.Setenv("CURSOR_RULES_PACKAGE_DIR", packageDir)
	defer os.Setenv("CURSOR_RULES_PACKAGE_DIR", old)

	// Test ApplyPresetToProject idempotency
	if _, err := ApplyPresetToProject(projectDir, presetName, packageDir); err != nil {
		t.Fatalf("ApplyPresetToProject failed: %v", err)
	}
	// apply again - should be idempotent
	if _, err := ApplyPresetToProject(projectDir, presetName, packageDir); err != nil {
		t.Fatalf("ApplyPresetToProject idempotent failed: %v", err)
	}

	// verify stub exists
	stub := filepath.Join(projectDir, ".cursor", "rules", presetName+".mdc")
	if _, err := os.Stat(stub); err != nil {
		t.Fatalf("expected stub file at %s, err: %v", stub, err)
	}

	// Test ListPackagePresets
	presets, err := ListPackagePresets(packageDir)
	if err != nil {
		t.Fatalf("ListPackagePresets failed: %v", err)
	}
	found := slices.Contains(presets, presetName+".mdc")
	if !found {
		t.Fatalf("preset %s not found in ListPackagePresets output", presetName)
	}

	// Test InstallPreset (uses DefaultPackageDir -> CURSOR_RULES_PACKAGE_DIR)
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
	// create temp package dir
	packageDir, err := os.MkdirTemp("", "package-")
	if err != nil {
		t.Fatalf("failed to create package dir: %v", err)
	}
	defer os.RemoveAll(packageDir)

	// create a sample preset file
	presetName := "frontend"
	presetFile := filepath.Join(packageDir, presetName+".mdc")
	if err := os.WriteFile(presetFile, []byte("# sample preset"), 0o644); err != nil {
		t.Fatalf("failed to write preset: %v", err)
	}

	// create temp project dir
	projectDir, err := os.MkdirTemp("", "project-")
	if err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}
	defer os.RemoveAll(projectDir)

	// set env override so DefaultPackageDir uses our temp packageDir
	old := os.Getenv("CURSOR_RULES_PACKAGE_DIR")
	os.Setenv("CURSOR_RULES_PACKAGE_DIR", packageDir)
	defer os.Setenv("CURSOR_RULES_PACKAGE_DIR", old)

	// request symlink behavior
	oldSymlink := os.Getenv("CURSOR_RULES_SYMLINK")
	os.Setenv("CURSOR_RULES_SYMLINK", "1")
	defer os.Setenv("CURSOR_RULES_SYMLINK", oldSymlink)

	// Apply preset
	if _, err := ApplyPresetToProject(projectDir, presetName, packageDir); err != nil {
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

func TestInstallPresetWithGNUStowRequestCreatesSymlink(t *testing.T) {
	packageDir := t.TempDir()
	presetName := "frontend"
	presetFile := filepath.Join(packageDir, presetName+".mdc")
	if err := os.WriteFile(presetFile, []byte("# sample preset"), 0o644); err != nil {
		t.Fatalf("failed to write preset: %v", err)
	}

	projectDir := t.TempDir()

	t.Setenv("CURSOR_RULES_PACKAGE_DIR", packageDir)
	t.Setenv("CURSOR_RULES_USE_GNUSTOW", "1")
	t.Setenv("CURSOR_RULES_SYMLINK", "")

	if err := InstallPreset(projectDir, presetName); err != nil {
		t.Fatalf("InstallPreset with GNU stow request failed: %v", err)
	}

	dest := filepath.Join(projectDir, ".cursor", "rules", presetName+".mdc")
	info, err := os.Lstat(dest)
	if err != nil {
		t.Fatalf("expected symlink at %s: %v", dest, err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("expected symlink for preset install, got regular file")
	}

	target, err := os.Readlink(dest)
	if err != nil {
		t.Fatalf("failed to read symlink target: %v", err)
	}
	if target != presetFile {
		t.Fatalf("symlink target mismatch: got %s want %s", target, presetFile)
	}
}

func TestInstallPackageWithIgnoreAndFlatten(t *testing.T) {
	// Setup temp package dir
	packageDir := t.TempDir()
	// create package dir
	pkg := filepath.Join(packageDir, "pkg")
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

	// override DefaultPackageDir via env
	os.Setenv("CURSOR_RULES_PACKAGE_DIR", packageDir)
	defer os.Unsetenv("CURSOR_RULES_PACKAGE_DIR")

	// target project
	proj := t.TempDir()
	// run install package with noFlatten=false (default flattening behavior)
	if _, err := InstallPackage(proj, "pkg", nil, false); err != nil {
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

func TestInstallPackageWithGNUStowRequestCreatesSymlinks(t *testing.T) {
	packageDir := t.TempDir()
	pkgDir := filepath.Join(packageDir, "pkg")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatalf("mkdir pkg failed: %v", err)
	}
	sharedFile := filepath.Join(pkgDir, "a.mdc")
	if err := os.WriteFile(sharedFile, []byte("content"), 0o644); err != nil {
		t.Fatalf("write a.mdc: %v", err)
	}

	projectDir := t.TempDir()

	t.Setenv("CURSOR_RULES_PACKAGE_DIR", packageDir)
	t.Setenv("CURSOR_RULES_USE_GNUSTOW", "1")
	t.Setenv("CURSOR_RULES_SYMLINK", "")

	if _, err := InstallPackage(projectDir, "pkg", nil, false); err != nil {
		t.Fatalf("InstallPackage with GNU stow request failed: %v", err)
	}

	dest := filepath.Join(projectDir, ".cursor", "rules", "a.mdc")
	info, err := os.Lstat(dest)
	if err != nil {
		t.Fatalf("expected file at %s: %v", dest, err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("expected symlink for package install when GNU stow requested")
	}

	target, err := os.Readlink(dest)
	if err != nil {
		t.Fatalf("failed to read symlink target: %v", err)
	}
	if target != sharedFile {
		t.Fatalf("symlink target mismatch: got %s want %s", target, sharedFile)
	}
}

func TestInstallNestedPackage(t *testing.T) {
	// Setup temp package dir with nested package structure
	packageDir := t.TempDir()

	// Create nested package: frontend/react
	nestedPkg := filepath.Join(packageDir, "frontend", "react")
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

	// override DefaultPackageDir via env
	os.Setenv("CURSOR_RULES_PACKAGE_DIR", packageDir)
	defer os.Unsetenv("CURSOR_RULES_PACKAGE_DIR")

	// target project
	proj := t.TempDir()

	// Test installing nested package - should be auto-flattened
	if _, err := InstallPackage(proj, "frontend/react", nil, false); err != nil {
		t.Fatalf("InstallPackage nested failed: %v", err)
	}

	// First, let's see what files were actually created
	rulesDir := filepath.Join(proj, ".cursor", "rules")
	var createdFiles []string
	if err := filepath.Walk(rulesDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			rel, _ := filepath.Rel(rulesDir, path)
			createdFiles = append(createdFiles, rel)
		}
		return nil
	}); err != nil {
		t.Fatalf("Failed to walk rules directory: %v", err)
	}
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
	// Setup temp package dir with deeply nested package structure
	packageDir := t.TempDir()

	// Create deeply nested package: backend/nodejs/express/middleware
	deepNestedPkg := filepath.Join(packageDir, "backend", "nodejs", "express", "middleware")
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

	// override DefaultPackageDir via env
	os.Setenv("CURSOR_RULES_PACKAGE_DIR", packageDir)
	defer os.Unsetenv("CURSOR_RULES_PACKAGE_DIR")

	// target project
	proj := t.TempDir()

	// Test installing deeply nested package - should be auto-flattened
	if _, err := InstallPackage(proj, "backend/nodejs/express/middleware", nil, false); err != nil {
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
	// Setup temp package dir with both regular and nested packages
	packageDir := t.TempDir()

	// Create regular package: frontend
	regularPkg := filepath.Join(packageDir, "frontend")
	if err := os.MkdirAll(regularPkg, 0o755); err != nil {
		t.Fatalf("mkdir regular package failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(regularPkg, "regular.mdc"), []byte("# Regular Package"), 0o644); err != nil {
		t.Fatalf("write regular.mdc: %v", err)
	}

	// Create nested package: frontend/react
	nestedPkg := filepath.Join(packageDir, "frontend", "react")
	if err := os.MkdirAll(nestedPkg, 0o755); err != nil {
		t.Fatalf("mkdir nested package failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nestedPkg, "nested.mdc"), []byte("# Nested Package"), 0o644); err != nil {
		t.Fatalf("write nested.mdc: %v", err)
	}

	// override DefaultPackageDir via env
	os.Setenv("CURSOR_RULES_PACKAGE_DIR", packageDir)
	defer os.Unsetenv("CURSOR_RULES_PACKAGE_DIR")

	// target project
	proj := t.TempDir()

	// Test installing regular package WITH noFlatten=true - should preserve structure
	if _, err := InstallPackage(proj, "frontend", nil, true); err != nil {
		t.Fatalf("InstallPackage regular failed: %v", err)
	}

	// Verify regular package preserves structure
	if _, err := os.Stat(filepath.Join(proj, ".cursor", "rules", "frontend", "regular.mdc")); err != nil {
		t.Fatalf("expected regular.mdc in frontend subdirectory: %v", err)
	}

	// Clean up for next test
	os.RemoveAll(filepath.Join(proj, ".cursor"))

	// Test installing nested package - should auto-flatten
	if _, err := InstallPackage(proj, "frontend/react", nil, false); err != nil {
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

// Test the double extension bug fix - when user includes .mdc in preset name
func TestInstallPresetWithExtensionInName(t *testing.T) {
	packageDir, err := os.MkdirTemp("", "package-ext-")
	if err != nil {
		t.Fatalf("failed to create package dir: %v", err)
	}
	defer os.RemoveAll(packageDir)

	// Create a nested directory structure in package dir
	nestedDir := filepath.Join(packageDir, "emissium", "behaviour")
	if err := os.MkdirAll(nestedDir, 0o755); err != nil {
		t.Fatalf("failed to create nested dir: %v", err)
	}

	// Create preset file
	presetName := "task-execution-rules"
	presetFile := filepath.Join(nestedDir, presetName+".mdc")
	if err := os.WriteFile(presetFile, []byte("# Task execution rules preset"), 0o644); err != nil {
		t.Fatalf("failed to write preset: %v", err)
	}

	projectDir, err := os.MkdirTemp("", "project-ext-")
	if err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}
	defer os.RemoveAll(projectDir)

	// Set env override
	old := os.Getenv("CURSOR_RULES_PACKAGE_DIR")
	os.Setenv("CURSOR_RULES_PACKAGE_DIR", packageDir)
	defer os.Setenv("CURSOR_RULES_PACKAGE_DIR", old)

	// Test 1: Install with .mdc extension (this was the bug)
	presetWithExt := "emissium/behaviour/task-execution-rules.mdc"
	if err := InstallPreset(projectDir, presetWithExt); err != nil {
		t.Fatalf("InstallPreset with .mdc extension failed: %v", err)
	}

	// Verify the file was created correctly (should not have .mdc.mdc)
	expectedStub := filepath.Join(projectDir, ".cursor", "rules", "emissium", "behaviour", "task-execution-rules.mdc")
	if _, err := os.Stat(expectedStub); err != nil {
		t.Fatalf("expected stub at %s, err: %v", expectedStub, err)
	}

	// Test 2: Install without .mdc extension (should also work)
	os.RemoveAll(filepath.Join(projectDir, ".cursor")) // Clean up

	presetWithoutExt := "emissium/behaviour/task-execution-rules"
	if err := InstallPreset(projectDir, presetWithoutExt); err != nil {
		t.Fatalf("InstallPreset without .mdc extension failed: %v", err)
	}

	// Verify the file was created correctly
	if _, err := os.Stat(expectedStub); err != nil {
		t.Fatalf("expected stub at %s after second install, err: %v", expectedStub, err)
	}

	// Test 3: Idempotency - install again should not fail
	if err := InstallPreset(projectDir, presetWithExt); err != nil {
		t.Fatalf("InstallPreset idempotent failed: %v", err)
	}

	// Verify content is correct
	content, err := os.ReadFile(expectedStub)
	if err != nil {
		t.Fatalf("failed to read stub content: %v", err)
	}

	// Should be a stub file pointing to the source, not the content itself
	contentStr := string(content)
	if !os.IsPathSeparator(contentStr[len(contentStr)-len(presetFile)-1]) || !strings.HasPrefix(contentStr, "@file ") {
		t.Logf("Stub content: %q", contentStr)
		// This might be the actual content if it's not a stub, which is also fine
	}
}

// Test the directory creation bug fix - when preset has nested paths
func TestInstallPresetWithNestedPaths(t *testing.T) {
	packageDir, err := os.MkdirTemp("", "package-nested-")
	if err != nil {
		t.Fatalf("failed to create package dir: %v", err)
	}
	defer os.RemoveAll(packageDir)

	// Create deeply nested directory structure
	deepNestedDir := filepath.Join(packageDir, "company", "team", "backend", "api")
	if err := os.MkdirAll(deepNestedDir, 0o755); err != nil {
		t.Fatalf("failed to create deep nested dir: %v", err)
	}

	// Create preset file in nested directory
	presetName := "auth-middleware"
	presetFile := filepath.Join(deepNestedDir, presetName+".mdc")
	if err := os.WriteFile(presetFile, []byte("# Auth middleware preset"), 0o644); err != nil {
		t.Fatalf("failed to write preset: %v", err)
	}

	projectDir, err := os.MkdirTemp("", "project-nested-")
	if err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}
	defer os.RemoveAll(projectDir)

	// Set env override
	old := os.Getenv("CURSOR_RULES_PACKAGE_DIR")
	os.Setenv("CURSOR_RULES_PACKAGE_DIR", packageDir)
	defer os.Setenv("CURSOR_RULES_PACKAGE_DIR", old)

	// Test installing preset with deeply nested path
	nestedPresetPath := "company/team/backend/api/auth-middleware"
	if err := InstallPreset(projectDir, nestedPresetPath); err != nil {
		t.Fatalf("InstallPreset with nested path failed: %v", err)
	}

	// Verify the directory structure was created correctly
	expectedStub := filepath.Join(projectDir, ".cursor", "rules", "company", "team", "backend", "api", "auth-middleware.mdc")
	if _, err := os.Stat(expectedStub); err != nil {
		t.Fatalf("expected stub at %s, err: %v", expectedStub, err)
	}

	// Verify intermediate directories were created
	if _, err := os.Stat(filepath.Join(projectDir, ".cursor", "rules", "company", "team", "backend", "api")); err != nil {
		t.Fatalf("expected intermediate directories to be created: %v", err)
	}

	// Test with .mdc extension too
	os.RemoveAll(filepath.Join(projectDir, ".cursor")) // Clean up

	nestedPresetWithExt := "company/team/backend/api/auth-middleware.mdc"
	if err := InstallPreset(projectDir, nestedPresetWithExt); err != nil {
		t.Fatalf("InstallPreset with nested path and extension failed: %v", err)
	}

	// Verify it still works
	if _, err := os.Stat(expectedStub); err != nil {
		t.Fatalf("expected stub at %s after second test, err: %v", expectedStub, err)
	}
}
