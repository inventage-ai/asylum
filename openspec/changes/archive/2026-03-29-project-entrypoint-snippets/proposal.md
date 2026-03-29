## Why

Project-level kit `EntrypointSnippet`s are silently dropped. The entrypoint script is assembled only in `EnsureBase` from global kits. `EnsureProject` assembles `DockerSnippet`s but never `EntrypointSnippet`s. This means any kit enabled at the project level (e.g., cx in `.asylum`) gets its Dockerfile commands executed but its entrypoint logic (like bind-mounting rules files) never runs.

## What Changes

- `EnsureProject` assembles a project-level entrypoint script from project kit `EntrypointSnippet`s
- The project entrypoint is written into the project image and executed by the base entrypoint
- Project kit `BannerLines` are also included so project-only kits appear in the welcome banner

## Capabilities

### New Capabilities
- `project-entrypoint`: Generation and execution of a project-level entrypoint script from project kit snippets

### Modified Capabilities
- `image-build`: `EnsureProject` assembles and embeds a project entrypoint script when project kits have `EntrypointSnippet`s or `BannerLines`
- `container-image`: The base entrypoint executes the project entrypoint script if it exists

## Impact

- `internal/image/image.go`: `EnsureProject` and `generateProjectDockerfile` gain entrypoint assembly logic
- `assets/entrypoint.sh` (tail): Base entrypoint sources the project entrypoint if present
- Hash computation for project images must include entrypoint content
