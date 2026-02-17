package core

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestListAgentFiles(t *testing.T) {
	packageDir, err := os.MkdirTemp("", "agents-pkg-")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(packageDir)
	agentsRoot := filepath.Join(packageDir, "agents")
	if err := os.MkdirAll(agentsRoot, 0o755); err != nil {
		t.Fatalf("create agents root: %v", err)
	}
	for _, name := range []string{"verifier.md", "debugger.md"} {
		if err := os.WriteFile(filepath.Join(agentsRoot, name), []byte("---\nname: "+name+"\n---\n"), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	// non-.md file is ignored
	if err := os.WriteFile(filepath.Join(agentsRoot, "readme.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write readme: %v", err)
	}

	names, err := ListAgentFiles(packageDir, "")
	if err != nil {
		t.Fatalf("ListAgentFiles: %v", err)
	}
	if len(names) != 2 {
		t.Fatalf("expected 2 agents, got %d: %v", len(names), names)
	}
	if !slices.Contains(names, "verifier") || !slices.Contains(names, "debugger") {
		t.Fatalf("expected verifier and debugger, got %v", names)
	}
}

func TestInstallAgentToProject(t *testing.T) {
	packageDir, err := os.MkdirTemp("", "agents-pkg-")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(packageDir)
	agentsRoot := filepath.Join(packageDir, "agents")
	if err := os.MkdirAll(agentsRoot, 0o755); err != nil {
		t.Fatalf("create agents root: %v", err)
	}
	agentContent := "---\nname: verifier\ndescription: Validates work\n---\n# Verifier\n"
	if err := os.WriteFile(filepath.Join(agentsRoot, "verifier.md"), []byte(agentContent), 0o644); err != nil {
		t.Fatalf("write verifier.md: %v", err)
	}

	projectDir, err := os.MkdirTemp("", "project-")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	defer os.RemoveAll(projectDir)

	strategy, err := InstallAgentToProject(projectDir, packageDir, "verifier", "")
	if err != nil {
		t.Fatalf("InstallAgentToProject: %v", err)
	}
	if strategy != StrategyCopy && strategy != StrategySymlink && strategy != StrategyStow {
		t.Fatalf("unexpected strategy: %s", strategy)
	}
	dest := filepath.Join(projectDir, ".cursor", "agents", "verifier.md")
	if _, err := os.Stat(dest); err != nil {
		t.Fatalf("expected verifier.md at %s: %v", dest, err)
	}
}

func TestInstallAgentToProject_NotFound(t *testing.T) {
	packageDir, err := os.MkdirTemp("", "agents-pkg-")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(packageDir)
	if err := os.MkdirAll(filepath.Join(packageDir, "agents"), 0o755); err != nil {
		t.Fatalf("create agents: %v", err)
	}
	projectDir, err := os.MkdirTemp("", "project-")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	defer os.RemoveAll(projectDir)

	_, err = InstallAgentToProject(projectDir, packageDir, "nonexistent", "")
	if err == nil {
		t.Fatal("expected error for nonexistent agent")
	}
}
