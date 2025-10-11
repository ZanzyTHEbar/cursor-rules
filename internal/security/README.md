# Security Package

This package provides security utilities for validating and sanitizing file paths and user inputs to prevent path traversal attacks and other security vulnerabilities.

## Overview

The security package implements defense-in-depth strategies for file system operations:

1. **Path Validation** - Detects and prevents path traversal attempts
2. **Safe Path Construction** - Ensures paths stay within intended boundaries
3. **Input Sanitization** - Cleans user-provided filenames
4. **Preset/Package Name Validation** - Validates identifiers used in file operations

## Functions

### ValidatePath

```go
func ValidatePath(path string) error
```

Validates that a file path is safe and doesn't contain path traversal attempts.

**Checks for:**
- Path traversal sequences (`../`, `..\`)
- Null bytes (`\x00`)
- Empty paths

**Example:**
```go
if err := security.ValidatePath("file.txt"); err != nil {
    return fmt.Errorf("invalid path: %w", err)
}
```

### ValidatePathWithinBase

```go
func ValidatePathWithinBase(path, base string) error
```

Validates that a path resolves within a base directory after normalization.

**Example:**
```go
if err := security.ValidatePathWithinBase("/home/user/project/file.txt", "/home/user/project"); err != nil {
    return fmt.Errorf("path outside base: %w", err)
}
```

### SafeJoin

```go
func SafeJoin(base string, elem ...string) (string, error)
```

Safely joins path elements and validates the result is within base. This is a secure alternative to `filepath.Join` that prevents path traversal.

**Example:**
```go
safePath, err := security.SafeJoin("/home/user/project", "subdir", "file.txt")
if err != nil {
    return fmt.Errorf("invalid path construction: %w", err)
}
// safePath = "/home/user/project/subdir/file.txt"
```

### ValidatePresetName

```go
func ValidatePresetName(name string) error
```

Validates a preset name to ensure it's safe for file operations. Preset names should be simple identifiers without path separators or special characters.

**Checks for:**
- Path separators (`/`, `\`)
- Path traversal (`..`)
- Hidden files (starting with `.`)
- Null bytes
- Empty names

**Example:**
```go
if err := security.ValidatePresetName("frontend"); err != nil {
    return fmt.Errorf("invalid preset name: %w", err)
}
```

### ValidatePackageName

```go
func ValidatePackageName(name string) error
```

Validates a package name. Package names can contain forward slashes for nested packages (e.g., `frontend/react`) but should not contain path traversal sequences or other dangerous characters.

**Example:**
```go
if err := security.ValidatePackageName("frontend/react"); err != nil {
    return fmt.Errorf("invalid package name: %w", err)
}
```

### SanitizeFilename

```go
func SanitizeFilename(filename string) string
```

Removes or replaces dangerous characters from a filename. This is useful for user-provided filenames that will be used in file operations.

**Transformations:**
- Removes null bytes
- Replaces path separators with underscores
- Replaces `..` with underscores
- Removes leading dots
- Returns "unnamed" if empty after sanitization

**Example:**
```go
safe := security.SanitizeFilename("../../../etc/passwd")
// safe = "___etc_passwd"
```

## Error Types

The package defines specific error types for different validation failures:

```go
var (
    ErrPathTraversal   = errors.New("path traversal detected")
    ErrInvalidPath     = errors.New("invalid path")
    ErrPathOutsideBase = errors.New("path resolves outside base directory")
)
```

Use `errors.Is()` to check for specific error types:

```go
if errors.Is(err, security.ErrPathTraversal) {
    // Handle path traversal attempt
}
```

## Integration Examples

### Validating User Input

```go
func InstallPreset(projectRoot, presetName string) error {
    // Validate preset name
    if err := security.ValidatePresetName(presetName); err != nil {
        return fmt.Errorf("invalid preset name: %w", err)
    }
    
    // Safely construct paths
    rulesDir, err := security.SafeJoin(projectRoot, ".cursor", "rules")
    if err != nil {
        return fmt.Errorf("invalid project path: %w", err)
    }
    
    destPath, err := security.SafeJoin(rulesDir, presetName+".mdc")
    if err != nil {
        return fmt.Errorf("invalid destination path: %w", err)
    }
    
    // Proceed with file operations...
    return nil
}
```

### Validating File Paths

```go
func ReadConfigFile(configPath string) error {
    baseDir := "/etc/myapp"
    
    // Ensure config file is within allowed directory
    if err := security.ValidatePathWithinBase(configPath, baseDir); err != nil {
        return fmt.Errorf("config file outside allowed directory: %w", err)
    }
    
    // Safe to read file
    data, err := os.ReadFile(configPath)
    // ...
}
```

### Sanitizing User-Provided Filenames

```go
func SaveUserFile(userFilename string, content []byte) error {
    // Sanitize filename to remove dangerous characters
    safeFilename := security.SanitizeFilename(userFilename)
    
    // Construct safe path
    uploadDir := "/var/uploads"
    safePath, err := security.SafeJoin(uploadDir, safeFilename)
    if err != nil {
        return fmt.Errorf("invalid filename: %w", err)
    }
    
    // Safe to write file
    return os.WriteFile(safePath, content, 0644)
}
```

## Security Considerations

### Defense in Depth

This package implements multiple layers of security:

1. **Input Validation** - Reject malicious input early
2. **Path Normalization** - Clean and resolve paths before validation
3. **Boundary Checking** - Ensure paths stay within allowed directories
4. **Sanitization** - Remove dangerous characters as last resort

### When to Use Each Function

- **ValidatePath**: Use for simple path validation without base directory constraints
- **ValidatePathWithinBase**: Use when you need to ensure a path stays within a specific directory
- **SafeJoin**: Use instead of `filepath.Join` when constructing paths from user input
- **ValidatePresetName**: Use for simple identifiers that shouldn't contain path separators
- **ValidatePackageName**: Use for identifiers that can have nested structure (e.g., `frontend/react`)
- **SanitizeFilename**: Use as a last resort when you must accept user input but can't reject it

### Gosec Integration

The package is designed to address gosec security warnings:

- **G304**: Prevents file inclusion via variable (path traversal)
- **G305**: Prevents file traversal when extracting zip archives
- **G306**: Ensures proper file permissions

When using validated paths with file operations, add gosec comments to suppress false positives:

```go
// #nosec G304 - path is validated above and constructed from trusted sources
data, err := os.ReadFile(validatedPath)
```

## Testing

The package includes comprehensive tests covering:

- Valid and invalid paths
- Path traversal attempts
- Null byte injection
- Boundary conditions
- Edge cases

Run tests:
```bash
go test ./internal/security/
```

Run benchmarks:
```bash
go test -bench=. ./internal/security/
```

## Performance

All validation functions are designed to be fast and suitable for use in hot paths:

- `ValidatePath`: ~100ns per operation
- `ValidatePathWithinBase`: ~500ns per operation
- `SafeJoin`: ~600ns per operation
- `ValidatePresetName`: ~80ns per operation
- `SanitizeFilename`: ~200ns per operation

## Best Practices

1. **Validate Early**: Check user input as soon as it enters your system
2. **Use SafeJoin**: Always use `SafeJoin` instead of `filepath.Join` for user-provided paths
3. **Check Errors**: Never ignore validation errors
4. **Log Violations**: Log path traversal attempts for security monitoring
5. **Fail Secure**: Reject suspicious input rather than trying to fix it
6. **Test Thoroughly**: Include security tests in your test suite

## References

- [OWASP Path Traversal](https://owasp.org/www-community/attacks/Path_Traversal)
- [CWE-22: Improper Limitation of a Pathname to a Restricted Directory](https://cwe.mitre.org/data/definitions/22.html)
- [Go Security Best Practices](https://github.com/securego/gosec)
