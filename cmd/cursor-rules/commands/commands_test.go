package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ZanzyTHEbar/cursor-rules/cli"
	"github.com/ZanzyTHEbar/cursor-rules/internal/testutil"
	"github.com/spf13/viper"
)

// TestInstallCommandErrors tests error handling in install command
func TestInstallCommandErrors(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) (string, string) // Returns sharedDir, projectDir
		presetName  string
		target      string
		wantErr     bool
		errContains string
	}{
		{
			name: "missing preset file",
			setup: func(t *testing.T) (string, string) {
				shared := t.TempDir()
				project := t.TempDir()
				return shared, project
			},
			presetName:  "nonexistent",
			target:      "cursor",
			wantErr:     true,
			errContains: "not found",
		},
		{
			name: "invalid target",
			setup: func(t *testing.T) (string, string) {
				shared := t.TempDir()
				project := t.TempDir()
				testutil.CreateTestPreset(t, shared, "test", testutil.ValidPresetWithFrontmatter())
				return shared, project
			},
			presetName:  "test",
			target:      "invalid-target",
			wantErr:     true,
			errContains: "unknown target",
		},
		{
			name: "invalid frontmatter",
			setup: func(t *testing.T) (string, string) {
				shared := t.TempDir()
				project := t.TempDir()
				testutil.CreateTestPreset(t, shared, "bad", testutil.PresetWithInvalidFrontmatter())
				return shared, project
			},
			presetName:  "bad",
			target:      "cursor",
			wantErr:     true,
			errContains: "parse",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shared, project := tt.setup(t)

			// Set environment
			os.Setenv("CURSOR_RULES_DIR", shared)
			defer os.Unsetenv("CURSOR_RULES_DIR")

			// Create context
			v := viper.New()
			v.Set("workdir", project)
			ctx := cli.NewAppContext(v, nil)

			// Create command
			cmd := NewInstallCmd(ctx)
			cmd.SetArgs([]string{tt.presetName, "--target", tt.target})

			// Execute
			err := cmd.Execute()

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Error %q does not contain %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestTransformCommandErrors tests error handling in transform command
func TestTransformCommandErrors(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) (string, string) // Returns sharedDir, presetName
		target      string
		wantErr     bool
		errContains string
	}{
		{
			name: "invalid target",
			setup: func(t *testing.T) (string, string) {
				shared := t.TempDir()
				testutil.CreateTestPreset(t, shared, "test", testutil.ValidPresetWithFrontmatter())
				return shared, "test"
			},
			target:      "invalid",
			wantErr:     true,
			errContains: "unknown",
		},
		{
			name: "missing preset",
			setup: func(t *testing.T) (string, string) {
				shared := t.TempDir()
				return shared, "nonexistent"
			},
			target:      "copilot-instr",
			wantErr:     true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shared, presetName := tt.setup(t)

			// Set environment
			os.Setenv("CURSOR_RULES_DIR", shared)
			defer os.Unsetenv("CURSOR_RULES_DIR")

			// Create context
			ctx := cli.NewAppContext(nil, nil)

			// Create command
			cmd := NewTransformCmd(ctx)
			cmd.SetArgs([]string{presetName, "--target", tt.target})

			// Execute
			err := cmd.Execute()

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Error %q does not contain %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestEffectiveCommandErrors tests error handling in effective command
func TestEffectiveCommandErrors(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) string
		target      string
		wantErr     bool
		errContains string
	}{
		{
			name: "invalid target",
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			target:      "invalid",
			wantErr:     true,
			errContains: "unknown target",
		},
		{
			name: "missing rules directory",
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			target:  "cursor",
			wantErr: false, // Should handle gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workdir := tt.setup(t)

			// Create context
			v := viper.New()
			v.Set("workdir", workdir)
			ctx := cli.NewAppContext(v, nil)

			// Create command
			cmd := NewEffectiveCmd(ctx)
			cmd.SetArgs([]string{"--target", tt.target})

			// Execute
			err := cmd.Execute()

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Error %q does not contain %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestRemoveCommandErrors tests error handling in remove command
func TestRemoveCommandErrors(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) string
		presetName  string
		wantErr     bool
		errContains string
	}{
		{
			name: "remove nonexistent preset",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				os.MkdirAll(filepath.Join(dir, ".cursor", "rules"), 0755)
				return dir
			},
			presetName:  "nonexistent",
			wantErr:     false, // Should handle gracefully
			errContains: "",
		},
		{
			name: "missing rules directory",
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			presetName:  "test",
			wantErr:     false, // Should handle gracefully
			errContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workdir := tt.setup(t)

			// Create context
			v := viper.New()
			v.Set("workdir", workdir)
			ctx := cli.NewAppContext(v, nil)

			// Create command
			cmd := NewRemoveCmd(ctx)
			cmd.SetArgs([]string{tt.presetName})

			// Execute
			err := cmd.Execute()

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Error %q does not contain %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestListCommandErrors tests error handling in list command
func TestListCommandErrors(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) string
		wantErr     bool
		errContains string
	}{
		{
			name: "missing shared directory",
			setup: func(t *testing.T) string {
				// Don't set CURSOR_RULES_DIR
				return t.TempDir()
			},
			wantErr: false, // Should handle gracefully
		},
		{
			name: "empty shared directory",
			setup: func(t *testing.T) string {
				shared := t.TempDir()
				os.Setenv("CURSOR_RULES_DIR", shared)
				return shared
			},
			wantErr: false, // Should handle gracefully (no presets)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workdir := tt.setup(t)
			defer os.Unsetenv("CURSOR_RULES_DIR")

			// Create context
			v := viper.New()
			v.Set("workdir", workdir)
			ctx := cli.NewAppContext(v, nil)

			// Create command
			cmd := NewListCmd(ctx)

			// Execute
			err := cmd.Execute()

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Error %q does not contain %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
