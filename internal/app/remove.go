package app

import (
	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
)

// RemoveRequest describes a remove request.
type RemoveRequest struct {
	Name    string
	Workdir string
}

// RemoveResponse captures remove results.
type RemoveResponse struct {
	Name           string
	Workdir        string
	RemovedPreset  bool
	RemovedCommand bool
}

// Remove removes a preset or command stub.
func (a *App) Remove(req RemoveRequest) (*RemoveResponse, error) {
	wd, err := a.ResolveWorkdir(req.Workdir, true)
	if err != nil {
		return nil, err
	}

	resp := &RemoveResponse{
		Name:    req.Name,
		Workdir: wd,
	}

	if err := core.RemovePreset(wd, req.Name); err == nil {
		resp.RemovedPreset = true
		return resp, nil
	}

	if err := core.RemoveCommand(wd, req.Name); err == nil {
		resp.RemovedCommand = true
		return resp, nil
	}

	return resp, nil
}
