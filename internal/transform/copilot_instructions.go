package transform

import (
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// CopilotInstructionsTransformer transforms Cursor rules to Copilot instructions format.
type CopilotInstructionsTransformer struct {
	DefaultGlobs  []string
	MaxTokens     int
	ValidateGlobs bool
}

// NewCopilotInstructionsTransformer creates a new transformer with default settings.
func NewCopilotInstructionsTransformer() *CopilotInstructionsTransformer {
	return &CopilotInstructionsTransformer{
		DefaultGlobs:  []string{"**"},
		MaxTokens:     2000,
		ValidateGlobs: true,
	}
}

// Transform converts Cursor frontmatter to Copilot instructions format.
func (t *CopilotInstructionsTransformer) Transform(node *yaml.Node, body string) (*yaml.Node, string, error) {
	var fm map[string]interface{}
	if err := node.Decode(&fm); err != nil {
		return nil, "", fmt.Errorf("decode frontmatter: %w", err)
	}

	result := make(map[string]interface{})

	// 1. Map description (required)
	if desc, ok := fm["description"].(string); ok {
		result["description"] = desc
	} else {
		// Provide default if missing
		result["description"] = "Imported from Cursor rules"
	}

	// 2. Transform apply_to -> applyTo
	if applyTo := t.extractApplyTo(fm); applyTo != "" {
		result["applyTo"] = applyTo
	} else {
		result["applyTo"] = strings.Join(t.DefaultGlobs, ",")
	}

	// 3. Validate globs if enabled
	if t.ValidateGlobs {
		applyToStr, ok := result["applyTo"].(string)
		if !ok {
			return nil, "", fmt.Errorf("applyTo is not a string")
		}
		if validateErr := t.validateGlobPattern(applyToStr); validateErr != nil {
			return nil, "", fmt.Errorf("invalid glob: %w", validateErr)
		}
	}

	// 4. Truncate body if exceeds token limit
	body = t.truncateBody(body)

	// 5. Encode back to YAML node
	out := &yaml.Node{}
	if err := out.Encode(result); err != nil {
		return nil, "", fmt.Errorf("encode frontmatter: %w", err)
	}

	return out, body, nil
}

// extractApplyTo extracts and normalizes the applyTo field from Cursor frontmatter.
func (t *CopilotInstructionsTransformer) extractApplyTo(fm map[string]interface{}) string {
	// Check apply_to (Cursor format)
	if applyTo, ok := fm["apply_to"]; ok {
		switch v := applyTo.(type) {
		case string:
			return v
		case []interface{}:
			strs := make([]string, len(v))
			for i, item := range v {
				strs[i] = fmt.Sprint(item)
			}
			return strings.Join(strs, ",")
		}
	}

	// Check applyTo (already Copilot format)
	if applyTo, ok := fm["applyTo"].(string); ok {
		return applyTo
	}

	return ""
}

// validateGlobPattern validates glob patterns for Copilot compatibility.
func (t *CopilotInstructionsTransformer) validateGlobPattern(pattern string) error {
	// Split comma-separated patterns
	patterns := strings.Split(pattern, ",")
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		// Test with filepath.Match (basic validation)
		if _, err := filepath.Match(p, "test.ts"); err != nil {
			return fmt.Errorf("invalid pattern %q: %w", p, err)
		}
	}
	return nil
}

// truncateBody truncates the body if it exceeds the token limit.
func (t *CopilotInstructionsTransformer) truncateBody(body string) string {
	// Rough token estimation: ~4 chars per token
	maxChars := t.MaxTokens * 4
	if len(body) <= maxChars {
		return body
	}
	return body[:maxChars] + "\n\n[... truncated for token limit ...]"
}

// Validate checks that required fields are present in transformed frontmatter.
func (t *CopilotInstructionsTransformer) Validate(node *yaml.Node) error {
	var fm map[string]interface{}
	if err := node.Decode(&fm); err != nil {
		return err
	}

	// Required fields
	if _, ok := fm["description"]; !ok {
		return fmt.Errorf("missing required field: description")
	}
	if _, ok := fm["applyTo"]; !ok {
		return fmt.Errorf("missing required field: applyTo")
	}

	return nil
}

// Target returns the identifier for Copilot instructions format.
func (t *CopilotInstructionsTransformer) Target() string {
	return "copilot-instr"
}

// Extension returns the file extension for Copilot instructions.
func (t *CopilotInstructionsTransformer) Extension() string {
	return ".instructions.md"
}

// OutputDir returns the output directory for Copilot instructions.
func (t *CopilotInstructionsTransformer) OutputDir() string {
	return ".github/instructions"
}
