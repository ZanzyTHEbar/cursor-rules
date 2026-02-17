package core

import (
	"os"
	"path/filepath"
)

// RemovePreset removes the stub file for a preset from the project's .cursor/rules
func RemovePreset(projectRoot, preset string) error {
	rulesDir := filepath.Join(projectRoot, ".cursor", "rules")
	target := filepath.Join(rulesDir, preset+".mdc")
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return nil // Return nil if file doesn't exist (idempotent)
	}
	return os.Remove(target)
}

// RemoveCommand removes a command file from the project's .cursor/commands
func RemoveCommand(projectRoot, command string) error {
	commandsDir := filepath.Join(projectRoot, ".cursor", "commands")
	target := filepath.Join(commandsDir, command+".md")
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return nil // Return nil if file doesn't exist (idempotent)
	}
	return os.Remove(target)
}

// RemoveSkill removes a skill directory from the project's .cursor/skills
func RemoveSkill(projectRoot, skillName string) error {
	skillDir := filepath.Join(projectRoot, ".cursor", "skills", skillName)
	if _, err := os.Stat(skillDir); os.IsNotExist(err) {
		return nil
	}
	return os.RemoveAll(skillDir)
}

// RemoveAgent removes an agent file from the project's .cursor/agents
func RemoveAgent(projectRoot, agentName string) error {
	agentPath := filepath.Join(projectRoot, ".cursor", "agents", agentName+".md")
	if _, err := os.Stat(agentPath); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(agentPath)
}
