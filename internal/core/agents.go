package core

import (
	"path/filepath"
	"strings"

	"github.com/ZanzyTHEbar/cursor-rules/internal/security"
)

const defaultAgentsSubdir = "agents"

// AgentsSubdir returns the subdir name for agents under the package dir (default "agents").
func AgentsSubdir(configured string) string {
	if s := strings.TrimSpace(configured); s != "" {
		return s
	}
	return defaultAgentsSubdir
}

// ListAgentFiles lists agent names (base names without .md) under packageDir/agents.
// agentsSubdir can be empty to use "agents".
func ListAgentFiles(packageDir, agentsSubdir string) ([]string, error) {
	subdir := AgentsSubdir(agentsSubdir)
	agentsRoot, err := security.SafeJoin(packageDir, subdir)
	if err != nil {
		return nil, err
	}
	return ListAgentFilesFrom(agentsRoot)
}

// ListAgentFilesFrom lists agent names (base without .md) in the given agents directory.
func ListAgentFilesFrom(agentsDir string) ([]string, error) {
	return listNamedFileResources(agentsDir, ".md")
}

// InstallAgentToProject installs a single agent .md file from packageDir into projectRoot/.cursor/agents/.
func InstallAgentToProject(projectRoot, packageDir, agentName, agentsSubdir string) (InstallStrategy, error) {
	subdir := AgentsSubdir(agentsSubdir)
	agentsRoot, err := security.SafeJoin(packageDir, subdir)
	if err != nil {
		return StrategyUnknown, err
	}
	return InstallAgentToDir(filepath.Join(projectRoot, ".cursor", "agents"), agentsRoot, agentName, ".md")
}

// InstallAgentToDir installs an agent file into the given agents directory.
func InstallAgentToDir(agentsDir, agentsRoot, agentName, ext string) (InstallStrategy, error) {
	return installNamedFileResourceTo(agentsDir, agentsRoot, agentName, ext)
}
