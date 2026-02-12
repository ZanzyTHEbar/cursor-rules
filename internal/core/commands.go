package core

import (
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/ZanzyTHEbar/cursor-rules/internal/errors"
)

var commandStubTmpl = `---
description: "Shared command: {{ .Command }}"
---
@file {{ .SourcePath }}
`

// DefaultSharedCommandsDir returns ~/.cursor-commands by default; environment overrides allowed.
func DefaultSharedCommandsDir() string {
	// Commands live under the main cursor-rules package directory. Use that by default.
	// But if CURSOR_COMMANDS_DIR is explicitly set, use that instead.
	if v := os.Getenv("CURSOR_COMMANDS_DIR"); v != "" {
		return v
	}
	return DefaultPackageDir()
}

// InstallCommand writes a small stub .md in the project's .cursor/commands/
// pointing to the command under commandsDir (default: ~/.cursor-commands).
func InstallCommand(projectRoot, command string) error {
	sharedDir := DefaultSharedCommandsDir()

	// Normalize command name: remove .md extension if present
	normalized := strings.TrimSuffix(command, ".md")
	src := filepath.Join(sharedDir, normalized+".md")

	// If source not found, allow package-style layout when stow is enabled
	if _, err := os.Stat(src); os.IsNotExist(err) {
		d := filepath.Join(sharedDir, normalized)
		if info, err := os.Stat(d); err != nil || !info.IsDir() {
			return errors.Newf(errors.CodeNotFound, "command not found: %s (expected %s)", command, src)
		}
	}

	commandsDir := filepath.Join(projectRoot, ".cursor", "commands")
	if err := os.MkdirAll(commandsDir, 0o755); err != nil {
		return err
	}

	dest := filepath.Join(commandsDir, normalized+".md")

	// If symlink/stow behavior requested, prefer that path
	if UseSymlink() || WantGNUStow() {
		return ApplyCommandWithOptionalSymlink(projectRoot, normalized, sharedDir)
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}

	t := template.Must(template.New("cmdstub").Parse(commandStubTmpl))
	data := map[string]string{
		"Command":    normalized,
		"SourcePath": src,
	}
	return AtomicWriteTemplate(filepath.Dir(dest), dest, t, data, 0o644)
}

// ApplyCommandToProject copies a command file into the project's .cursor/commands as a stub (@file).
func ApplyCommandToProject(projectRoot, command, sourceDir string) error {
	normalized := strings.TrimSuffix(command, ".md")
	src := filepath.Join(sourceDir, normalized+".md")
	if _, err := os.Stat(src); err != nil {
		return errors.Newf(errors.CodeNotFound, "shared command not found: %s", src)
	}
	commandsDir := filepath.Join(projectRoot, ".cursor", "commands")
	if err := os.MkdirAll(commandsDir, 0o755); err != nil {
		return err
	}
	dest := filepath.Join(commandsDir, normalized+".md")
	if _, err := os.Stat(dest); err == nil {
		return nil
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString("---\n@file " + src + "\n")
	if err != nil {
		return err
	}
	return nil
}

// ApplyCommandWithOptionalSymlink attempts to apply a command via stow/symlink or stub.
func ApplyCommandWithOptionalSymlink(projectRoot, command, sourceDir string) error {
	commandsDir := filepath.Join(projectRoot, ".cursor", "commands")
	if err := os.MkdirAll(commandsDir, 0o755); err != nil {
		return err
	}
	src := filepath.Join(sourceDir, command+".md")
	dest := filepath.Join(commandsDir, command+".md")
	// Delegate to shared ApplySourceToDest which handles stow -> symlink -> stub
	_, err := ApplySourceToDest(sourceDir, src, dest, command)
	return err
}

// ListSharedCommands returns list of .md files found in commandsDir.
func ListSharedCommands(commandsDir string) ([]string, error) {
	var out []string
	entries, err := os.ReadDir(commandsDir)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if filepath.Ext(e.Name()) == ".md" {
			out = append(out, e.Name())
		}
	}
	return out, nil
}

// InstallCommandPackage installs an entire package directory from sharedDir into the project's
// .cursor/commands. The package is a directory under sharedDir (e.g. "tools" or "git-helpers").
// It supports excluding specific files via the excludes slice and respects a
// .cursor-commands-ignore file placed inside the package which lists patterns to skip.
// By default, packages are flattened into .cursor/commands/. Use noFlatten=true to preserve structure.
func InstallCommandPackage(projectRoot, packageName string, excludes []string, noFlatten bool) error {
	sharedDir := DefaultSharedCommandsDir()
	return InstallPackageGeneric(projectRoot, sharedDir, packageName, "commands", []string{".md"}, ".cursor-commands-ignore", excludes, noFlatten)
}
