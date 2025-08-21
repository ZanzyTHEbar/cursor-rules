package core

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

// StartWatcher watches sharedDir for changes and optionally auto-applies presets to projects.
func StartWatcher(sharedDir string, autoApply bool) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	go func() {
		defer watcher.Close()
		for {
			select {
			case ev, ok := <-watcher.Events:
				if !ok {
					return
				}
				// on write/create, log
				if ev.Op&fsnotify.Write == fsnotify.Write || ev.Op&fsnotify.Create == fsnotify.Create {
					fmt.Printf("detected change: %s\n", ev.Name)
					if autoApply {
						// naive debounce
						time.Sleep(500 * time.Millisecond)
						presets, err := ListSharedPresets(sharedDir)
						if err != nil {
							fmt.Printf("watcher: list presets err: %v\n", err)
							continue
						}
						for _, p := range presets {
							name := p[:len(p)-len(filepath.Ext(p))]
							// noop if ApplyPresetToProject not given project context;
							fmt.Printf("watcher sees preset: %s (autoApply=%v)\n", name, autoApply)
							// auto-apply requires a mapping from shared presets to project paths.
							// Currently we do not have project context here, so skip applying
							// to avoid accidental writes. This prevents the previous incorrect
							// invocation that passed swapped arguments to ApplyPresetToProject.
							fmt.Printf("watcher: autoApply not configured, skipping apply for preset %s\n", name)
						}
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Printf("watcher error: %v\n", err)
			}
		}
	}()

	// watch the directory
	if err := watcher.Add(sharedDir); err != nil {
		return err
	}
	return nil
}
