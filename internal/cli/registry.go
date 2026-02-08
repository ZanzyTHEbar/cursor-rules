package cli

// Simple global palette registry for incremental migration.
var DefaultPalette Palette

// Register registers command factories into the global palette.
func Register(factories ...CommandFactory) {
	DefaultPalette.Register(factories...)
}
