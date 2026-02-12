package transform

import (
	"github.com/ZanzyTHEbar/cursor-rules/internal/errors"
	"gopkg.in/yaml.v3"
)

// CopilotPromptsTransformer transforms Cursor rules to Copilot prompts format.
type CopilotPromptsTransformer struct {
	*CopilotInstructionsTransformer
	DefaultMode  string
	DefaultTools []string
}

// NewCopilotPromptsTransformer creates a new transformer with default settings.
func NewCopilotPromptsTransformer() *CopilotPromptsTransformer {
	return &CopilotPromptsTransformer{
		CopilotInstructionsTransformer: NewCopilotInstructionsTransformer(),
		DefaultMode:                    "chat",
		DefaultTools:                   []string{},
	}
}

// Transform converts Cursor frontmatter to Copilot prompts format.
func (t *CopilotPromptsTransformer) Transform(node *yaml.Node, body string) (*yaml.Node, string, error) {
	// Start with instructions transform
	transformed, body, err := t.CopilotInstructionsTransformer.Transform(node, body)
	if err != nil {
		return nil, "", err
	}

	var fm map[string]interface{}
	if err := transformed.Decode(&fm); err != nil {
		return nil, "", err
	}

	// Add prompt-specific fields
	if _, ok := fm["mode"]; !ok {
		fm["mode"] = t.DefaultMode
	}

	if _, ok := fm["tools"]; !ok && len(t.DefaultTools) > 0 {
		fm["tools"] = t.DefaultTools
	}

	// Remove applyTo (not used in prompts)
	delete(fm, "applyTo")

	// Encode back to YAML node
	out := &yaml.Node{}
	if err := out.Encode(fm); err != nil {
		return nil, "", errors.Wrapf(err, errors.CodeInternal, "encode frontmatter")
	}

	return out, body, nil
}

// Validate checks that required fields are present for prompts.
func (t *CopilotPromptsTransformer) Validate(node *yaml.Node) error {
	var fm map[string]interface{}
	if err := node.Decode(&fm); err != nil {
		return err
	}

	// Required fields for prompts
	if _, ok := fm["description"]; !ok {
		return errors.New(errors.CodeInvalidArgument, "missing required field: description")
	}
	if _, ok := fm["mode"]; !ok {
		return errors.New(errors.CodeInvalidArgument, "missing required field: mode")
	}

	// Validate mode enum
	if mode, ok := fm["mode"].(string); ok {
		validModes := map[string]bool{"agent": true, "edit": true, "chat": true}
		if !validModes[mode] {
			return errors.Newf(errors.CodeInvalidArgument, "invalid mode: %s (must be agent, edit, or chat)", mode)
		}
	}

	return nil
}

// Target returns the identifier for Copilot prompts format.
func (t *CopilotPromptsTransformer) Target() string {
	return "copilot-prompt"
}

// Extension returns the file extension for Copilot prompts.
func (t *CopilotPromptsTransformer) Extension() string {
	return ".prompt.md"
}

// OutputDir returns the output directory for Copilot prompts.
func (t *CopilotPromptsTransformer) OutputDir() string {
	return ".github/prompts"
}
