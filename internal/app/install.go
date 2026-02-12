package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
	"github.com/ZanzyTHEbar/cursor-rules/internal/errors"
	"github.com/ZanzyTHEbar/cursor-rules/internal/manifest"
	"github.com/ZanzyTHEbar/cursor-rules/internal/transform"
)

// InstallRequest describes a single preset/package install.
type InstallRequest struct {
	Name              string
	PackageDir        string
	Workdir           string
	Excludes          []string
	NoFlatten         bool
	Target            string
	AllTargets        bool
	ShowInstallMethod bool
}

// InstallAllRequest describes install-all behavior.
type InstallAllRequest struct {
	PackageDir             string
	Workdir                string
	Excludes               []string
	NoFlatten              bool
	Target                 string
	AllTargets             bool
	ShowInstallMethodFirst bool
}

// InstallResult captures an install outcome per target.
type InstallResult struct {
	Name       string
	Target     string
	OutputDir  string
	Strategy   core.InstallStrategy
	ShowMethod bool
}

// InstallResponse captures install outcomes.
type InstallResponse struct {
	Results []InstallResult
}

// InstallAllResponse captures install-all outcomes.
type InstallAllResponse struct {
	PackageDir string
	Packages   []string
	Results    []InstallResult
}

// Install installs a preset or package according to the request.
func (a *App) Install(req *InstallRequest) (*InstallResponse, error) {
	wd, err := a.ResolveWorkdir(req.Workdir, true)
	if err != nil {
		return nil, err
	}

	packageDir := strings.TrimSpace(req.PackageDir)
	if packageDir == "" {
		cfg, _, err := a.LoadConfig("")
		if err != nil {
			return nil, errors.Wrapf(err, errors.CodeInternal, "load config")
		}
		packageDir = a.ResolvePackageDir(cfg)
	}

	results, err := a.installInternal(&installInternalRequest{
		Workdir:           wd,
		PackageDir:        packageDir,
		Name:              req.Name,
		Excludes:          req.Excludes,
		NoFlatten:         req.NoFlatten,
		Target:            req.Target,
		AllTargets:        req.AllTargets,
		ShowInstallMethod: req.ShowInstallMethod,
	})
	if err != nil {
		return nil, err
	}

	return &InstallResponse{Results: results}, nil
}

// InstallAll installs all packages in the resolved package directory.
func (a *App) InstallAll(req *InstallAllRequest) (*InstallAllResponse, error) {
	wd, err := a.ResolveWorkdir(req.Workdir, true)
	if err != nil {
		return nil, err
	}

	packageDir := strings.TrimSpace(req.PackageDir)
	if packageDir == "" {
		cfg, _, err := a.LoadConfig("")
		if err != nil {
			return nil, errors.Wrapf(err, errors.CodeInternal, "load config")
		}
		packageDir = a.ResolvePackageDir(cfg)
	}

	pkgs, err := core.ListPackageDirs(packageDir)
	if err != nil {
		return nil, errors.Wrapf(err, errors.CodeInternal, "list packages")
	}

	resp := &InstallAllResponse{
		PackageDir: packageDir,
		Packages:   pkgs,
	}
	if len(pkgs) == 0 {
		return resp, nil
	}

	for idx, name := range pkgs {
		show := req.ShowInstallMethodFirst && idx == 0
		results, err := a.installInternal(&installInternalRequest{
			Workdir:           wd,
			PackageDir:        packageDir,
			Name:              name,
			Excludes:          req.Excludes,
			NoFlatten:         req.NoFlatten,
			Target:            req.Target,
			AllTargets:        req.AllTargets,
			ShowInstallMethod: show,
		})
		if err != nil {
			return nil, err
		}
		resp.Results = append(resp.Results, results...)
	}

	return resp, nil
}

type installInternalRequest struct {
	Workdir           string
	PackageDir        string
	Name              string
	Excludes          []string
	NoFlatten         bool
	Target            string
	AllTargets        bool
	ShowInstallMethod bool
}

func (a *App) installInternal(req *installInternalRequest) ([]InstallResult, error) {
	if req.Name == "" {
		return nil, errors.New(errors.CodeInvalidArgument, "missing preset or package name")
	}

	pkgPath := filepath.Join(req.PackageDir, req.Name)
	info, statErr := os.Stat(pkgPath)
	isPackage := statErr == nil && info.IsDir()

	var m *manifest.Manifest
	if isPackage {
		var err error
		m, err = manifest.Load(pkgPath)
		if err != nil {
			// Manifest load errors are non-fatal; proceed without it
			m = nil
		}
	}

	var targets []string
	if req.AllTargets && m != nil && len(m.Targets) > 0 {
		targets = m.Targets
	} else {
		targets = []string{req.Target}
	}

	effectiveExcludes := append([]string{}, req.Excludes...)
	if m != nil && len(m.Exclude) > 0 {
		effectiveExcludes = append(effectiveExcludes, m.Exclude...)
	}

	results := make([]InstallResult, 0, len(targets))
	for _, tgt := range targets {
		transformer, err := a.transformer(tgt)
		if err != nil {
			return nil, err
		}

		var strategy core.InstallStrategy
		if isPackage {
			if transformer.Target() == "cursor" && (core.UseSymlink() || core.WantGNUStow()) {
				strategy, err = core.InstallPackageFromPackageDir(req.Workdir, req.PackageDir, req.Name, effectiveExcludes, req.NoFlatten)
				if err != nil {
					return nil, errors.Wrapf(err, errors.CodeInternal, "install to %s failed", tgt)
				}
			} else {
				strategy, err = installPackageWithTransformer(req.Workdir, pkgPath, req.Name, transformer, effectiveExcludes, req.NoFlatten)
				if err != nil {
					return nil, errors.Wrapf(err, errors.CodeInternal, "install to %s failed", tgt)
				}
			}
		} else {
			strategy, err = installPresetWithTransformer(req.Workdir, pkgPath, req.Name, transformer, req.PackageDir)
			if err != nil {
				return nil, errors.Wrapf(err, errors.CodeInternal, "install to %s failed", tgt)
			}
		}

		results = append(results, InstallResult{
			Name:       req.Name,
			Target:     transformer.Target(),
			OutputDir:  transformer.OutputDir(),
			Strategy:   strategy,
			ShowMethod: req.ShowInstallMethod && transformer.Target() == "cursor",
		})
	}

	return results, nil
}

func (a *App) transformer(target string) (transform.Transformer, error) {
	if a == nil || a.Transformers == nil {
		return nil, errors.New(errors.CodeFailedPrecondition, "no transformers configured")
	}
	return a.Transformers.Transformer(target)
}

// installPackageWithTransformer installs a package directory using the specified transformer.
func installPackageWithTransformer(
	workDir, pkgPath, presetName string,
	transformer transform.Transformer,
	excludes []string,
	noFlatten bool,
) (core.InstallStrategy, error) {
	outDir := filepath.Join(workDir, transformer.OutputDir())
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return core.StrategyUnknown, errors.Wrapf(err, errors.CodeInternal, "create output dir for package %q", presetName)
	}

	if err := filepath.Walk(pkgPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrapf(err, errors.CodeInternal, "walk package %q", presetName)
		}
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".mdc") {
			return nil
		}

		relPath, err := filepath.Rel(pkgPath, path)
		if err != nil {
			return errors.Wrapf(err, errors.CodeInternal, "get relative path")
		}
		if shouldExclude(relPath, excludes) {
			return nil
		}
		if err := transformAndWriteFile(path, relPath, outDir, transformer, noFlatten); err != nil {
			return errors.Wrapf(err, errors.CodeInternal, "install file from package %q", presetName)
		}
		return nil
	}); err != nil {
		return core.StrategyUnknown, err
	}
	return core.StrategyCopy, nil
}

// installPresetWithTransformer installs a single preset file using the specified transformer.
func installPresetWithTransformer(
	workDir, presetPath, presetName string,
	transformer transform.Transformer,
	packageDir string,
) (core.InstallStrategy, error) {
	if !strings.HasSuffix(presetPath, ".mdc") {
		presetPath += ".mdc"
	}
	if _, err := os.Stat(presetPath); os.IsNotExist(err) {
		return core.StrategyUnknown, errors.Newf(errors.CodeNotFound, "preset %q not found: %s", presetName, presetPath)
	}

	if transformer.Target() == "cursor" {
		packageDirResolved := strings.TrimSpace(packageDir)
		if packageDirResolved == "" {
			packageDirResolved = core.DefaultPackageDir()
		}
		if core.UseSymlink() || core.WantGNUStow() {
			return core.ApplyPresetWithOptionalSymlink(workDir, presetName, packageDirResolved)
		}
	}

	outDir := filepath.Join(workDir, transformer.OutputDir())
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return core.StrategyUnknown, errors.Wrapf(err, errors.CodeInternal, "create output dir for preset %q", presetName)
	}

	if err := transformAndWriteFile(presetPath, filepath.Base(presetPath), outDir, transformer, false); err != nil {
		return core.StrategyUnknown, errors.Wrapf(err, errors.CodeInternal, "install preset %q", presetName)
	}
	return core.StrategyCopy, nil
}

// transformAndWriteFile reads, transforms, and writes a single file.
func transformAndWriteFile(
	srcPath, relPath, outDir string,
	transformer transform.Transformer,
	noFlatten bool,
) error {
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return errors.Wrapf(err, errors.CodeInternal, "read %s", srcPath)
	}

	frontmatter, body, err := transform.SplitFrontmatter(data)
	if err != nil {
		return errors.Wrapf(err, errors.CodeInternal, "parse %s", srcPath)
	}

	transformedFM, transformedBody, err := transformer.Transform(frontmatter, body)
	if err != nil {
		return errors.Wrapf(err, errors.CodeInternal, "transform %s", srcPath)
	}

	if validateErr := transformer.Validate(transformedFM); validateErr != nil {
		return errors.Wrapf(validateErr, errors.CodeInvalidArgument, "validate %s", srcPath)
	}

	var outPath string
	if noFlatten {
		outPath = filepath.Join(outDir, relPath)
	} else {
		outPath = filepath.Join(outDir, filepath.Base(relPath))
	}
	outPath = strings.TrimSuffix(outPath, ".mdc") + transformer.Extension()

	output, err := transform.MarshalMarkdown(transformedFM, transformedBody)
	if err != nil {
		return errors.Wrapf(err, errors.CodeInternal, "marshal %s", srcPath)
	}

	if info, statErr := os.Lstat(outPath); statErr == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			if rmErr := os.Remove(outPath); rmErr != nil {
				return errors.Wrapf(rmErr, errors.CodeInternal, "remove existing symlink %s", outPath)
			}
		}
	}

	existing, readErr := os.ReadFile(outPath)
	if readErr == nil && bytes.Equal(existing, output) {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}

	// #nosec G306 - rule files are meant to be world-readable
	return os.WriteFile(outPath, output, 0o644)
}

func shouldExclude(relPath string, patterns []string) bool {
	for _, pattern := range patterns {
		matched, err := filepath.Match(pattern, relPath)
		if err == nil && matched {
			return true
		}
		matched, err = filepath.Match(pattern, filepath.Dir(relPath))
		if err == nil && matched {
			return true
		}
	}
	return false
}
