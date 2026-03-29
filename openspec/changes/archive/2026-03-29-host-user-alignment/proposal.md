## Why

The container user is hardcoded as `claude` with home directory `/home/claude`. This creates problems when tools store absolute-path symlinks (e.g., Claude Code's `~/.claude/` folder), because those paths break inside the container where the home directory is different. It also causes confusion — users see a different identity inside the sandbox than on their host. Aligning the container user with the host user (same username, UID, GID, and home directory path) makes absolute symlinks work transparently and removes an entire class of path-mismatch bugs.

## What Changes

- **New `USER_HOME` build arg** in Dockerfile.core: `useradd -m -d $USER_HOME` instead of relying on default `/home/$USERNAME`
- **Pass host home dir** from `image.EnsureBase` via `os.UserHomeDir()`
- **Replace all 23+ hardcoded `/home/claude/` references** with `$HOME` (in shell scripts) or a runtime-resolved path (in Go code)
- **Kit CacheDirs** change from absolute paths (`/home/claude/.m2`) to `~/.m2` — resolved at container setup time using the host home dir
- **Kit DockerSnippets** that use `USER claude` / `usermod ... claude` switch to `USER ${USERNAME}` or use the build arg
- **Agent ContainerConfigDir** methods return paths relative to home dir, resolved at runtime
- **Entrypoint scripts** use `$HOME` instead of `/home/claude`
- **container.go** resolves all container paths using the host home dir instead of hardcoded `/home/claude`
- **Username matches host user** — passed as build arg from `os.UserInfo()`. The hardcoded `claude` username made sense when only Claude Code was supported, but with multiple agents it's misleading. The host username is more natural and avoids path assumptions.

## Capabilities

### New Capabilities
- `host-user-alignment`: Container user created with host's home directory path as build arg

### Modified Capabilities
- `profile-image-build`: Base image build passes USER_HOME; Dockerfile.core uses it for user creation
- `profile-container-setup`: All container volume mounts and env vars use dynamic home dir instead of hardcoded /home/claude

## Impact

- **assets/Dockerfile.core**: New `USER_HOME` ARG, `useradd -d $USER_HOME`, `WORKDIR $USER_HOME`
- **internal/image/image.go**: Pass `USER_HOME` build arg from `os.UserHomeDir()`
- **internal/container/container.go**: Replace all `/home/claude/` with dynamic path from host home dir; export `CONTAINER_HOME` or similar for use in volume/env resolution
- **internal/agent/*.go**: `ContainerConfigDir()` returns path using dynamic home dir
- **internal/kit/*.go**: CacheDirs use `~/` prefix resolved at runtime; DockerSnippets use `${USERNAME}` instead of `claude`
- **assets/entrypoint.core**: All `/home/claude` references become `$HOME`
- **assets/Dockerfile.tail**: `${USERNAME:-claude}` references become `${USERNAME}`
