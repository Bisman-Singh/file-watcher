package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	dir := flag.String("dir", ".", "directory to watch")
	pattern := flag.String("pattern", "*", "file glob pattern to match (e.g., *.go, *.txt)")
	cmd := flag.String("cmd", "", "command to execute when files change")
	recursive := flag.Bool("recursive", true, "watch directories recursively")
	debounceMs := flag.Int("debounce", 300, "debounce interval in milliseconds")
	flag.Parse()

	if *cmd == "" {
		fmt.Fprintln(os.Stderr, "Error: -cmd flag is required")
		fmt.Fprintln(os.Stderr, "Usage: file-watcher -dir . -pattern '*.go' -cmd 'go build ./...'")
		os.Exit(1)
	}

	// Verify directory exists
	info, err := os.Stat(*dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot access directory %q: %v\n", *dir, err)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: %q is not a directory\n", *dir)
		os.Exit(1)
	}

	fmt.Printf("Watching %q for pattern %q (recursive=%v, debounce=%dms)\n", *dir, *pattern, *recursive, *debounceMs)
	fmt.Printf("Command: %s\n\n", *cmd)

	debouncer := NewDebouncer(time.Duration(*debounceMs)*time.Millisecond, func() {
		executeCommand(*cmd)
	})

	watcher := NewWatcher(*dir, *pattern, *recursive, func(changed []string) {
		fmt.Printf("[%s] Changes detected:\n", time.Now().Format("15:04:05"))
		for _, path := range changed {
			fmt.Printf("  - %s\n", path)
		}
		debouncer.Trigger()
	})

	// Handle graceful shutdown
	stop := make(chan struct{})
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down watcher...")
		close(stop)
	}()

	if err := watcher.Watch(stop); err != nil {
		fmt.Fprintf(os.Stderr, "Watcher error: %v\n", err)
		os.Exit(1)
	}
}

func executeCommand(cmdStr string) {
	fmt.Printf("[%s] Running: %s\n", time.Now().Format("15:04:05"), cmdStr)

	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Command failed: %v\n", err)
	} else {
		fmt.Printf("[%s] Command completed successfully\n", time.Now().Format("15:04:05"))
	}
	fmt.Println()
}
