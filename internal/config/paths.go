package config

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	EnvPackageDir    = "CURSOR_RULES_PACKAGE_DIR"
	EnvConfigDir     = "CURSOR_RULES_CONFIG_DIR"
	EnvUserDir       = "CURSOR_USER_DIR"     // base for user/global dirs (default ~/.cursor)
	EnvUserRules     = "CURSOR_RULES_DIR"    // user rules dir (default <user-dir>/rules)
	EnvUserCommands  = "CURSOR_COMMANDS_DIR" // user commands dir (default <user-dir>/commands)
	EnvUserSkills    = "CURSOR_SKILLS_DIR"   // user skills dir (default <user-dir>/skills)
	EnvUserAgents    = "CURSOR_AGENTS_DIR"   // user agents dir (default <user-dir>/agents)
	EnvUserHooks     = "CURSOR_HOOKS_DIR"    // user hooks script dir (default <user-dir>/hooks)
	EnvUserHooksJSON = "CURSOR_HOOKS_JSON"   // user hooks.json path (default <user-dir>/hooks.json)
	EnvOpenCodeDir   = "OPENCODE_CONFIG_DIR"
)

// defaultCursorRulesBase returns ~/.cursor (or cwd/.cursor / ".cursor") for fallback.
func defaultCursorRulesBase() string {
	home, err := os.UserHomeDir()
	if err == nil && home != "" {
		return filepath.Join(home, ".cursor")
	}
	if env := os.Getenv("HOME"); env != "" {
		return filepath.Join(env, ".cursor")
	}
	if cwd, cwdErr := os.Getwd(); cwdErr == nil && cwd != "" {
		return filepath.Join(cwd, ".cursor")
	}
	return ".cursor"
}

// DefaultPackageDir returns the default package directory: CURSOR_RULES_PACKAGE_DIR or ~/.cursor/rules.
// Config file packageDir is applied in ResolvePackageDir.
func DefaultPackageDir() string {
	if v := strings.TrimSpace(os.Getenv(EnvPackageDir)); v != "" {
		return v
	}
	return filepath.Join(defaultCursorRulesBase(), "rules")
}

// DefaultConfigDir returns the config directory. Precedence: CURSOR_RULES_CONFIG_DIR >
// when CURSOR_RULES_PACKAGE_DIR is set, parent of package dir (single-root) > DefaultPackageDir().
func DefaultConfigDir() string {
	if v := strings.TrimSpace(os.Getenv(EnvConfigDir)); v != "" {
		return v
	}
	if strings.TrimSpace(os.Getenv(EnvPackageDir)) != "" {
		return filepath.Dir(DefaultPackageDir())
	}
	return DefaultPackageDir()
}

// DefaultConfigPath returns the standard config file path, honoring config dir override.
func DefaultConfigPath() string {
	return filepath.Join(DefaultConfigDir(), "config.yaml")
}

// ResolvePackageDir returns the effective package directory (env > config > default).
func ResolvePackageDir(cfg *Config) string {
	if v := strings.TrimSpace(os.Getenv(EnvPackageDir)); v != "" {
		return v
	}
	if cfg != nil {
		if v := strings.TrimSpace(cfg.PackageDir); v != "" {
			return v
		}
	}
	return DefaultPackageDir()
}

// ResolveConfigPath returns an explicit config file path if provided, otherwise default.
func ResolveConfigPath(cfgFile string) string {
	if v := strings.TrimSpace(cfgFile); v != "" {
		return v
	}
	return DefaultConfigPath()
}

// Project-side Cursor dirs (relative to project root). Use with filepath.Join(projectRoot, ...) or the helpers below.

// ProjectCursorRulesDir returns the project's .cursor/rules path.
func ProjectCursorRulesDir(projectRoot string) string {
	return filepath.Join(projectRoot, ".cursor", "rules")
}

// ProjectCursorCommandsDir returns the project's .cursor/commands path.
func ProjectCursorCommandsDir(projectRoot string) string {
	return filepath.Join(projectRoot, ".cursor", "commands")
}

// ProjectCursorSkillsDir returns the project's .cursor/skills path.
func ProjectCursorSkillsDir(projectRoot string) string {
	return filepath.Join(projectRoot, ".cursor", "skills")
}

// ProjectCursorAgentsDir returns the project's .cursor/agents path.
func ProjectCursorAgentsDir(projectRoot string) string {
	return filepath.Join(projectRoot, ".cursor", "agents")
}

// ProjectCursorHooksDir returns the project's .cursor/hooks path (script directory).
func ProjectCursorHooksDir(projectRoot string) string {
	return filepath.Join(projectRoot, ".cursor", "hooks")
}

// ProjectCursorHooksJSON returns the project's .cursor/hooks.json path.
func ProjectCursorHooksJSON(projectRoot string) string {
	return filepath.Join(projectRoot, ".cursor", "hooks.json")
}

// DefaultOpenCodeConfigDir returns the global OpenCode config directory.
// Precedence: OPENCODE_CONFIG_DIR > XDG_CONFIG_HOME/opencode > ~/.config/opencode.
func DefaultOpenCodeConfigDir() string {
	if v := strings.TrimSpace(os.Getenv(EnvOpenCodeDir)); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); v != "" {
		return filepath.Join(v, "opencode")
	}
	home, err := os.UserHomeDir()
	if err == nil && home != "" {
		return filepath.Join(home, ".config", "opencode")
	}
	if env := os.Getenv("HOME"); env != "" {
		return filepath.Join(env, ".config", "opencode")
	}
	if cwd, cwdErr := os.Getwd(); cwdErr == nil && cwd != "" {
		return filepath.Join(cwd, ".config", "opencode")
	}
	return filepath.Join(".config", "opencode")
}

// ProjectOpenCodeRulesDir returns the project's .opencode/rules path.
func ProjectOpenCodeRulesDir(projectRoot string) string {
	return filepath.Join(projectRoot, ".opencode", "rules")
}

// ProjectOpenCodeCommandsDir returns the project's .opencode/commands path.
func ProjectOpenCodeCommandsDir(projectRoot string) string {
	return filepath.Join(projectRoot, ".opencode", "commands")
}

// ProjectOpenCodeSkillsDir returns the project's .opencode/skills path.
func ProjectOpenCodeSkillsDir(projectRoot string) string {
	return filepath.Join(projectRoot, ".opencode", "skills")
}

// ProjectOpenCodeAgentsDir returns the project's .opencode/agents path.
func ProjectOpenCodeAgentsDir(projectRoot string) string {
	return filepath.Join(projectRoot, ".opencode", "agents")
}

// UserOpenCodeRulesDir returns the global OpenCode rules dir.
func UserOpenCodeRulesDir() string {
	return filepath.Join(DefaultOpenCodeConfigDir(), "rules")
}

// UserOpenCodeCommandsDir returns the global OpenCode commands dir.
func UserOpenCodeCommandsDir() string {
	return filepath.Join(DefaultOpenCodeConfigDir(), "commands")
}

// UserOpenCodeSkillsDir returns the global OpenCode skills dir.
func UserOpenCodeSkillsDir() string {
	return filepath.Join(DefaultOpenCodeConfigDir(), "skills")
}

// UserOpenCodeAgentsDir returns the global OpenCode agents dir.
func UserOpenCodeAgentsDir() string {
	return filepath.Join(DefaultOpenCodeConfigDir(), "agents")
}

// EffectiveOpenCodeRulesDir returns the project or global OpenCode rules dir.
func EffectiveOpenCodeRulesDir(projectRoot string, isUser bool) string {
	if isUser {
		return UserOpenCodeRulesDir()
	}
	return ProjectOpenCodeRulesDir(projectRoot)
}

// EffectiveOpenCodeCommandsDir returns the project or global OpenCode commands dir.
func EffectiveOpenCodeCommandsDir(projectRoot string, isUser bool) string {
	if isUser {
		return UserOpenCodeCommandsDir()
	}
	return ProjectOpenCodeCommandsDir(projectRoot)
}

// EffectiveOpenCodeSkillsDir returns the project or global OpenCode skills dir.
func EffectiveOpenCodeSkillsDir(projectRoot string, isUser bool) string {
	if isUser {
		return UserOpenCodeSkillsDir()
	}
	return ProjectOpenCodeSkillsDir(projectRoot)
}

// EffectiveOpenCodeAgentsDir returns the project or global OpenCode agents dir.
func EffectiveOpenCodeAgentsDir(projectRoot string, isUser bool) string {
	if isUser {
		return UserOpenCodeAgentsDir()
	}
	return ProjectOpenCodeAgentsDir(projectRoot)
}

// DefaultUserCursorDir returns the default user/global Cursor base directory (~/.cursor).
func DefaultUserCursorDir() string {
	home, err := os.UserHomeDir()
	if err == nil && home != "" {
		return filepath.Join(home, ".cursor")
	}
	if env := os.Getenv("HOME"); env != "" {
		return filepath.Join(env, ".cursor")
	}
	return ".cursor"
}

// UserCursorDir returns the user/global Cursor base (env CURSOR_USER_DIR or ~/.cursor).
func UserCursorDir() string {
	if v := strings.TrimSpace(os.Getenv(EnvUserDir)); v != "" {
		return v
	}
	return DefaultUserCursorDir()
}

// UserCursorRulesDir returns the user rules dir (env CURSOR_RULES_DIR or <user-dir>/rules).
func UserCursorRulesDir() string {
	if v := strings.TrimSpace(os.Getenv(EnvUserRules)); v != "" {
		return v
	}
	return filepath.Join(UserCursorDir(), "rules")
}

// UserCursorCommandsDir returns the user commands dir (env CURSOR_COMMANDS_DIR or <user-dir>/commands).
func UserCursorCommandsDir() string {
	if v := strings.TrimSpace(os.Getenv(EnvUserCommands)); v != "" {
		return v
	}
	return filepath.Join(UserCursorDir(), "commands")
}

// UserCursorSkillsDir returns the user skills dir (env CURSOR_SKILLS_DIR or <user-dir>/skills).
func UserCursorSkillsDir() string {
	if v := strings.TrimSpace(os.Getenv(EnvUserSkills)); v != "" {
		return v
	}
	return filepath.Join(UserCursorDir(), "skills")
}

// UserCursorAgentsDir returns the user agents dir (env CURSOR_AGENTS_DIR or <user-dir>/agents).
func UserCursorAgentsDir() string {
	if v := strings.TrimSpace(os.Getenv(EnvUserAgents)); v != "" {
		return v
	}
	return filepath.Join(UserCursorDir(), "agents")
}

// UserCursorHooksDir returns the user hooks script dir (env CURSOR_HOOKS_DIR or <user-dir>/hooks).
func UserCursorHooksDir() string {
	if v := strings.TrimSpace(os.Getenv(EnvUserHooks)); v != "" {
		return v
	}
	return filepath.Join(UserCursorDir(), "hooks")
}

// UserCursorHooksJSON returns the user hooks.json path (env CURSOR_HOOKS_JSON or <user-dir>/hooks.json).
func UserCursorHooksJSON() string {
	if v := strings.TrimSpace(os.Getenv(EnvUserHooksJSON)); v != "" {
		return v
	}
	return filepath.Join(UserCursorDir(), "hooks.json")
}

// GlobalProjectRoot returns a project root such that projectRoot/.cursor equals the user base.
// Use this as the effective workdir when operating in --global mode.
// When cfg is non-nil, user base is derived from packageDir (env CURSOR_USER_DIR overrides).
func GlobalProjectRoot(cfg *Config) string {
	return filepath.Dir(EffectiveUserBase(cfg))
}

// EffectiveUserBase returns the user/global Cursor base. Precedence: CURSOR_USER_DIR env >
// when cfg != nil, dir(ResolvePackageDir(cfg)) (so packageDir is source of truth) > default ~/.cursor.
func EffectiveUserBase(cfg *Config) string {
	if v := strings.TrimSpace(os.Getenv(EnvUserDir)); v != "" {
		return v
	}
	if cfg != nil {
		return filepath.Dir(ResolvePackageDir(cfg))
	}
	return DefaultUserCursorDir()
}

// EffectiveUserRulesDir returns user rules dir. Precedence: CURSOR_RULES_DIR env >
// when cfg != nil, ResolvePackageDir(cfg) (same path as package dir) > UserCursorRulesDir().
func EffectiveUserRulesDir(cfg *Config) string {
	if v := strings.TrimSpace(os.Getenv(EnvUserRules)); v != "" {
		return v
	}
	if cfg != nil {
		return ResolvePackageDir(cfg)
	}
	return UserCursorRulesDir()
}

// EffectiveUserCommandsDir returns user commands dir (env override or derived from EffectiveUserBase(cfg)).
func EffectiveUserCommandsDir(cfg *Config) string {
	if v := strings.TrimSpace(os.Getenv(EnvUserCommands)); v != "" {
		return v
	}
	return filepath.Join(EffectiveUserBase(cfg), "commands")
}

// EffectiveUserSkillsDir returns user skills dir (env override or derived from EffectiveUserBase(cfg)).
func EffectiveUserSkillsDir(cfg *Config) string {
	if v := strings.TrimSpace(os.Getenv(EnvUserSkills)); v != "" {
		return v
	}
	return filepath.Join(EffectiveUserBase(cfg), "skills")
}

// EffectiveUserAgentsDir returns user agents dir (env override or derived from EffectiveUserBase(cfg)).
func EffectiveUserAgentsDir(cfg *Config) string {
	if v := strings.TrimSpace(os.Getenv(EnvUserAgents)); v != "" {
		return v
	}
	return filepath.Join(EffectiveUserBase(cfg), "agents")
}

// EffectiveUserHooksDir returns user hooks dir (env override or derived from EffectiveUserBase(cfg)).
func EffectiveUserHooksDir(cfg *Config) string {
	if v := strings.TrimSpace(os.Getenv(EnvUserHooks)); v != "" {
		return v
	}
	return filepath.Join(EffectiveUserBase(cfg), "hooks")
}

// EffectiveUserHooksJSON returns user hooks.json path (env override or derived from EffectiveUserBase(cfg)).
func EffectiveUserHooksJSON(cfg *Config) string {
	if v := strings.TrimSpace(os.Getenv(EnvUserHooksJSON)); v != "" {
		return v
	}
	return filepath.Join(EffectiveUserBase(cfg), "hooks.json")
}

// EffectiveRulesDir returns rules dir: EffectiveUserRulesDir(cfg) when isUser, else projectRoot/.cursor/rules.
func EffectiveRulesDir(projectRoot string, isUser bool, cfg *Config) string {
	if isUser {
		return EffectiveUserRulesDir(cfg)
	}
	return ProjectCursorRulesDir(projectRoot)
}

// EffectiveCommandsDir returns commands dir: EffectiveUserCommandsDir(cfg) when isUser, else projectRoot/.cursor/commands.
func EffectiveCommandsDir(projectRoot string, isUser bool, cfg *Config) string {
	if isUser {
		return EffectiveUserCommandsDir(cfg)
	}
	return ProjectCursorCommandsDir(projectRoot)
}

// EffectiveSkillsDir returns skills dir: EffectiveUserSkillsDir(cfg) when isUser, else projectRoot/.cursor/skills.
func EffectiveSkillsDir(projectRoot string, isUser bool, cfg *Config) string {
	if isUser {
		return EffectiveUserSkillsDir(cfg)
	}
	return ProjectCursorSkillsDir(projectRoot)
}

// EffectiveAgentsDir returns agents dir: EffectiveUserAgentsDir(cfg) when isUser, else projectRoot/.cursor/agents.
func EffectiveAgentsDir(projectRoot string, isUser bool, cfg *Config) string {
	if isUser {
		return EffectiveUserAgentsDir(cfg)
	}
	return ProjectCursorAgentsDir(projectRoot)
}

// EffectiveHooksDir returns hooks dir: EffectiveUserHooksDir(cfg) when isUser, else projectRoot/.cursor/hooks.
func EffectiveHooksDir(projectRoot string, isUser bool, cfg *Config) string {
	if isUser {
		return EffectiveUserHooksDir(cfg)
	}
	return ProjectCursorHooksDir(projectRoot)
}

// EffectiveHooksJSON returns hooks.json path: EffectiveUserHooksJSON(cfg) when isUser, else projectRoot/.cursor/hooks.json.
func EffectiveHooksJSON(projectRoot string, isUser bool, cfg *Config) string {
	if isUser {
		return EffectiveUserHooksJSON(cfg)
	}
	return ProjectCursorHooksJSON(projectRoot)
}

// EffectiveCursorDirs returns the correct paths for rules, commands, skills, agents, and hooks.
// When isUser is true (--global), paths are derived from packageDir (config) with env overrides.
// When false, uses projectRoot/.cursor/<subdir>.
func EffectiveCursorDirs(projectRoot string, isUser bool, cfg *Config) struct {
	Rules     string
	Commands  string
	Skills    string
	Agents    string
	Hooks     string
	HooksJSON string
} {
	return struct {
		Rules     string
		Commands  string
		Skills    string
		Agents    string
		Hooks     string
		HooksJSON string
	}{
		Rules:     EffectiveRulesDir(projectRoot, isUser, cfg),
		Commands:  EffectiveCommandsDir(projectRoot, isUser, cfg),
		Skills:    EffectiveSkillsDir(projectRoot, isUser, cfg),
		Agents:    EffectiveAgentsDir(projectRoot, isUser, cfg),
		Hooks:     EffectiveHooksDir(projectRoot, isUser, cfg),
		HooksJSON: EffectiveHooksJSON(projectRoot, isUser, cfg),
	}
}
