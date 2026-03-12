## Why

Asylum uses a two-tier image strategy: a shared base image and optional per-project images with additional packages. The image package handles building, hash-based rebuild detection, and project image generation.

## What Changes

- Create `internal/image` package with base image build, project image generation, and auto-rebuild detection
- Embed Dockerfile and entrypoint.sh via `go:embed` in the assets package
- Hash computation for rebuild detection (SHA256 of Dockerfile + entrypoint.sh for base, packages config for project)
- Project Dockerfile generation from packages config

## Capabilities

### New Capabilities
- `image-build`: Two-tier image build with hash-based auto-rebuild detection and project Dockerfile generation

### Modified Capabilities

None.

## Impact

- Adds `internal/image/image.go` and `assets/assets.go`
- Assets directory will hold Dockerfile and entrypoint.sh (populated in the container-assets change)
- For now, creates placeholder asset files so the embed compiles
