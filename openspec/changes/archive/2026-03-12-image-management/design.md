## Context

PLAN.md section 5.1 specifies the two-tier image strategy, hash-based rebuild detection, and project Dockerfile generation format.

## Goals / Non-Goals

**Goals:**
- `assets` package with `go:embed` for Dockerfile and entrypoint.sh
- `EnsureBase(version string, noCache bool) error` — builds base image if needed
- `EnsureProject(cfg config.Config, version string) (string, error)` — returns image tag to use
- Hash computation and comparison for rebuild detection
- Project Dockerfile generation from merged packages config

**Non-Goals:**
- Actual Dockerfile/entrypoint.sh content — that's the container-assets change
- Placeholder files are sufficient for now

## Decisions

- **Assets package**: `assets/assets.go` uses `go:embed` to embed `Dockerfile` and `entrypoint.sh`. Other packages import `assets.Dockerfile` and `assets.Entrypoint`.
- **Hash strategy**: SHA256 of concatenated Dockerfile + entrypoint.sh bytes. Compared against `asylum.hash` label on the image.
- **Temp dir for builds**: Write embedded assets to a temp dir, build from there, clean up. This is needed because `docker build` needs filesystem paths.
- **Project image tag**: `asylum:proj-<hash>` where hash is first 12 chars of SHA256 of the merged packages config YAML.
- **Rebuild cascade**: If base image is rebuilt, all project images are invalidated (they FROM asylum:latest).

## Risks / Trade-offs

- Embedding placeholder files means the image package compiles but won't produce a working Docker image until the container-assets change lands.
