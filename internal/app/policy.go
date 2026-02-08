package app

// PolicyResponse captures policy output.
type PolicyResponse struct {
	Message string
}

// Policy returns a placeholder response.
func (a *App) Policy() *PolicyResponse {
	return &PolicyResponse{Message: "policy command not yet implemented"}
}
