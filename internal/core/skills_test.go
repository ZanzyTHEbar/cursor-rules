package core

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestListSkillDirs(t *testing.T) {
	packageDir, err := os.MkdirTemp("", "skills-pkg-")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(packageDir)
	skillsRoot := filepath.Join(packageDir, "skills")
	if err := os.MkdirAll(skillsRoot, 0o755); err != nil {
		t.Fatalf("create skills root: %v", err)
	}

	// valid skill with SKILL.md
	skillA := filepath.Join(skillsRoot, "my-skill")
	if err := os.MkdirAll(skillA, 0o755); err != nil {
		t.Fatalf("create skill dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillA, "SKILL.md"), []byte("---\nname: my-skill\ndescription: test\n---\n# My Skill\n"), 0o644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}

	// dir without SKILL.md is ignored
	if err := os.MkdirAll(filepath.Join(skillsRoot, "empty-dir"), 0o755); err != nil {
		t.Fatalf("create empty dir: %v", err)
	}

	// another valid skill
	skillB := filepath.Join(skillsRoot, "other-skill")
	if err := os.MkdirAll(skillB, 0o755); err != nil {
		t.Fatalf("create skill dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillB, "SKILL.md"), []byte("---\nname: other-skill\ndescription: other\n---\n"), 0o644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}

	names, err := ListSkillDirs(packageDir, "")
	if err != nil {
		t.Fatalf("ListSkillDirs: %v", err)
	}
	if len(names) != 2 {
		t.Fatalf("expected 2 skills, got %d: %v", len(names), names)
	}
	if !slices.Contains(names, "my-skill") || !slices.Contains(names, "other-skill") {
		t.Fatalf("expected my-skill and other-skill, got %v", names)
	}

	// empty skills subdir
	names, err = ListSkillDirs(packageDir, "nonexistent")
	if err != nil {
		t.Fatalf("ListSkillDirs nonexistent: %v", err)
	}
	if len(names) != 0 {
		t.Fatalf("expected 0 skills for nonexistent subdir, got %d", len(names))
	}
}

func TestReadSkillMeta(t *testing.T) {
	dir, err := os.MkdirTemp("", "skill-meta-")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)
	content := "---\nname: test-skill\ndescription: A test skill\n---\n# Body\n"
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}
	name, desc, err := ReadSkillMeta(dir)
	if err != nil {
		t.Fatalf("ReadSkillMeta: %v", err)
	}
	if name != "test-skill" || desc != "A test skill" {
		t.Fatalf("expected name=test-skill description=A test skill, got name=%q description=%q", name, desc)
	}
}

func TestInstallSkillToProject(t *testing.T) {
	packageDir, err := os.MkdirTemp("", "skills-pkg-")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(packageDir)
	skillDir := filepath.Join(packageDir, "skills", "deploy-app")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("create skill dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: deploy-app\ndescription: Deploy\n---\n# Deploy\n"), 0o644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(skillDir, "scripts"), 0o755); err != nil {
		t.Fatalf("create scripts: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "scripts", "deploy.sh"), []byte("#!/bin/bash\necho deploy"), 0o755); err != nil {
		t.Fatalf("write script: %v", err)
	}

	projectDir, err := os.MkdirTemp("", "project-")
	if err != nil {
		t.Fatalf("create project dir: %v", err)
	}
	defer os.RemoveAll(projectDir)

	strategy, err := InstallSkillToProject(projectDir, packageDir, "deploy-app", "")
	if err != nil {
		t.Fatalf("InstallSkillToProject: %v", err)
	}
	if strategy != StrategyCopy && strategy != StrategySymlink {
		t.Fatalf("expected copy or symlink strategy, got %s", strategy)
	}
	destSkillMD := filepath.Join(projectDir, ".cursor", "skills", "deploy-app", "SKILL.md")
	if _, err := os.Stat(destSkillMD); err != nil {
		t.Fatalf("expected SKILL.md at %s: %v", destSkillMD, err)
	}
	destScript := filepath.Join(projectDir, ".cursor", "skills", "deploy-app", "scripts", "deploy.sh")
	if _, err := os.Stat(destScript); err != nil {
		t.Fatalf("expected deploy.sh at %s: %v", destScript, err)
	}
}

func TestInstallSkillToProject_NotFound(t *testing.T) {
	packageDir, err := os.MkdirTemp("", "skills-pkg-")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(packageDir)
	if err := os.MkdirAll(filepath.Join(packageDir, "skills"), 0o755); err != nil {
		t.Fatalf("create skills: %v", err)
	}
	projectDir, err := os.MkdirTemp("", "project-")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	defer os.RemoveAll(projectDir)

	_, err = InstallSkillToProject(projectDir, packageDir, "nonexistent", "")
	if err == nil {
		t.Fatal("expected error for nonexistent skill")
	}
}
