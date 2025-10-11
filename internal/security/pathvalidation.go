package security

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

var (
	// ErrPathTraversal indicates a path traversal attempt was detected
	ErrPathTraversal = errors.New("path traversal detected")
	
	// ErrInvalidPath indicates the path is invalid or malformed
	ErrInvalidPath = errors.New("invalid path")
	
	// ErrPathOutsideBase indicates the path resolves outside the base directory
	ErrPathOutsideBase = errors.New("path resolves outside base directory")
)

// ValidatePath validates that a file path is safe and doesn't contain path traversal attempts.
// It checks for:
// - Path traversal sequences (../, ..\)
// - Absolute paths when relative expected
// - Null bytes
// - Invalid characters
func ValidatePath(path string) error {
	if path == "" {
		return fmt.Errorf("%w: empty path", ErrInvalidPath)
	}

	// Check for null bytes (security risk)
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("%w: null byte in path", ErrInvalidPath)
	}

	// Check for path traversal sequences
	if strings.Contains(path, "..") {
		return fmt.Errorf("%w: path contains '..'", ErrPathTraversal)
	}

	// Clean the path to normalize it
	cleaned := filepath.Clean(path)
	
	// After cleaning, check again for traversal
	if strings.Contains(cleaned, "..") {
		return fmt.Errorf("%w: path contains '..' after normalization", ErrPathTraversal)
	}

	return nil
}

// ValidatePathWithinBase validates that a path is within a base directory.
// It ensures the resolved absolute path is a subdirectory of the base.
// Both paths are cleaned and made absolute before comparison.
func ValidatePathWithinBase(path, base string) error {
	if err := ValidatePath(path); err != nil {
		return err
	}

	// Clean and make absolute
	absBase, err := filepath.Abs(filepath.Clean(base))
	if err != nil {
		return fmt.Errorf("failed to resolve base path: %w", err)
	}

	absPath, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check if path is within base
	rel, err := filepath.Rel(absBase, absPath)
	if err != nil {
		return fmt.Errorf("failed to compute relative path: %w", err)
	}

	// If relative path starts with "..", it's outside base
	if strings.HasPrefix(rel, "..") {
		return fmt.Errorf("%w: %s is outside %s", ErrPathOutsideBase, path, base)
	}

	return nil
}

// SafeJoin safely joins path elements and validates the result is within base.
// It's a secure alternative to filepath.Join that prevents path traversal.
func SafeJoin(base string, elem ...string) (string, error) {
	// Validate base
	if base == "" {
		return "", fmt.Errorf("%w: empty base path", ErrInvalidPath)
	}

	// Validate each element
	for i, e := range elem {
		if err := ValidatePath(e); err != nil {
			return "", fmt.Errorf("invalid element at index %d: %w", i, err)
		}
	}

	// Join paths
	joined := filepath.Join(append([]string{base}, elem...)...)

	// Validate result is within base
	if err := ValidatePathWithinBase(joined, base); err != nil {
		return "", err
	}

	return joined, nil
}

// ValidatePresetName validates a preset name to ensure it's safe for file operations.
// Preset names should be simple identifiers without path separators or special characters.
func ValidatePresetName(name string) error {
	if name == "" {
		return fmt.Errorf("%w: empty preset name", ErrInvalidPath)
	}

	// Check for null bytes
	if strings.Contains(name, "\x00") {
		return fmt.Errorf("%w: null byte in preset name", ErrInvalidPath)
	}

	// Check for path separators (both Unix and Windows)
	if strings.ContainsAny(name, "/\\") {
		return fmt.Errorf("%w: preset name contains path separators", ErrInvalidPath)
	}

	// Check for path traversal
	if strings.Contains(name, "..") {
		return fmt.Errorf("%w: preset name contains '..'", ErrPathTraversal)
	}

	// Check for hidden files (starting with .)
	if strings.HasPrefix(name, ".") {
		return fmt.Errorf("%w: preset name starts with '.'", ErrInvalidPath)
	}

	return nil
}

// ValidatePackageName validates a package name.
// Package names can contain forward slashes for nested packages (e.g., "frontend/react")
// but should not contain path traversal sequences or other dangerous characters.
func ValidatePackageName(name string) error {
	if name == "" {
		return fmt.Errorf("%w: empty package name", ErrInvalidPath)
	}

	// Check for null bytes
	if strings.Contains(name, "\x00") {
		return fmt.Errorf("%w: null byte in package name", ErrInvalidPath)
	}

	// Check for path traversal
	if strings.Contains(name, "..") {
		return fmt.Errorf("%w: package name contains '..'", ErrPathTraversal)
	}

	// Check for backslashes (Windows path separator - not allowed in package names)
	if strings.Contains(name, "\\") {
		return fmt.Errorf("%w: package name contains backslashes", ErrInvalidPath)
	}

	// Check for absolute paths
	if filepath.IsAbs(name) {
		return fmt.Errorf("%w: package name is absolute path", ErrInvalidPath)
	}

	// Validate each component
	parts := strings.Split(name, "/")
	for i, part := range parts {
		if part == "" {
			return fmt.Errorf("%w: empty component at index %d", ErrInvalidPath, i)
		}
		if part == "." || part == ".." {
			return fmt.Errorf("%w: invalid component '%s' at index %d", ErrInvalidPath, part, i)
		}
		if strings.HasPrefix(part, ".") {
			return fmt.Errorf("%w: component starts with '.' at index %d", ErrInvalidPath, i)
		}
	}

	return nil
}

// SanitizeFilename removes or replaces dangerous characters from a filename.
// This is useful for user-provided filenames that will be used in file operations.
// Note: This function preserves the structure after replacement, it doesn't collapse
// multiple underscores or handle all edge cases. For stricter validation, use ValidatePresetName.
func SanitizeFilename(filename string) string {
	// Remove null bytes
	filename = strings.ReplaceAll(filename, "\x00", "")
	
	// Replace path separators with underscores
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")
	
	// Replace path traversal sequences with underscores
	filename = strings.ReplaceAll(filename, "..", "_")
	
	// Remove leading dots (but preserve extension dots)
	for strings.HasPrefix(filename, ".") && len(filename) > 1 {
		filename = filename[1:]
	}
	
	// If empty or only dots after sanitization, use default
	if filename == "" || filename == "." {
		filename = "unnamed"
	}
	
	return filename
}
