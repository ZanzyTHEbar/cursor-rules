package manifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadManifest(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		content     string
		wantErr     bool
		wantVersion string
		wantTargets int
	}{
		{
			name: "valid manifest",
			content: `version: "1.0"
targets:
  - cursor
  - copilot-instr
  - copilot-prompt`,
			wantErr:     false,
			wantVersion: "1.0",
			wantTargets: 3,
		},
		{
			name: "minimal manifest",
			content: `version: "1.0"
targets:
  - cursor`,
			wantErr:     false,
			wantVersion: "1.0",
			wantTargets: 1,
		},
		{
			name:        "empty manifest",
			content:     `{}`,
			wantErr:     false,
			wantVersion: "",
			wantTargets: 0,
		},
		{
			name: "manifest with overrides",
			content: `version: "1.0"
targets:
  - copilot-prompt
overrides:
  copilot-prompt:
    defaultMode: "agent"
    defaultTools:
      - "githubRepo"`,
			wantErr:     false,
			wantVersion: "1.0",
			wantTargets: 1,
		},
		{
			name: "invalid YAML",
			content: `version: "1.0
targets: [unclosed`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test directory
			testDir := filepath.Join(tmpDir, tt.name)
			os.MkdirAll(testDir, 0755)

			// Write manifest file
			manifestPath := filepath.Join(testDir, "cursor-rules-manifest.yaml")
			if err := os.WriteFile(manifestPath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test manifest: %v", err)
			}

			// Load manifest
			m, err := Load(testDir)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if m == nil {
				t.Error("Expected manifest, got nil")
				return
			}

			if m.Version != tt.wantVersion {
				t.Errorf("Version = %q, want %q", m.Version, tt.wantVersion)
			}

			if len(m.Targets) != tt.wantTargets {
				t.Errorf("Targets count = %d, want %d", len(m.Targets), tt.wantTargets)
			}
		})
	}
}

func TestLoadManifestMissing(t *testing.T) {
	tmpDir := t.TempDir()

	// Load from directory without manifest
	m, err := Load(tmpDir)

	if err != nil {
		t.Errorf("Unexpected error for missing manifest: %v", err)
	}

	if m != nil {
		t.Error("Expected nil manifest for missing file, got non-nil")
	}
}

func TestHasTarget(t *testing.T) {
	tests := []struct {
		name    string
		targets []string
		check   string
		want    bool
	}{
		{
			name:    "target exists",
			targets: []string{"cursor", "copilot-instr"},
			check:   "cursor",
			want:    true,
		},
		{
			name:    "target doesn't exist",
			targets: []string{"cursor"},
			check:   "copilot-instr",
			want:    false,
		},
		{
			name:    "empty targets",
			targets: []string{},
			check:   "cursor",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Manifest{Targets: tt.targets}
			got := m.HasTarget(tt.check)
			if got != tt.want {
				t.Errorf("HasTarget(%q) = %v, want %v", tt.check, got, tt.want)
			}
		})
	}
}

func TestHasTargetNilSafe(t *testing.T) {
	var m *Manifest = nil
	if m.HasTarget("cursor") {
		t.Error("Nil manifest should return false for HasTarget")
	}
}

func TestGetOverride(t *testing.T) {
	m := &Manifest{
		Overrides: map[string]Override{
			"copilot-prompt": {
				DefaultMode:  "agent",
				DefaultTools: []string{"githubRepo"},
			},
		},
	}

	// Test existing override
	override := m.GetOverride("copilot-prompt")
	if override == nil {
		t.Fatal("Expected override, got nil")
	}
	if override.DefaultMode != "agent" {
		t.Errorf("DefaultMode = %q, want %q", override.DefaultMode, "agent")
	}

	// Test non-existing override
	override = m.GetOverride("cursor")
	if override != nil {
		t.Error("Expected nil for non-existing override")
	}
}

func TestGetOverrideNilSafe(t *testing.T) {
	var m *Manifest = nil
	override := m.GetOverride("cursor")
	if override != nil {
		t.Error("Nil manifest should return nil for GetOverride")
	}
}
