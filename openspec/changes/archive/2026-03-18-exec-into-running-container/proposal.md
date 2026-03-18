## Why

When an agent is running in a container and you open a second terminal to run `asylum shell` or `asylum run <cmd>` in the same project directory, it fails with a Docker name conflict because it tries to `docker run` a new container with the same deterministic name. The only workaround is `docker exec` manually, which defeats the purpose of asylum.

## What Changes

- Detect when a container for the current project is already running
- For `shell`, `run`, and `shell --admin` modes: `docker exec` into the running container instead of `docker run`
- Agent mode (default, no subcommand) keeps the current behavior — starting a second agent in the same container would be confusing

## Capabilities

### New Capabilities

- `container-exec`: Exec into a running container for shell/run commands instead of failing with a name conflict

### Modified Capabilities

## Impact

- `internal/docker/docker.go`: Add function to check if a container is running by name
- `internal/container/container.go`: Add `ExecArgs` function that builds `docker exec` args
- `cmd/asylum/main.go`: Before building `docker run` args, check if container is running and mode is shell/run; if so, use `docker exec` instead
