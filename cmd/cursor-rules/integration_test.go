package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ZanzyTHEbar/cursor-rules/cli"
	"github.com/ZanzyTHEbar/cursor-rules/cmd/cursor-rules/commands"
)

// TestInstallCommandIntegration tests the install command end-to-end
func TestInstallCommandIntegration(t *testing.T) {
	// Setup test environment
	tmpShared := t.TempDir()
	tmpProject := t.TempDir()

	// Create test rule
	testRule := `---
description: "Test rule"
apply_to: "**/*.ts"
priority: 1
---
Use strict mode.`

	ruleFile := filepath.Join(tmpShared, "test.mdc")
	if err := os.WriteFile(ruleFile, []byte(testRule), 0644); err != nil {
		t.Fatalf("Failed to create test rule: %v", err)
	}

	// Set environment
	os.Setenv("CURSOR_RULES_DIR", tmpShared)
	defer os.Unsetenv("CURSOR_RULES_DIR")

	tests := []struct {
		name       string
		target     string
		outputFile string
		checkFunc  func(t *testing.T, path string)
	}{
		{
			name:       "install to cursor",
			target:     "cursor",
			outputFile: filepath.Join(tmpProject, ".cursor/rules/test.mdc"),
			checkFunc: func(t *testing.T, path string) {
				if _, err := os.Stat(path); os.IsNotExist(err) {
					t.Errorf("Output file not created: %s", path)
				}
				content, _ := os.ReadFile(path)
				if len(content) == 0 {
					t.Error("Output file is empty")
				}
			},
		},
		{
			name:       "install to copilot-instr",
			target:     "copilot-instr",
			outputFile: filepath.Join(tmpProject, ".github/instructions/test.instructions.md"),
			checkFunc: func(t *testing.T, path string) {
				if _, err := os.Stat(path); os.IsNotExist(err) {
					t.Errorf("Output file not created: %s", path)
				}
				content, _ := os.ReadFile(path)
				contentStr := string(content)
				// Check transformation
				if !contains(contentStr, "applyTo:") {
					t.Error("applyTo field not found in output")
				}
				if contains(contentStr, "priority:") {
					t.Error("priority field should be removed")
				}
			},
		},
		{
			name:       "install to copilot-prompt",
			target:     "copilot-prompt",
			outputFile: filepath.Join(tmpProject, ".github/prompts/test.prompt.md"),
			checkFunc: func(t *testing.T, path string) {
				if _, err := os.Stat(path); os.IsNotExist(err) {
					t.Errorf("Output file not created: %s", path)
				}
				content, _ := os.ReadFile(path)
				contentStr := string(content)
				// Check transformation
				if !contains(contentStr, "mode:") {
					t.Error("mode field not found in output")
				}
				if contains(contentStr, "applyTo:") {
					t.Error("applyTo field should be removed for prompts")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create command context
			ctx := cli.NewAppContext(nil, nil)
			ctx.Viper.Set("workdir", tmpProject)

			// Create and execute command
			cmd := commands.NewInstallCmd(ctx)
			cmd.SetArgs([]string{"test", "--target", tt.target})

			if err := cmd.Execute(); err != nil {
				t.Fatalf("Command execution failed: %v", err)
			}

			// Run checks
			tt.checkFunc(t, tt.outputFile)
		})
	}
}

// TestInstallIdempotency tests that repeated installs don't modify files
func TestInstallIdempotency(t *testing.T) {
	tmpShared := t.TempDir()
	tmpProject := t.TempDir()

	// Create test rule
	testRule := `---
description: "Test rule"
apply_to: "**/*.ts"
---
Content`

	ruleFile := filepath.Join(tmpShared, "test.mdc")
	if err := os.WriteFile(ruleFile, []byte(testRule), 0644); err != nil {
		t.Fatalf("Failed to create test rule: %v", err)
	}

	os.Setenv("CURSOR_RULES_DIR", tmpShared)
	defer os.Unsetenv("CURSOR_RULES_DIR")

	ctx := cli.NewAppContext(nil, nil)
	ctx.Viper.Set("workdir", tmpProject)

	// First install
	cmd1 := commands.NewInstallCmd(ctx)
	cmd1.SetArgs([]string{"test", "--target", "copilot-instr"})
	if err := cmd1.Execute(); err != nil {
		t.Fatalf("First install failed: %v", err)
	}

	outputFile := filepath.Join(tmpProject, ".github/instructions/test.instructions.md")
	content1, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	info1, _ := os.Stat(outputFile)
	mtime1 := info1.ModTime()

	// Second install (should be idempotent)
	cmd2 := commands.NewInstallCmd(ctx)
	cmd2.SetArgs([]string{"test", "--target", "copilot-instr"})
	if err := cmd2.Execute(); err != nil {
		t.Fatalf("Second install failed: %v", err)
	}

	content2, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file after second install: %v", err)
	}

	// Content should be identical
	if string(content1) != string(content2) {
		t.Error("Content changed after second install (not idempotent)")
	}

	// Modification time should be unchanged (file not rewritten)
	info2, _ := os.Stat(outputFile)
	mtime2 := info2.ModTime()

	if !mtime1.Equal(mtime2) {
		t.Error("File was rewritten (modification time changed)")
	}
}

// TestPackageInstallation tests installing a package directory
func TestPackageInstallation(t *testing.T) {
	tmpShared := t.TempDir()
	tmpProject := t.TempDir()

	// Create package directory
	pkgDir := filepath.Join(tmpShared, "testpkg")
	os.MkdirAll(pkgDir, 0755)

	// Create multiple rules
	rules := map[string]string{
		"rule1.mdc": `---
description: "Rule 1"
apply_to: "**/*.ts"
---
Content 1`,
		"rule2.mdc": `---
description: "Rule 2"
apply_to: "**/*.js"
---
Content 2`,
	}

	for name, content := range rules {
		path := filepath.Join(pkgDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create rule %s: %v", name, err)
		}
	}

	os.Setenv("CURSOR_RULES_DIR", tmpShared)
	defer os.Unsetenv("CURSOR_RULES_DIR")

	ctx := cli.NewAppContext(nil, nil)
	ctx.Viper.Set("workdir", tmpProject)

	// Install package
	cmd := commands.NewInstallCmd(ctx)
	cmd.SetArgs([]string{"testpkg", "--target", "copilot-instr"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Package install failed: %v", err)
	}

	// Verify all files installed
	for name := range rules {
		outputName := filepath.Base(name)
		outputName = outputName[:len(outputName)-4] + ".instructions.md"
		outputPath := filepath.Join(tmpProject, ".github/instructions", outputName)

		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			t.Errorf("Package file not installed: %s", outputName)
		}
	}
}

// TestExclusionPatterns tests that exclusion patterns work
func TestExclusionPatterns(t *testing.T) {
	tmpShared := t.TempDir()
	tmpProject := t.TempDir()

	// Create package with files to exclude
	pkgDir := filepath.Join(tmpShared, "testpkg")
	os.MkdirAll(pkgDir, 0755)

	files := map[string]string{
		"include.mdc": `---
description: "Include"
---
Content`,
		"exclude.draft.mdc": `---
description: "Exclude"
---
Content`,
	}

	for name, content := range files {
		path := filepath.Join(pkgDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", name, err)
		}
	}

	os.Setenv("CURSOR_RULES_DIR", tmpShared)
	defer os.Unsetenv("CURSOR_RULES_DIR")

	ctx := cli.NewAppContext(nil, nil)
	ctx.Viper.Set("workdir", tmpProject)

	// Install with exclusion
	cmd := commands.NewInstallCmd(ctx)
	cmd.SetArgs([]string{"testpkg", "--target", "copilot-instr", "--exclude", "*.draft.mdc"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Install with exclusion failed: %v", err)
	}

	// Check included file exists
	includePath := filepath.Join(tmpProject, ".github/instructions/include.instructions.md")
	if _, err := os.Stat(includePath); os.IsNotExist(err) {
		t.Error("Included file not installed")
	}

	// Check excluded file doesn't exist
	excludePath := filepath.Join(tmpProject, ".github/instructions/exclude.draft.instructions.md")
	if _, err := os.Stat(excludePath); !os.IsNotExist(err) {
		t.Error("Excluded file was installed")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
