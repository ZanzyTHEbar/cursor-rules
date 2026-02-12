package app

import (
	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
	"github.com/ZanzyTHEbar/cursor-rules/internal/errors"
)

// ListRequest describes a rules listing request.
type ListRequest struct {
	ConfigPath string
}

// ListResponse contains rules tree data.
type ListResponse struct {
	PackageDir string
	Tree       *core.RulesTree
}

// ListRules returns the rules tree for the configured package directory.
func (a *App) ListRules(req ListRequest) (*ListResponse, error) {
	cfg, _, err := a.LoadConfig(req.ConfigPath)
	if err != nil {
		return nil, errors.Wrapf(err, errors.CodeInternal, "load config")
	}
	packageDir := a.ResolvePackageDir(cfg)
	tree, err := core.BuildRulesTree(packageDir)
	if err != nil {
		return nil, err
	}
	return &ListResponse{
		PackageDir: packageDir,
		Tree:       tree,
	}, nil
}
