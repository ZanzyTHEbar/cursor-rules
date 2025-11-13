package core

// InstallStrategy describes how a preset/package was applied to a project.
type InstallStrategy string

const (
	StrategyUnknown InstallStrategy = "unknown"
	StrategyStow    InstallStrategy = "stow"
	StrategySymlink InstallStrategy = "symlink"
	StrategyCopy    InstallStrategy = "copy"
)
