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
						mapping, _ := LoadWatcherMapping(sharedDir)
						for _, p := range presets {
							name := p[:len(p)-len(filepath.Ext(p))]
							fmt.Printf("watcher sees preset: %s (autoApply=%v)\n", name, autoApply)
							if mapping != nil {
								if projects, ok := mapping[name]; ok {
									for _, proj := range projects {
										if err := ApplyPresetToProject(proj, name, sharedDir); err != nil {
											fmt.Printf("watcher: failed to apply %s -> %s: %v\n", name, proj, err)
										} else {
											fmt.Printf("watcher: applied %s -> %s\n", name, proj)
										}
									}
								} else {
									fmt.Printf("watcher: no mapping for preset %s, skipping\n", name)
								}
							} else {
								fmt.Printf("watcher: mapping not provided, skipping auto-apply for %s\n", name)
							}
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
