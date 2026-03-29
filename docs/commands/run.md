# run

Run a command inside the container.

## Usage

```
asylum run <command> [args...]
asylum run -- <command> [args...]
```

## Description

Executes a one-off command in the container. If no container is running, one is started first. The `--` separator is optional.

## Examples

```sh
# Run a test suite
asylum run python test.py

# List files
asylum run ls -la

# Run with explicit separator
asylum run -- npm test

# Combine with flags
asylum -p 8080 run python -m http.server
```

## Notes

- The command runs as your host user (same username, UID, and home directory).
- The exit code from the command is forwarded to the host.
- The container persists after the command finishes if other sessions are still attached. If this was the only session, the container is removed.
