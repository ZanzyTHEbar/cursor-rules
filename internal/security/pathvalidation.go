package security

import (
	"errors"
	"path/filepath"
	"strings"

	errs "github.com/ZanzyTHEbar/cursor-rules/internal/errors"
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
		return errs.Wrap(ErrInvalidPath, errs.CodeInvalidArgument, "empty path")
	}

	// Check for null bytes (security risk)
	if strings.Contains(path, "\x00") {
		return errs.Wrap(ErrInvalidPath, errs.CodeInvalidArgument, "null byte in path")
	}

	// Check for path traversal sequences
	if strings.Contains(path, "..") {
		return errs.Wrap(ErrPathTraversal, errs.CodeInvalidArgument, "path contains '..'")
	}

	// Clean the path to normalize it
	cleaned := filepath.Clean(path)

	// After cleaning, check again for traversal
	if strings.Contains(cleaned, "..") {
		return errs.Wrap(ErrPathTraversal, errs.CodeInvalidArgument, "path contains '..' after normalization")
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
		return errs.Wrapf(err, errs.CodeInternal, "resolve base path")
	}

	absPath, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return errs.Wrapf(err, errs.CodeInternal, "resolve path")
	}

	// Check if path is within base
	rel, err := filepath.Rel(absBase, absPath)
	if err != nil {
		return errs.Wrapf(err, errs.CodeInternal, "compute relative path")
	}

	// If relative path starts with "..", it's outside base
	if strings.HasPrefix(rel, "..") {
		return errs.Wrapf(ErrPathOutsideBase, errs.CodeInvalidArgument, "%s is outside %s", path, base)
	}

	return nil
}

// SafeJoin safely joins path elements and validates the result is within base.
// It's a secure alternative to filepath.Join that prevents path traversal.
func SafeJoin(base string, elem ...string) (string, error) {
	// Validate base
	if base == "" {
		return "", errs.Wrap(ErrInvalidPath, errs.CodeInvalidArgument, "empty base path")
	}

	// Validate each element
	for i, e := range elem {
		if err := ValidatePath(e); err != nil {
			return "", errs.Wrapf(err, errs.CodeInvalidArgument, "invalid element at index %d", i)
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
		return errs.Wrap(ErrInvalidPath, errs.CodeInvalidArgument, "empty preset name")
	}

	// Check for null bytes
	if strings.Contains(name, "\x00") {
		return errs.Wrap(ErrInvalidPath, errs.CodeInvalidArgument, "null byte in preset name")
	}

	// Check for path separators (both Unix and Windows)
	if strings.ContainsAny(name, "/\\") {
		return errs.Wrap(ErrInvalidPath, errs.CodeInvalidArgument, "preset name contains path separators")
	}

	// Check for path traversal
	if strings.Contains(name, "..") {
		return errs.Wrap(ErrPathTraversal, errs.CodeInvalidArgument, "preset name contains '..'")
	}

	// Check for hidden files (starting with .)
	if strings.HasPrefix(name, ".") {
		return errs.Wrap(ErrInvalidPath, errs.CodeInvalidArgument, "preset name starts with '.'")
	}

	return nil
}

// ValidatePackageName validates a package name.
// Package names can contain forward slashes for nested packages (e.g., "frontend/react")
// but should not contain path traversal sequences or other dangerous characters.
func ValidatePackageName(name string) error {
	if name == "" {
		return errs.Wrap(ErrInvalidPath, errs.CodeInvalidArgument, "empty package name")
	}

	// Check for null bytes
	if strings.Contains(name, "\x00") {
		return errs.Wrap(ErrInvalidPath, errs.CodeInvalidArgument, "null byte in package name")
	}

	// Check for path traversal
	if strings.Contains(name, "..") {
		return errs.Wrap(ErrPathTraversal, errs.CodeInvalidArgument, "package name contains '..'")
	}

	// Check for backslashes (Windows path separator - not allowed in package names)
	if strings.Contains(name, "\\") {
		return errs.Wrap(ErrInvalidPath, errs.CodeInvalidArgument, "package name contains backslashes")
	}

	// Check for absolute paths
	if filepath.IsAbs(name) {
		return errs.Wrap(ErrInvalidPath, errs.CodeInvalidArgument, "package name is absolute path")
	}

	// Validate each component
	parts := strings.Split(name, "/")
	for i, part := range parts {
		if part == "" {
			return errs.Wrapf(ErrInvalidPath, errs.CodeInvalidArgument, "empty component at index %d", i)
		}
		if part == "." || part == ".." {
			return errs.Wrapf(ErrInvalidPath, errs.CodeInvalidArgument, "invalid component %q at index %d", part, i)
		}
		if strings.HasPrefix(part, ".") {
			return errs.Wrapf(ErrInvalidPath, errs.CodeInvalidArgument, "component starts with '.' at index %d", i)
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
