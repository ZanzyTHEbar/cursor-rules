//go:build e2e
// +build e2e

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// E2E tests require the binary to be built first:
// go build -o bin/cursor-rules ./cmd/cursor-rules
// go test -tags=e2e ./...

func TestE2ECompleteWorkflow(t *testing.T) {
	// Setup
	tmpShared := t.TempDir()
	tmpProject := t.TempDir()

	// Create test rules
	createTestRules(t, tmpShared)

	// Set environment
	os.Setenv("CURSOR_RULES_DIR", tmpShared)
	defer os.Unsetenv("CURSOR_RULES_DIR")

	binary := "./bin/cursor-rules"
	if _, err := os.Stat(binary); os.IsNotExist(err) {
		t.Skip("Binary not found. Run: make build")
	}

	t.Run("install to cursor", func(t *testing.T) {
		cmd := exec.Command(binary, "install", "frontend", "--workdir", tmpProject)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Install failed: %v\nOutput: %s", err, output)
		}

		// Verify file created
		outputFile := filepath.Join(tmpProject, ".cursor/rules/frontend.mdc")
		if _, err := os.Stat(outputFile); os.IsNotExist(err) {
			t.Error("Output file not created")
		}
	})

	t.Run("install to copilot-instr", func(t *testing.T) {
		cmd := exec.Command(binary, "install", "frontend", "--target", "copilot-instr", "--workdir", tmpProject)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Install failed: %v\nOutput: %s", err, output)
		}

		// Verify file created and transformed
		outputFile := filepath.Join(tmpProject, ".github/instructions/frontend.instructions.md")
		if _, err := os.Stat(outputFile); os.IsNotExist(err) {
			t.Error("Output file not created")
		}

		content, _ := os.ReadFile(outputFile)
		contentStr := string(content)

		if !strings.Contains(contentStr, "applyTo:") {
			t.Error("applyTo field not found")
		}
		if strings.Contains(contentStr, "priority:") {
			t.Error("priority field should be removed")
		}
	})

	t.Run("transform preview", func(t *testing.T) {
		cmd := exec.Command(binary, "transform", "frontend", "--target", "copilot-instr")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Transform failed: %v\nOutput: %s", err, output)
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "applyTo:") {
			t.Error("Transform output missing applyTo field")
		}
	})

	t.Run("effective rules", func(t *testing.T) {
		cmd := exec.Command(binary, "effective", "--target", "copilot-instr", "--workdir", tmpProject)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Effective failed: %v\nOutput: %s", err, output)
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "frontend.instructions.md") {
			t.Error("Effective output missing installed rule")
		}
	})

	t.Run("idempotency check", func(t *testing.T) {
		outputFile := filepath.Join(tmpProject, ".github/instructions/frontend.instructions.md")

		// Get initial content
		content1, _ := os.ReadFile(outputFile)

		// Install again
		cmd := exec.Command(binary, "install", "frontend", "--target", "copilot-instr", "--workdir", tmpProject)
		if _, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("Second install failed: %v", err)
		}

		// Get content after second install
		content2, _ := os.ReadFile(outputFile)

		// Should be identical
		if string(content1) != string(content2) {
			t.Error("Content changed after second install (not idempotent)")
		}
	})
}

func TestE2EMultiTargetWorkflow(t *testing.T) {
	tmpShared := t.TempDir()
	tmpProject := t.TempDir()

	// Create package with manifest
	createPackageWithManifest(t, tmpShared)

	os.Setenv("CURSOR_RULES_DIR", tmpShared)
	defer os.Unsetenv("CURSOR_RULES_DIR")

	binary := "./bin/cursor-rules"
	if _, err := os.Stat(binary); os.IsNotExist(err) {
		t.Skip("Binary not found. Run: make build")
	}

	t.Run("install to all targets", func(t *testing.T) {
		cmd := exec.Command(binary, "install", "fullstack", "--all-targets", "--workdir", tmpProject)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Multi-target install failed: %v\nOutput: %s", err, output)
		}

		// Verify all targets
		targets := map[string]string{
			".cursor/rules/api.mdc":                    "cursor",
			".github/instructions/api.instructions.md": "copilot-instr",
			".github/prompts/api.prompt.md":            "copilot-prompt",
		}

		for path, target := range targets {
			fullPath := filepath.Join(tmpProject, path)
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				t.Errorf("Target %s not installed: %s", target, path)
			}
		}
	})
}

func TestE2EErrorHandling(t *testing.T) {
	binary := "./bin/cursor-rules"
	if _, err := os.Stat(binary); os.IsNotExist(err) {
		t.Skip("Binary not found. Run: make build")
	}

	t.Run("invalid target", func(t *testing.T) {
		cmd := exec.Command(binary, "install", "frontend", "--target", "invalid")
		output, err := cmd.CombinedOutput()
		if err == nil {
			t.Error("Expected error for invalid target")
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "unknown target") {
			t.Error("Error message should mention unknown target")
		}
	})

	t.Run("missing preset", func(t *testing.T) {
		tmpProject := t.TempDir()
		cmd := exec.Command(binary, "install", "nonexistent", "--workdir", tmpProject)
		output, err := cmd.CombinedOutput()
		if err == nil {
			t.Error("Expected error for missing preset")
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "not found") {
			t.Error("Error message should mention not found")
		}
	})
}

// Helper functions

func createTestRules(t *testing.T, sharedDir string) {
	rules := map[string]string{
		"frontend.mdc": `---
description: "Frontend rules"
apply_to:
  - "**/*.tsx"
  - "**/*.jsx"
priority: 1
---
Use functional components.`,
		"backend.mdc": `---
description: "Backend rules"
apply_to: "**/*.ts"
priority: 2
---
Use async/await.`,
	}

	for name, content := range rules {
		path := filepath.Join(sharedDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create rule %s: %v", name, err)
		}
	}
}

func createPackageWithManifest(t *testing.T, sharedDir string) {
	pkgDir := filepath.Join(sharedDir, "fullstack")
	os.MkdirAll(pkgDir, 0755)

	manifest := `version: "1.0"
targets:
  - cursor
  - copilot-instr
  - copilot-prompt`

	manifestPath := filepath.Join(pkgDir, "cursor-rules-manifest.yaml")
	if err := os.WriteFile(manifestPath, []byte(manifest), 0644); err != nil {
		t.Fatalf("Failed to create manifest: %v", err)
	}

	rule := `---
description: "API rules"
apply_to: "**/*.ts"
---
API content`

	rulePath := filepath.Join(pkgDir, "api.mdc")
	if err := os.WriteFile(rulePath, []byte(rule), 0644); err != nil {
		t.Fatalf("Failed to create rule: %v", err)
	}
}
