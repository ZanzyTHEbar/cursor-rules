package core

import (
	"os"
	"path/filepath"

	"github.com/ZanzyTHEbar/cursor-rules/internal/errors"
)

// RemovePreset removes the stub file for a preset from the given rules directory.
func RemovePreset(rulesDir, preset string) error {
	target := filepath.Join(rulesDir, preset+".mdc")
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return nil // Return nil if file doesn't exist (idempotent)
	}
	return os.Remove(target)
}

// RemoveCommand removes a command file from the given commands directory.
func RemoveCommand(commandsDir, command string) error {
	normalized, err := normalizeCommandName(command)
	if err != nil {
		return errors.Wrapf(err, errors.CodeInvalidArgument, "invalid command name")
	}
	if removed, err := removeInstalledNamedFileResourceFrom(commandsDir, normalized, ".md"); err != nil {
		return err
	} else if removed {
		return nil
	}
	if _, err := removeInstalledNamedDirResourceFrom(commandsDir, normalized); err != nil {
		return err
	}
	return nil
}

// RemoveSkill removes a skill directory from the given skills directory.
func RemoveSkill(skillsDir, skillName string) error {
	_, err := removeInstalledNamedDirResourceFrom(skillsDir, skillName)
	return err
}

// RemoveAgent removes an agent file from the given agents directory.
func RemoveAgent(agentsDir, agentName string) error {
	_, err := removeInstalledNamedFileResourceFrom(agentsDir, agentName, ".md")
	return err
}
