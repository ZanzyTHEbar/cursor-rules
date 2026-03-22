package core

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ZanzyTHEbar/cursor-rules/internal/security"
)

const defaultAgentsSubdir = "agents"
const legacyAgentsSubdir = "agent"

// AgentsSubdir returns the subdir name for agents under the package dir (default "agents").
func AgentsSubdir(configured string) string {
	if s := strings.TrimSpace(configured); s != "" {
		return s
	}
	return defaultAgentsSubdir
}

// ResolveAgentsSubdir returns the effective source subdir for agents.
// When no explicit config is provided, it prefers `agents/` but falls back to
// `agent/` for compatibility with existing content trees.
func ResolveAgentsSubdir(packageDir, configured string) string {
	if s := strings.TrimSpace(configured); s != "" && s != defaultAgentsSubdir {
		return s
	}
	if s := strings.TrimSpace(configured); s == defaultAgentsSubdir {
		root := filepath.Join(packageDir, defaultAgentsSubdir)
		if info, err := os.Stat(root); err == nil && info.IsDir() {
			return defaultAgentsSubdir
		}
	}
	for _, candidate := range []string{defaultAgentsSubdir, legacyAgentsSubdir} {
		root := filepath.Join(packageDir, candidate)
		if info, err := os.Stat(root); err == nil && info.IsDir() {
			return candidate
		}
	}
	return defaultAgentsSubdir
}

// ListAgentFiles lists agent names (base names without .md) under packageDir/agents.
// agentsSubdir can be empty to use "agents".
func ListAgentFiles(packageDir, agentsSubdir string) ([]string, error) {
	subdir := ResolveAgentsSubdir(packageDir, agentsSubdir)
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
	subdir := ResolveAgentsSubdir(packageDir, agentsSubdir)
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
