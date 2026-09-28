# Design

## Context

Claude Code 2.1.284 resolves its temp paths like this. The behaviour was read from the binary.

| Item | Resolution |
|---|---|
| Temp root | `$CLAUDE_CODE_TMPDIR`, falling back to `os.tmpdir()` (`$TMPDIR`, then `/tmp`) |
| Per-user dir | `<root>/claude-<uid>`, created 0700 |
| Safety checks | Owner must be the current uid; opened with `O_NOFOLLOW`, so symlinks are refused |
| Scratchpad | `<root>/claude-<uid>/<project-slug>/<session>/scratchpad` |
| Task output | `<root>/claude-<uid>/<project-slug>/<session>/tasks` |
| Edit diffs | `<root>/claude-<uid>/bash-edit-diff/` |
| Sockets | `$XDG_RUNTIME_DIR`, falling back to `<root>/claude-<uid>`, then `cc-socks/<pid>.sock`; falls back to `/tmp` when longer than 103 bytes |

Asylum already mounts per-project data from `~/.asylum/projects/<container>/`, for example `history/`. The container home matches the host home, so a mount at the real host path shows the same path on both sides. `Claude.EnvVars()` takes no arguments and cannot see the container name.

## Goals / Non-Goals

**Goals:**
- Claude's temp root lives on the host, one directory per project.
- A path the agent prints opens unchanged on the host.

**Non-Goals:**
- Persisting temp files of other agents or tools.
- Pruning or size limits.
- Supporting Claude Code builds that ignore `CLAUDE_CODE_TMPDIR`.

## Decisions

### `CLAUDE_CODE_TMPDIR` rather than `TMPDIR`
`TMPDIR` is read by every tool (`go build`, `go test`, npm, pytest, gradle). Pointing it at a virtiofs mount would slow them down, fill the host with junk, and put foreign sockets on a host filesystem. `CLAUDE_CODE_TMPDIR` moves only Claude's tree.

### Per-project directory at its real host path
The directory is `~/.asylum/projects/<container>/tmp`, and it is mounted at the same absolute path.
- Per-project keeps one container from reading another project's scratch files, and it matches `history/` and `credentials/`.
- The identical path makes printed paths valid on the host.
- A single global `~/.asylum/tmp` was rejected. Claude would separate projects by slug, but every container could read every project's files.
- Mounting the host's `/tmp/claude-<uid>` was rejected. Docker Desktop shares only the user dir by default, and the mount would expose host-side sessions too.

### Mount the directory, never symlink
Claude refuses a symlinked temp root. Resolve the host path with `filepath.EvalSymlinks` before mounting, as the agent config mount already does, so a symlinked `~/.asylum` still works.

### `XDG_RUNTIME_DIR=/run/user/<uid>`, created in the base image
- This keeps `cc-socks/` off virtiofs, where Unix socket support is unverified.
- The directory is created in `Dockerfile.core` next to `useradd`, using the `USER_ID` build arg, with owner `USER_ID:GROUP_ID` and mode 0700. The image build is the only place where root setup is free. The entrypoint must not install anything and runs as the user.
- The env var is set only when Claude is installed, which keeps the change scoped to Claude. The directory itself is harmless in every image.
- The uid in the env value comes from `os.Getuid()`, which is the same source as the `USER_ID` build arg.

### Gating lives in `container.go`
Claude is installed when `opts.Agent.Name() == "claude"` or `claude` is in `opts.Config.AgentCompanions(...)`. One helper answers this. A single `claudeTempArgs` function emits the mount and both env vars, because they only make sense together and `coreEnvVars` has no container name. Widening `Agent.EnvVars()` to take a container name was rejected because it would touch all five agents for one consumer.

### Cleanup grants write permission and retries
Tools such as the Go toolchain download write 0555 directories. `os.RemoveAll` cannot delete entries inside them, and inside the container virtiofs refuses the `chmod` that would fix it. Cleanup runs on the host, where `chmod` works. So cleanup tries `os.RemoveAll` first. Only on a permission error does it walk the tree, add owner `rwx` to each directory, and retry once. Files need no change, because deleting a file depends only on its parent directory. Chmod-ing everything up front was rejected because it would touch every file on the common path for nothing.

## Risks / Trade-offs

- [A future Claude build stops honouring `CLAUDE_CODE_TMPDIR` for the scratchpad] → Files fall back to the container `/tmp`, which is today's behaviour. Nothing breaks.
- [Scratch files grow without bound] → Accepted by design. Losing them is the problem this change fixes. `asylum cleanup` reclaims the space.
- [virtiofs is slower than the container `/tmp` for `bash-edit-diff` git repos] → The repos are small, and they are created once per session.
- [Rootless Docker or userns remapping makes the mounted dir's owner differ from the container uid] → Claude refuses the root with a clear error. This is out of scope because host-user alignment already assumes matching uids.
- [Base image hash changes] → Every base image rebuilds once on upgrade. This is expected for any `Dockerfile.core` edit.

- [Read-only trees in the temp root cannot be deleted from inside the container] → Accepted. Cleanup runs on the host and handles them.

## Migration Plan

No data migration. Existing sessions keep their old scratchpads in the container `/tmp` until the container restarts. New sessions write to the mount. Rollback means removing the mount and env vars, and the host directory then stays inert until `asylum cleanup`.
