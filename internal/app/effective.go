package app

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ZanzyTHEbar/cursor-rules/internal/core"
)

// EffectiveRequest describes an effective rules request.
type EffectiveRequest struct {
	Target  string
	Workdir string
}

// EffectiveFile is a single effective rules file entry.
type EffectiveFile struct {
	Name    string
	Content string
}

// EffectiveResponse captures effective rules output.
type EffectiveResponse struct {
	Target        string
	SourceDir     string
	CursorContent string
	Files         []EffectiveFile
	Missing       bool
	MissingReason string
	Extension     string
}

// EffectiveRules returns effective rules for a target.
func (a *App) EffectiveRules(req EffectiveRequest) (*EffectiveResponse, error) {
	wd, err := a.ResolveWorkdir(req.Workdir, true)
	if err != nil {
		return nil, err
	}

	if req.Target == "cursor" {
		out, err := core.EffectiveRules(wd)
		if err != nil {
			return nil, err
		}
		return &EffectiveResponse{
			Target:        "cursor",
			SourceDir:     filepath.Join(wd, ".cursor", "rules"),
			CursorContent: out,
		}, nil
	}

	transformer, err := a.transformer(req.Target)
	if err != nil {
		return nil, err
	}

	rulesDir := filepath.Join(wd, transformer.OutputDir())
	resp := &EffectiveResponse{
		Target:    transformer.Target(),
		SourceDir: rulesDir,
		Extension: transformer.Extension(),
	}

	if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
		resp.Missing = true
		resp.MissingReason = fmt.Sprintf("No rules found in %s", rulesDir)
		return resp, nil
	}

	var files []string
	if err := filepath.Walk(rulesDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(path, transformer.Extension()) {
			files = append(files, path)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to walk rules directory: %w", err)
	}

	if len(files) == 0 {
		resp.Missing = true
		resp.MissingReason = fmt.Sprintf("No %s files found in %s", transformer.Extension(), rulesDir)
		return resp, nil
	}

	sort.Strings(files)
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		resp.Files = append(resp.Files, EffectiveFile{
			Name:    filepath.Base(file),
			Content: string(data),
		})
	}

	return resp, nil
}
