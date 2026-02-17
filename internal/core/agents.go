package core

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ZanzyTHEbar/cursor-rules/internal/errors"
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
		return nil, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid package dir or agents subdir")
	}
	entries, err := os.ReadDir(agentsRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if filepath.Ext(name) != ".md" {
			continue
		}
		base := strings.TrimSuffix(name, ".md")
		if err := security.ValidatePackageName(base); err != nil {
			continue
		}
		names = append(names, base)
	}
	sort.Strings(names)
	return names, nil
}

// InstallAgentToProject installs a single agent .md file from packageDir into projectRoot/.cursor/agents/.
func InstallAgentToProject(projectRoot, packageDir, agentName, agentsSubdir string) (InstallStrategy, error) {
	if err := security.ValidatePackageName(agentName); err != nil {
		return StrategyUnknown, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid agent name")
	}
	subdir := AgentsSubdir(agentsSubdir)
	agentsRoot, err := security.SafeJoin(packageDir, subdir)
	if err != nil {
		return StrategyUnknown, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid path")
	}
	src := filepath.Join(agentsRoot, agentName+".md")
	if _, err := os.Stat(src); err != nil {
		if os.IsNotExist(err) {
			return StrategyUnknown, errors.Newf(errors.CodeNotFound, "agent not found: %s", agentName)
		}
		return StrategyUnknown, err
	}
	destDir, err := security.SafeJoin(projectRoot, ".cursor", "agents")
	if err != nil {
		return StrategyUnknown, errors.Wrapf(err, errors.CodeInvalidArgument, "invalid project path")
	}
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return StrategyUnknown, err
	}
	dest := filepath.Join(destDir, agentName+".md")
	if _, err := os.Stat(dest); err == nil {
		return StrategyCopy, nil
	}
	return ApplySourceToDest(agentsRoot, src, dest, agentName)
}
