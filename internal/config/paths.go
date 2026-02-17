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

// GlobalProjectRoot returns a project root such that projectRoot/.cursor equals UserCursorDir().
// Use this as the effective workdir when operating in --global mode.
func GlobalProjectRoot() string {
	return filepath.Dir(UserCursorDir())
}
