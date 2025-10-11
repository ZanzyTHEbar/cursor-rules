package transform

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"
)

// SplitFrontmatter separates YAML frontmatter from markdown body.
// Expects format: ---\nYAML\n---\nBody
func SplitFrontmatter(data []byte) (*yaml.Node, string, error) {
	// Split on --- delimiters
	parts := bytes.SplitN(data, []byte("---"), 3)
	if len(parts) < 3 {
		return nil, "", fmt.Errorf("invalid frontmatter format: expected --- delimiters")
	}

	// Parse YAML frontmatter
	var node yaml.Node
	if err := yaml.Unmarshal(parts[1], &node); err != nil {
		return nil, "", fmt.Errorf("parse YAML frontmatter: %w", err)
	}

	// Extract body (trim leading/trailing whitespace)
	body := string(bytes.TrimSpace(parts[2]))

	return &node, body, nil
}

// MarshalMarkdown combines YAML frontmatter and body into a markdown file.
func MarshalMarkdown(frontmatter *yaml.Node, body string) ([]byte, error) {
	// Marshal frontmatter to YAML
	fmBytes, err := yaml.Marshal(frontmatter)
	if err != nil {
		return nil, fmt.Errorf("marshal frontmatter: %w", err)
	}

	// Combine with delimiters and body
	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(fmBytes)
	buf.WriteString("---\n\n")
	buf.WriteString(body)

	return buf.Bytes(), nil
}
