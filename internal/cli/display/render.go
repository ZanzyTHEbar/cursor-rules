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

func listHeading(entry app.ListTargetEntry) string {
	if entry.Kind == "hooks" {
		return entry.Target + " (hook presets)"
	}
	return entry.Target + " (" + entry.Kind + ")"
}

// RenderListResponse writes list output (rules tree plus target-scoped entries).
func RenderListResponse(p Printer, resp *app.ListResponse) {
	if resp == nil {
		return
	}
	for _, e := range resp.Errors {
		p.Warn("warning: %s\n", e)
	}
	if resp.IncludesRules() {
		p.Info("%s\n", FormatRulesTree(resp.Tree))
	}
	for _, entry := range resp.Targets {
		if len(entry.Items) == 0 {
			continue
		}
		p.Info("%s:\n", listHeading(entry))
		for _, item := range entry.Items {
			p.Info("  - %s\n", item)
		}
	}
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
	if len(resp.Skills) > 0 {
		p.Info("skills:\n")
		for _, s := range resp.Skills {
			p.Info("- %s\n", s)
		}
	}
	if len(resp.Agents) > 0 {
		p.Info("agents:\n")
		for _, a := range resp.Agents {
			p.Info("- %s\n", a)
		}
	}
	if len(resp.Hooks) > 0 {
		p.Info("hook presets:\n")
		for _, h := range resp.Hooks {
			p.Info("- %s\n", h)
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
	for _, match := range resp.Matches {
		if !match.Removed {
			continue
		}
		if match.Kind == "hooks" {
			p.Success("Removed hooks from target %s (%s)\n", match.Target, match.Path)
			return
		}
		p.Success("Removed %s %q from target %s (%s)\n", match.Kind, match.Name, match.Target, match.Path)
		return
	}
	for _, match := range resp.Matches {
		if match.Kind == "hooks" {
			p.Info("Nothing removed from target %s (%s).\n", match.Target, match.Path)
			return
		}
		p.Info("Nothing removed for %s %q from target %s (%s).\n", match.Kind, match.Name, match.Target, match.Path)
		return
	}
	p.Info("Nothing removed.\n")
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

// RenderLinkGlobalResponse writes link-global output.
func RenderLinkGlobalResponse(p Printer, resp *app.LinkGlobalResponse) {
	if resp == nil {
		return
	}
	p.Info("Base dir: %s\n", resp.BaseDir)
	for _, r := range resp.Results {
		if r.Error != "" {
			p.Error("  %s -> %s: %s\n", r.Link, r.Target, r.Error)
			continue
		}
		p.Success("  %s -> %s\n", r.Link, r.Target)
	}
}
