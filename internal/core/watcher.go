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

	// run watcher loop in a resilient goroutine: recover from panics and keep running
	go func() {
		defer watcher.Close()
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("watcher panic recovered: %v\n", r)
			}
		}()

		// simple event debounce: track last event time per preset filename
		lastEvent := make(map[string]time.Time)
		const debounce = 300 * time.Millisecond

		for {
			select {
			case ev, ok := <-watcher.Events:
				if !ok {
					fmt.Printf("watcher: events channel closed\n")
					return
				}
				// on write/create, log
				if ev.Op&fsnotify.Write == fsnotify.Write || ev.Op&fsnotify.Create == fsnotify.Create {
					fmt.Printf("detected change: %s\n", ev.Name)
					if !autoApply {
						continue
					}

					// debounce by filename: ignore rapid repeated events within debounce window
					now := time.Now()
					key := filepath.Base(ev.Name)
					if t, ok := lastEvent[key]; ok && now.Sub(t) < debounce {
						// skip noisy duplicate
						continue
					}
					lastEvent[key] = now

					// small delay to allow file writers to finish
					time.Sleep(200 * time.Millisecond)

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
			case err, ok := <-watcher.Errors:
				if !ok {
					fmt.Printf("watcher: errors channel closed\n")
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
