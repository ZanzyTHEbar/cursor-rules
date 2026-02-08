package display

import (
	"path/filepath"

	"github.com/ZanzyTHEbar/cursor-rules/internal/app"
)

// RenderInstallResponse writes install output.
func RenderInstallResponse(p Printer, resp *app.InstallResponse) {
	if resp == nil {
		return
	}
	renderInstallResults(p, resp.Results)
}

// RenderInstallAllResponse writes install-all output.
func RenderInstallAllResponse(p Printer, resp *app.InstallAllResponse) {
	if resp == nil {
		return
	}
	if len(resp.Packages) == 0 {
		p.Info("No packages found in %s\n", resp.PackageDir)
		return
	}
	renderInstallResults(p, resp.Results)
}

func renderInstallResults(p Printer, results []app.InstallResult) {
	for _, result := range results {
		if result.ShowMethod {
			p.Info("Install method: %s\n", result.Strategy)
		}
		p.Success("✅ Installed %q to %s\n", result.Name, result.OutputDir)
	}
}

// RenderListResponse writes list output.
func RenderListResponse(p Printer, resp *app.ListResponse) {
	if resp == nil {
		return
	}
	p.Info("%s\n", FormatRulesTree(resp.Tree))
}

// RenderSyncResponse writes sync output.
func RenderSyncResponse(p Printer, resp *app.SyncResponse) {
	if resp == nil {
		return
	}
	p.Success("Package dir: %s\n", resp.PackageDir)
	for _, preset := range resp.Presets {
		p.Info("- %s\n", preset)
	}
	if len(resp.Commands) > 0 {
		p.Info("commands in package dir:\n")
		for _, cmd := range resp.Commands {
			p.Info("- %s\n", cmd)
		}
	}

	for _, applied := range resp.Applied {
		if applied.DryRun {
			p.Info("would apply %s -> %s/.cursor/rules/\n", applied.Name, applied.Workdir)
			continue
		}
		if applied.Error != "" {
			p.Error("failed to apply %s: %s\n", applied.Name, applied.Error)
			continue
		}
		p.Success("applied %s -> %s/.cursor/rules/ (method: %s)\n", applied.Name, applied.Workdir, applied.Strategy)
	}
}

// RenderEffectiveResponse writes effective output.
func RenderEffectiveResponse(p Printer, resp *app.EffectiveResponse) {
	if resp == nil {
		return
	}
	if resp.Target == "cursor" {
		p.Info("%s\n", resp.CursorContent)
		return
	}

	p.Info("# Effective Rules (%s)\n\n", resp.Target)
	p.Info("Source: %s\n\n", resp.SourceDir)
	if resp.Missing {
		p.Warn("%s\n", resp.MissingReason)
		return
	}
	for _, file := range resp.Files {
		p.Info("## %s\n\n", file.Name)
		p.Info("%s\n", file.Content)
		p.Info("\n---\n")
	}
}

// RenderTransformResponse writes transform preview output.
func RenderTransformResponse(p Printer, resp *app.TransformResponse) {
	if resp == nil {
		return
	}
	p.Info("Transforming %q to %s format:\n\n", resp.Name, resp.Target)
	for _, item := range resp.Items {
		base := filepath.Base(item.SourcePath)
		if item.Error != "" {
			p.Error("❌ %s: %s\n", base, item.Error)
			continue
		}
		if item.Warning != "" {
			p.Warn("⚠️  %s: validation warning: %s\n", base, item.Warning)
		}
		p.Info("📄 %s.mdc → %s\n", item.BaseName, item.OutputName)
		p.Info("%s\n", item.Output)
		p.Info("---\n")
	}
}

// RenderInitResponse writes init output.
func RenderInitResponse(p Printer, resp *app.InitResponse) {
	if resp == nil {
		return
	}
	p.Success("Initialized project at %s/.cursor/rules/\n", resp.Workdir)
}

// RenderRemoveResponse writes remove output.
func RenderRemoveResponse(p Printer, resp *app.RemoveResponse) {
	if resp == nil {
		return
	}
	if resp.RemovedPreset {
		p.Success("Removed preset %q from %s/.cursor/rules/\n", resp.Name, resp.Workdir)
		return
	}
	if resp.RemovedCommand {
		p.Success("Removed command %q from %s/.cursor/commands/\n", resp.Name, resp.Workdir)
	}
}

// RenderConfigInitResponse writes config init output.
func RenderConfigInitResponse(p Printer, resp *app.ConfigInitResponse) {
	if resp == nil {
		return
	}
	if resp.BackupPath != "" {
		p.Info("Existing config backed up to %s\n", resp.BackupPath)
	}
	p.Info("Config written to %s (enableStow=%t)\n", resp.ConfigPath, resp.EnableStow)
}
