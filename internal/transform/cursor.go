package transform

import (
	"gopkg.in/yaml.v3"
)

// CursorTransformer is an identity transformer that passes through Cursor rules unchanged.
type CursorTransformer struct{}

// NewCursorTransformer creates a new CursorTransformer instance.
func NewCursorTransformer() *CursorTransformer {
	return &CursorTransformer{}
}

// Transform passes through frontmatter and body unchanged.
func (t *CursorTransformer) Transform(frontmatter *yaml.Node, body string) (*yaml.Node, string, error) {
	return frontmatter, body, nil
}

// Validate performs basic validation on Cursor frontmatter.
func (t *CursorTransformer) Validate(frontmatter *yaml.Node) error {
	var fm map[string]interface{}
	if err := frontmatter.Decode(&fm); err != nil {
		return err
	}
	// Cursor rules are flexible; no strict validation required
	return nil
}

// Target returns the identifier for Cursor format.
func (t *CursorTransformer) Target() string {
	return "cursor"
}

// Extension returns the file extension for Cursor rules.
func (t *CursorTransformer) Extension() string {
	return ".mdc"
}

// OutputDir returns the output directory for Cursor rules.
func (t *CursorTransformer) OutputDir() string {
	return ".cursor/rules"
}
