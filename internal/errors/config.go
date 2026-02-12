package errors

// ErrLegacyConfigKey returns an error when the config file uses a renamed key
// that is no longer supported. The caller should replace the legacy key with
// the replacement and re-run.
func ErrLegacyConfigKey(legacyKey, replacement string) error {
	return Newf(CodeInvalidArgument,
		"config uses legacy key %q; replace with %q and re-run",
		legacyKey, replacement)
}
