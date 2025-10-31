package core

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

// StartWatcher watches sharedDir recursively for changes and optionally auto-applies presets to projects.
// It runs until ctx is canceled.
func StartWatcher(ctx context.Context, sharedDir string, autoApply bool) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	// add root and subdirectories
	if err := addRecursive(watcher, sharedDir); err != nil {
		watcher.Close()
		return err
	}

	// run watcher loop in a resilient goroutine: recover from panics and keep running
	go func() {
		defer watcher.Close()
		defer func() {
			if r := recover(); r != nil {
				slog.Warn("watcher panic recovered", "panic", r)
			}
		}()

		// simple event debounce: track last event time per preset filename
		lastEvent := make(map[string]time.Time)
		const debounce = 300 * time.Millisecond

		for {
			select {
			case <-ctx.Done():
				slog.Info("watcher context canceled; shutting down")
				return
			case ev, ok := <-watcher.Events:
				if !ok {
					slog.Warn("watcher events channel closed")
					return
				}
				// on write/create, log
				if ev.Op&fsnotify.Write == fsnotify.Write || ev.Op&fsnotify.Create == fsnotify.Create {
					slog.Info("detected change", "path", ev.Name)
					if !autoApply {
						continue
					}

					// debounce by filename: ignore rapid repeated events within debounce window
					now := time.Now()
					key := ev.Name
					if t, ok := lastEvent[key]; ok && now.Sub(t) < debounce {
						// skip noisy duplicate
						continue
					}
					lastEvent[key] = now

					// when new directories are created, add them recursively
					if info, statErr := os.Stat(ev.Name); statErr == nil && info.IsDir() && ev.Op&fsnotify.Create == fsnotify.Create {
						if err := addRecursive(watcher, ev.Name); err != nil {
							slog.Warn("failed to add directory recursively", "path", ev.Name, "error", err)
						}
						continue
					}

					// Wait for file to stabilize to avoid reading partially written files
					if stable := waitForStability(ev.Name, 3, 100*time.Millisecond); !stable {
						slog.Debug("file did not stabilize in time", "path", ev.Name)
						continue
					}

					presets, err := ListSharedPresets(sharedDir)
					if err != nil {
						slog.Warn("watcher list presets error", "error", err)
						continue
					}
					mapping, mapErr := LoadWatcherMapping(sharedDir)
					if mapErr != nil {
						slog.Warn("failed to load watcher mapping", "error", mapErr)
						mapping = nil
					}
					for _, p := range presets {
						name := p[:len(p)-len(filepath.Ext(p))]
						slog.Debug("watcher sees preset", "preset", name, "autoApply", autoApply)
						if mapping != nil {
							if projects, ok := mapping[name]; ok {
								for _, proj := range projects {
									if err := ApplyPresetToProject(proj, name, sharedDir); err != nil {
										slog.Warn("watcher failed to apply preset", "preset", name, "project", proj, "error", err)
									} else {
										slog.Info("watcher applied preset", "preset", name, "project", proj)
									}
								}
							} else {
								slog.Debug("watcher skipping preset without mapping", "preset", name)
							}
						} else {
							slog.Debug("watcher mapping not provided; skipping auto-apply", "preset", name)
						}
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					slog.Warn("watcher errors channel closed")
					return
				}
				slog.Error("watcher error", "error", err)
			}
		}
	}()

	return nil
}

// addRecursive walks root and adds watches for all directories.
func addRecursive(w *fsnotify.Watcher, root string) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if err := w.Add(path); err != nil {
				return err
			}
		}
		return nil
	})
}

// waitForStability checks the file size timestamp over a handful of intervals and
// returns true if it remains unchanged across checks.
func waitForStability(path string, checks int, interval time.Duration) bool {
	var lastSize int64 = -1
	for range checks {
		fi, err := os.Stat(path)
		if err != nil {
			return false
		}
		size := fi.Size()
		if lastSize == size {
			return true
		}
		lastSize = size
		time.Sleep(interval)
	}
	return false
}
