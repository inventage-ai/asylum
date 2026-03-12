## Context

PLAN.md specifies shelling out to Docker CLI rather than using the Docker SDK. This package wraps the common operations needed by image management and container runtime.

## Goals / Non-Goals

**Goals:**
- `Build(contextDir, dockerfilePath string, tag string, labels map[string]string, buildArgs map[string]string, noCache bool) error`
- `InspectLabel(image, label string) (string, error)` — returns a single label value
- `RemoveImages(images ...string) error`
- `PruneImages(filterLabel string) error`
- `DockerAvailable() error` — checks that docker daemon is running

**Non-Goals:**
- No `docker run` assembly here — that's in the container package which constructs args and calls `syscall.Exec`
- No test coverage — purely a CLI passthrough layer

## Decisions

- **Thin functions**: Each function builds `[]string` args, runs `exec.Command("docker", args...)`, and returns the error. Output is forwarded to os.Stdout/os.Stderr for build progress.
- **InspectLabel**: Uses `docker inspect --format '{{index .Config.Labels "key"}}'` and returns trimmed stdout.
- **No struct**: Package-level functions. No state to encapsulate.

## Risks / Trade-offs

- Depends on `docker` being in PATH and the daemon running. `DockerAvailable()` checks this upfront.
