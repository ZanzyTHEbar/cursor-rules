package app

import (
	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
)

// InitRequest describes a project init request.
type InitRequest struct {
	Workdir string
}

// InitResponse captures init results.
type InitResponse struct {
	Workdir string
}

// InitProject initializes a project workspace.
func (a *App) InitProject(req InitRequest) (*InitResponse, error) {
	wd, err := a.ResolveWorkdir(req.Workdir, true)
	if err != nil {
		return nil, err
	}
	if err := core.InitProject(wd); err != nil {
		return nil, err
	}
	return &InitResponse{Workdir: wd}, nil
}
