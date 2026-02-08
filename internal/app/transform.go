package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ZanzyTHEbar/cursor-rules/internal/transform"
)

// TransformRequest describes a transform preview request.
type TransformRequest struct {
	Name       string
	Target     string
	PackageDir string
}

// TransformItem is a single previewed transformation.
type TransformItem struct {
	SourcePath string
	BaseName   string
	OutputName string
	Output     string
	Warning    string
	Error      string
}

// TransformResponse contains preview results.
type TransformResponse struct {
	Name   string
	Target string
	Items  []TransformItem
}

// TransformPreview performs a dry-run transform for a preset or package.
func (a *App) TransformPreview(req TransformRequest) (*TransformResponse, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("missing preset name")
	}

	transformer, err := a.transformer(req.Target)
	if err != nil {
		return nil, err
	}

	packageDir := strings.TrimSpace(req.PackageDir)
	if packageDir == "" {
		cfg, _, err := a.LoadConfig("")
		if err != nil {
			return nil, fmt.Errorf("load config: %w", err)
		}
		packageDir = a.ResolvePackageDir(cfg)
	}
	pkgPath := filepath.Join(packageDir, req.Name)

	info, err := os.Stat(pkgPath)
	if err != nil {
		pkgPath += ".mdc"
		info, err = os.Stat(pkgPath)
		if err != nil {
			return nil, fmt.Errorf("preset not found: %s", req.Name)
		}
	}

	resp := &TransformResponse{
		Name:   req.Name,
		Target: transformer.Target(),
	}

	if info.IsDir() {
		err = filepath.Walk(pkgPath, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), ".mdc") {
				return err
			}
			resp.Items = append(resp.Items, previewTransform(path, transformer))
			return nil
		})
		if err != nil {
			return nil, err
		}
		return resp, nil
	}

	resp.Items = append(resp.Items, previewTransform(pkgPath, transformer))
	return resp, nil
}

func previewTransform(path string, transformer transform.Transformer) TransformItem {
	item := TransformItem{
		SourcePath: path,
		BaseName:   strings.TrimSuffix(filepath.Base(path), ".mdc"),
		OutputName: strings.TrimSuffix(filepath.Base(path), ".mdc") + transformer.Extension(),
	}

	data, err := os.ReadFile(path)
	if err != nil {
		item.Error = err.Error()
		return item
	}

	fm, body, err := transform.SplitFrontmatter(data)
	if err != nil {
		item.Error = err.Error()
		return item
	}

	transformedFM, transformedBody, err := transformer.Transform(fm, body)
	if err != nil {
		item.Error = err.Error()
		return item
	}

	if validateErr := transformer.Validate(transformedFM); validateErr != nil {
		item.Warning = validateErr.Error()
	}

	output, err := transform.MarshalMarkdown(transformedFM, transformedBody)
	if err != nil {
		item.Error = err.Error()
		return item
	}
	item.Output = string(output)
	return item
}
