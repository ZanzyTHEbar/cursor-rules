package transform

import (
	"gopkg.in/yaml.v3"
)

// Transformer converts Cursor frontmatter to target format.
type Transformer interface {
	// Transform mutates frontmatter node and body, returning transformed versions.
	Transform(frontmatter *yaml.Node, body string) (*yaml.Node, string, error)

	// Validate checks transformed frontmatter against schema requirements.
	Validate(frontmatter *yaml.Node) error

	// Target returns the output format identifier (e.g., "cursor", "copilot-instr").
	Target() string

	// Extension returns the file extension for output files (e.g., ".mdc", ".instructions.md").
	Extension() string

	// OutputDir returns the relative path from project root for output files.
	OutputDir() string
}

// Result holds the output of a transformation operation.
type Result struct {
	Frontmatter map[string]interface{}
	Body        string
	Warnings    []string
}
