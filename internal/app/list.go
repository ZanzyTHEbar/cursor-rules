package app

import (
	"path/filepath"

	"github.com/ZanzyTHEbar/cursor-rules/internal/config"
	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
	"github.com/ZanzyTHEbar/cursor-rules/internal/errors"
)

// ListRequest describes a rules listing request.
type ListRequest struct {
	ConfigPath string
	Global     bool // if true, list from user dirs (~/.cursor/...) instead of package dir
}

// ListResponse contains rules tree data plus commands, skills, agents, and hooks from the package dir.
type ListResponse struct {
	PackageDir string
	Tree       *core.RulesTree
	Commands   []string
	Skills     []string
	Agents     []string
	Hooks      []string
}

// ListRules returns the rules tree and available commands, skills, agents, and hooks for the configured package directory or user dirs when Global is true.
func (a *App) ListRules(req ListRequest) (*ListResponse, error) {
	cfg, _, err := a.LoadConfig(req.ConfigPath)
	if err != nil {
		return nil, errors.Wrapf(err, errors.CodeInternal, "load config")
	}
	var packageDir string
	if req.Global {
		packageDir = config.UserCursorDir()
		// List from user dirs: rules tree from UserCursorRulesDir(), commands from UserCursorCommandsDir(), etc.
		tree, err := core.BuildRulesTree(config.UserCursorRulesDir())
		if err != nil {
			return nil, err
		}
		resp := &ListResponse{
			PackageDir: packageDir,
			Tree:       tree,
		}
		if cmds, err := core.ListSharedCommands(config.UserCursorCommandsDir()); err == nil {
			resp.Commands = cmds
		}
		skillsParent := filepath.Dir(config.UserCursorSkillsDir())
		skillsSubdir := filepath.Base(config.UserCursorSkillsDir())
		if skills, err := core.ListSkillDirs(skillsParent, skillsSubdir); err == nil {
			resp.Skills = skills
		}
		agentsParent := filepath.Dir(config.UserCursorAgentsDir())
		agentsSubdir := filepath.Base(config.UserCursorAgentsDir())
		if agents, err := core.ListAgentFiles(agentsParent, agentsSubdir); err == nil {
			resp.Agents = agents
		}
		hooksParent := filepath.Dir(config.UserCursorHooksDir())
		hooksSubdir := filepath.Base(config.UserCursorHooksDir())
		if hooks, err := core.ListHookPresets(hooksParent, hooksSubdir); err == nil {
			resp.Hooks = hooks
		}
		return resp, nil
	}
	packageDir = a.ResolvePackageDir(cfg)
	tree, err := core.BuildRulesTree(packageDir)
	if err != nil {
		return nil, err
	}
	resp := &ListResponse{
		PackageDir: packageDir,
		Tree:       tree,
	}
	if cmds, err := core.ListSharedCommands(packageDir); err == nil {
		resp.Commands = cmds
	}
	if skills, err := core.ListSkillDirs(packageDir, cfg.SkillsSubdir); err == nil {
		resp.Skills = skills
	}
	if agents, err := core.ListAgentFiles(packageDir, cfg.AgentsSubdir); err == nil {
		resp.Agents = agents
	}
	if hooks, err := core.ListHookPresets(packageDir, cfg.HooksSubdir); err == nil {
		resp.Hooks = hooks
	}
	return resp, nil
}
