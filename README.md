# File Watcher

A directory file watcher written in Go. Monitors a directory for file changes matching a glob pattern and executes a command when changes are detected. Includes debouncing to coalesce rapid changes.

## Features

- Watch a directory for file creation, modification, and deletion
- Filter by glob pattern (e.g., `*.go`, `*.txt`)
- Recursive or non-recursive directory watching
- Debounce rapid changes (configurable, default 300ms)
- Execute any shell command on change
- Graceful shutdown on SIGINT/SIGTERM
- Uses only the Go standard library (polling-based)

## Build

```bash
go build -o file-watcher .
```

## Usage

```bash
# Watch current directory for Go file changes and rebuild
./file-watcher -dir . -pattern '*.go' -cmd 'go build ./...'

# Watch a specific directory for any file changes
./file-watcher -dir /path/to/project -pattern '*' -cmd 'echo files changed'

# Non-recursive watch with custom debounce
./file-watcher -dir ./src -pattern '*.js' -cmd 'npm run build' -recursive=false -debounce 500

# Watch for config file changes
./file-watcher -dir /etc/myapp -pattern '*.yaml' -cmd 'systemctl reload myapp'
```

### Flags

| Flag         | Default | Description                              |
|--------------|---------|------------------------------------------|
| `-dir`       | `.`     | Directory to watch                       |
| `-pattern`   | `*`     | File glob pattern to match               |
| `-cmd`       | (none)  | Command to execute on changes (required) |
| `-recursive` | `true`  | Watch directories recursively            |
| `-debounce`  | `300`   | Debounce interval in milliseconds        |

### Example Output

```
Watching "." for pattern "*.go" (recursive=true, debounce=300ms)
Command: go build ./...

[14:23:01] Changes detected:
  - main.go
  - watcher.go
[14:23:01] Running: go build ./...
[14:23:02] Command completed successfully
```


