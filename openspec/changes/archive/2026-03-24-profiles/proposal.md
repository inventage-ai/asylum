## Why

Asylum's Dockerfile, entrypoint, config, and container setup have grown into a monolith that installs and configures every supported language unconditionally. Adding a new language or tool means touching multiple unrelated files (Dockerfile, entrypoint.sh, container.go's `CacheDirs`, onboarding tasks, config defaults). This coupling makes the codebase harder to extend and forces every user to pay the image-size and build-time cost for languages they don't use. Profiles introduce a single abstraction that groups all language-specific concerns — installation, configuration, caching, onboarding, and entrypoint setup — into cohesive, self-contained units.

## What Changes

- New `internal/profile` package with a `Profile` struct and registry of built-in profiles
- Hierarchical profile system: top-level profiles (java, python, node) contain sub-profiles (maven, gradle, uv, npm, pnpm, yarn)
- Profiles hook into multiple lifecycle phases: image build (Dockerfile snippets), container setup (cache dirs, volumes, env vars), onboarding (task registration), entrypoint (shell snippets), and config (defaults)
- Decompose the monolithic `assets/Dockerfile` into core + profile snippets + tail — the image builder assembles them based on active profiles
- Decompose `assets/entrypoint.sh` similarly — profiles contribute entrypoint snippets
- `CacheDirs` constant in `container.go` becomes dynamic, aggregated from active profiles
- New `profiles` field in config YAML, settable at any layer (global, project, local, CLI)
- Global-config profiles are baked into the base image; project-config profiles into the project image
- `profiles: nil` (unspecified) defaults to all built-in profiles for backwards compatibility
- `profiles: []` (explicit empty) means core only — no language toolchains installed
- Existing `packages`, `versions`, `env`, `features`, `onboarding` config fields remain and override profile defaults

## Capabilities

### New Capabilities
- `profile-system`: Core profile abstraction — interface, registry, resolution of hierarchical profiles, activation semantics (nil=all, explicit selection, parent-implied-by-child)
- `profile-image-build`: Dockerfile decomposition into core/snippets/tail and assembly based on active profiles at the global and project image tiers
- `profile-entrypoint`: Entrypoint decomposition — profiles contribute shell snippets assembled into the final entrypoint
- `profile-container-setup`: Dynamic cache dirs, volumes, and env vars contributed by active profiles to container assembly
- `profile-config-integration`: Profile config defaults injected into the merge chain, overridable at any layer; `profiles` field in config YAML and CLI

### Modified Capabilities
- `container-image`: Image build now assembles Dockerfile from core + profile snippets + tail instead of using a monolithic embedded file
- `project-onboarding`: Onboarding tasks are now registered by profiles rather than hardcoded in main.go

## Impact

- **internal/profile/** (new): Profile interface, registry, built-in profiles (java, python, node + sub-profiles)
- **assets/Dockerfile**: Split into core and tail fragments; language-specific sections move into profile `DockerSnippet()` methods
- **assets/entrypoint.sh**: Split similarly; language-specific setup moves into profile `EntrypointSnippet()` methods
- **assets/assets.go**: Updated `go:embed` declarations for new file structure
- **internal/image/image.go**: `EnsureBase` and `EnsureProject` assemble Dockerfiles from profiles; hash computation includes profile snippets
- **internal/container/container.go**: `CacheDirs` replaced by dynamic aggregation from profiles; `appendVolumes` and `appendEnvVars` consume profile contributions
- **internal/config/config.go**: New `Profiles` field, parsing, merge semantics
- **cmd/asylum/main.go**: Profile resolution, passing resolved profiles to image/container/onboarding; CLI flag for `--profiles`
- **internal/onboarding/**: NPMTask moves into the node/npm profile; onboarding tasks sourced from active profiles
