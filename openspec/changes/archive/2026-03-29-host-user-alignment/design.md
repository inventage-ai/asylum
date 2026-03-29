## Context

The container user is `claude` with UID/GID from the host but a hardcoded home directory `/home/claude`. This breaks absolute-path symlinks created by tools on the host (e.g., Claude Code stores symlinks in `~/.claude/` that point to `/Users/simon/.claude/...`). Inside the container, those paths resolve to nothing because the home dir is `/home/claude`, not `/Users/simon`.

There are 23+ hardcoded `/home/claude` references across Go code, kit definitions, entrypoint scripts, and Dockerfiles. Kit DockerSnippets also hardcode `USER claude` for root/user switching.

## Goals / Non-Goals

**Goals:**
- Container username matches host username (e.g., `simon` instead of `claude`)
- Container home directory matches host home directory (e.g., `/Users/simon` on macOS)
- Absolute-path symlinks from host tools work transparently inside the container
- All hardcoded `/home/claude` and `claude` username references eliminated
- Kit snippets use build args instead of hardcoded usernames
- Base image hash changes when host user or home dir changes (triggers rebuild)

**Non-Goals:**
- Config sharing modes (shared/isolated/project) — separate follow-up change
- Changing how agent config is mounted — that stays as-is for now

## Decisions

### 1. Host user identity as build args

```dockerfile
ARG USERNAME=claude
ARG USER_ID=1000
ARG GROUP_ID=1000
ARG USER_HOME=/home/claude
RUN useradd -m -d ${USER_HOME} -u ${USER_ID} -g ${GROUP_ID} -s /bin/zsh ${USERNAME}
```

Passed from Go:
```go
import "os/user"

u, _ := user.Current()
buildArgs["USERNAME"]  = u.Username
buildArgs["USER_ID"]   = u.Uid
buildArgs["GROUP_ID"]  = u.Gid
buildArgs["USER_HOME"] = u.HomeDir
```

The defaults (`claude`, `1000`, `/home/claude`) preserve backward compatibility but are never used in practice since the build always passes all four args.

### 2. Container home dir stored in a package-level variable

Rather than passing the home dir through every function, `container.go` resolves it once from `os.UserHomeDir()` (which returns the host home dir — same value that was baked into the image):

```go
func containerHome() string {
    home, _ := os.UserHomeDir()
    return home
}
```

This works because the Go binary runs on the host, and the host's home dir IS the container's home dir now.

### 3. Kit CacheDirs use tilde prefix

CacheDirs change from absolute paths to tilde-prefixed relative paths:

```go
// Before
CacheDirs: map[string]string{"maven": "/home/claude/.m2"}

// After
CacheDirs: map[string]string{"maven": "~/.m2"}
```

`container.go` resolves `~/` to the actual home dir when constructing volume mounts. This uses the existing `config.ExpandTilde` function.

### 4. Kit DockerSnippets use ${USERNAME} build arg

Kit snippets that switch to root and back use the build arg:

```dockerfile
# Before
USER root
RUN apt-get install -y maven
USER claude

# After
USER root
RUN apt-get install -y maven
USER ${USERNAME}
```

The `${USERNAME}` build arg is available because it's declared in `Dockerfile.core` and persists through all subsequent stages.

### 5. Agent ContainerConfigDir uses home-relative paths

```go
// Before
func (Claude) ContainerConfigDir() string { return "/home/claude/.claude" }

// After — resolved at runtime
func (Claude) ContainerConfigDir() string {
    home, _ := os.UserHomeDir()
    return filepath.Join(home, ".claude")
}
```

Similarly for env vars like `CLAUDE_CONFIG_DIR`.

### 6. Entrypoint uses $HOME throughout

All `/home/claude` references in `entrypoint.core` become `$HOME`:

```bash
# Before
if [ -d "/home/claude/.ssh" ]; then
    chmod 700 /home/claude/.ssh

# After
if [ -d "$HOME/.ssh" ]; then
    chmod 700 "$HOME/.ssh"
```

This is the most natural fix — `$HOME` is always set correctly by the shell.

### 7. Image hash includes home dir

The base image hash already includes the build args (USER_ID, GROUP_ID, USERNAME). Adding USER_HOME means a different host home dir triggers a rebuild. This is correct — the home dir path is baked into the image's user creation.

## Risks / Trade-offs

**Base image is now host-home-dir-specific** → It already was host-UID-specific. Teams sharing images would need matching home dir paths, but in practice teams on the same OS have the same home dir structure (`/Users/<name>` on macOS, `/home/<name>` on Linux).

**Existing images become stale** → The hash change triggers a rebuild automatically on next run. No manual intervention needed.

**Kit DockerSnippets must use ${USERNAME} not claude** → Simple mechanical change, but if a kit author forgets, the snippet breaks. The default value (`claude`) catches omissions during development but won't match in production.

**Container username visible in prompts/logs** → Users will see their own username in the shell prompt, `whoami`, etc. This is actually an improvement — it's less confusing than seeing `claude` everywhere.
