package transform

import (
	"strings"

	"github.com/ZanzyTHEbar/cursor-rules/internal/errors"
	"gopkg.in/yaml.v3"
)

// OpenCodeRulesTransformer transforms Cursor rules into markdown files suitable
// for the opencode-rules plugin (`.opencode/rules` or `~/.config/opencode/rules`).
type OpenCodeRulesTransformer struct{}

// NewOpenCodeRulesTransformer creates a transformer for OpenCode rule files.
func NewOpenCodeRulesTransformer() *OpenCodeRulesTransformer {
	return &OpenCodeRulesTransformer{}
}

// Transform keeps the markdown body and only preserves frontmatter fields that
// the opencode-rules plugin understands.
func (t *OpenCodeRulesTransformer) Transform(node *yaml.Node, body string) (*yaml.Node, string, error) {
	var fm map[string]interface{}
	if err := node.Decode(&fm); err != nil {
		return nil, "", errors.Wrapf(err, errors.CodeInternal, "decode frontmatter")
	}

	allowed := map[string]struct{}{
		"globs":    {},
		"keywords": {},
		"tools":    {},
		"model":    {},
		"agent":    {},
		"command":  {},
		"project":  {},
		"branch":   {},
		"os":       {},
		"ci":       {},
		"match":    {},
	}

	result := make(map[string]interface{})
	for key, value := range fm {
		if _, ok := allowed[key]; ok {
			result[key] = value
		}
	}

	if applyTo, ok := fm["apply_to"]; ok {
		result["globs"] = normalizeOpenCodeGlobs(applyTo)
	} else if applyTo, ok := fm["applyTo"]; ok {
		result["globs"] = normalizeOpenCodeGlobs(applyTo)
	}

	out := &yaml.Node{}
	if err := out.Encode(result); err != nil {
		return nil, "", errors.Wrapf(err, errors.CodeInternal, "encode frontmatter")
	}

	return out, body, nil
}

// Validate checks that the transformed rule contains only supported metadata.
func (t *OpenCodeRulesTransformer) Validate(node *yaml.Node) error {
	var fm map[string]interface{}
	if err := node.Decode(&fm); err != nil {
		return err
	}
	if match, ok := fm["match"].(string); ok {
		switch match {
		case "", "any", "all":
		default:
			return errors.Newf(errors.CodeInvalidArgument, "invalid match: %s (must be any or all)", match)
		}
	}
	return nil
}

// Target returns the identifier for OpenCode rules format.
func (t *OpenCodeRulesTransformer) Target() string {
	return "opencode-rules"
}

// Extension returns the file extension for OpenCode rule files.
func (t *OpenCodeRulesTransformer) Extension() string {
	return ".mdc"
}

// OutputDir returns the project-local output directory for OpenCode rules.
func (t *OpenCodeRulesTransformer) OutputDir() string {
	return ".opencode/rules"
}

func normalizeOpenCodeGlobs(value interface{}) []string {
	switch v := value.(type) {
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			return nil
		}
		return []string{trimmed}
	case []interface{}:
		out := make([]string, 0, len(v))
		for _, item := range v {
			trimmed := strings.TrimSpace(toString(item))
			if trimmed != "" {
				out = append(out, trimmed)
			}
		}
		if len(out) == 0 {
			return nil
		}
		return out
	case []string:
		out := make([]string, 0, len(v))
		for _, item := range v {
			trimmed := strings.TrimSpace(item)
			if trimmed != "" {
				out = append(out, trimmed)
			}
		}
		if len(out) == 0 {
			return nil
		}
		return out
	default:
		return nil
	}
}

func toString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	default:
		return ""
	}
}
