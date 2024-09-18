package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// FileState stores the last known modification time and size for a file.
type FileState struct {
	ModTime time.Time
	Size    int64
}

// Watcher polls a directory for file changes and triggers a callback.
type Watcher struct {
	Dir       string
	Pattern   string
	Recursive bool
	Interval  time.Duration
	OnChange  func(changed []string)

	states map[string]FileState
}

// NewWatcher creates a new Watcher.
func NewWatcher(dir, pattern string, recursive bool, onChange func([]string)) *Watcher {
	return &Watcher{
		Dir:       dir,
		Pattern:   pattern,
		Recursive: recursive,
		Interval:  500 * time.Millisecond,
		OnChange:  onChange,
		states:    make(map[string]FileState),
	}
}

// Watch starts the polling loop. It blocks until the stop channel is closed.
func (w *Watcher) Watch(stop <-chan struct{}) error {
	// Initial scan to set baseline
	if err := w.scan(true); err != nil {
		return fmt.Errorf("initial scan failed: %w", err)
	}

	ticker := time.NewTicker(w.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return nil
		case <-ticker.C:
			if err := w.scan(false); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: scan error: %v\n", err)
			}
		}
	}
}

func (w *Watcher) scan(initial bool) error {
	currentFiles := make(map[string]FileState)
	var changed []string

	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip files we can't read
		}

		// Skip directories themselves (but walk into them)
		if info.IsDir() {
			if !w.Recursive && path != w.Dir {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file matches the pattern
		matched, err := filepath.Match(w.Pattern, filepath.Base(path))
		if err != nil {
			return nil
		}
		if !matched {
			return nil
		}

		state := FileState{
			ModTime: info.ModTime(),
			Size:    info.Size(),
		}
		currentFiles[path] = state

		if !initial {
			prev, exists := w.states[path]
			if !exists {
				// New file
				changed = append(changed, path)
			} else if prev.ModTime != state.ModTime || prev.Size != state.Size {
				// Modified file
				changed = append(changed, path)
			}
		}

		return nil
	}

	if err := filepath.Walk(w.Dir, walkFn); err != nil {
		return err
	}

	// Check for deleted files
	if !initial {
		for path := range w.states {
			if _, exists := currentFiles[path]; !exists {
				changed = append(changed, path+" (deleted)")
			}
		}
	}

	w.states = currentFiles

	if len(changed) > 0 && w.OnChange != nil {
		w.OnChange(changed)
	}

	return nil
}
