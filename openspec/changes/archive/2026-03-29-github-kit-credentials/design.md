## Context

The credential system currently supports only generated-content mounts (`CredentialMount.Content`). Each mount's content is written to a staging directory and bind-mounted read-only. The GitHub kit needs to mount `~/.config/gh/hosts.yml` directly from the host — no content transformation needed.

## Goals / Non-Goals

**Goals:**
- GitHub `gh` CLI authenticated in containers via host config mount
- Extend `CredentialMount` to support direct host file mounting

**Non-Goals:**
- GitHub token generation or OAuth flow
- Explicit mode for GitHub (only one config file, auto is sufficient)

## Decisions

### Mount the directory, not individual files
Mount `~/.config/gh/` rather than `~/.config/gh/hosts.yml`. Tools commonly do atomic writes (write temp + rename), which replaces the file inode. A bind-mounted file would go stale, but a bind-mounted directory sees the new file immediately. This also means auth changes on the host are reflected in the container without restart.

### Add `HostPath` to `CredentialMount`
Adding `HostPath string` to `CredentialMount` enables pass-through mounting. When `HostPath` is set, the container code bind-mounts it directly instead of writing `Content` to a staging file. The two fields are mutually exclusive — a mount uses either `HostPath` or `Content`, not both.

This is preferable to reading the file into `Content` because: (1) it avoids copying credentials to disk in the staging area, and (2) it's the natural pattern for pass-through mounting that other kits may need.

### Skip silently when directory doesn't exist
If `~/.config/gh/` doesn't exist, the credential func returns empty — no error, no warning. The user simply hasn't authenticated gh on the host.
