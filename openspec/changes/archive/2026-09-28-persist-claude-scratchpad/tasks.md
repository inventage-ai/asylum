# Tasks

## 1. Base image

- [x] 1.1 In `assets/Dockerfile.core`, create `/run/user/${USER_ID}` next to `useradd`, owned by `${USER_ID}:${GROUP_ID}` with mode 0700. Verify by building the base image and running `stat -c '%a %u' /run/user/$(id -u)` in a container, which should print `700 <uid>`.

## 2. Container assembly

- [x] 2.1 Add a helper in `internal/container/container.go` that reports whether Claude is installed in the session (primary agent or companion). Verify with a table-driven test covering claude as primary, claude as companion, and no claude.
- [x] 2.2 In a new `claudeTempArgs`, called from `RunArgs`, when Claude is installed, create `~/.asylum/projects/<cname>/tmp`, resolve symlinks, and mount it at the identical path. Verify with tests for all three cases from 2.1, asserting that the mount is present or absent.
- [x] 2.3 In the same function, when Claude is installed, set `CLAUDE_CODE_TMPDIR` to the temp dir path and `XDG_RUNTIME_DIR` to `/run/user/<uid>`, and never set `TMPDIR`. Verify with tests asserting both env vars and the absence of `TMPDIR`, including a companion-only case.
- [x] 2.4 Run `go test ./...` and `go vet ./...`. Both should pass.

## 3. Documentation

- [x] 3.1 Add the `tmp/` project data subdirectory and its purpose to `docs/concepts/mounts.md`. Check `docs/commands/cleanup.md` and `assets/asylum-reference.md` and mention the directory where they list project data. Verify by reading the rendered pages.
- [x] 3.2 Add an **Added** entry to `CHANGELOG.md` under Unreleased. Verify that it names the persistence, the host-openable paths, and that `asylum cleanup` removes the directory.

## 4. End-to-end check

- [x] 4.1 Start a Claude container, have Claude write a file to its scratchpad, open that exact path on the host, restart the container, and confirm the file still exists. Also confirm `ls /run/user/$(id -u)/cc-socks` shows sockets there and not under the mount.
- [x] 4.2 Run `asylum cleanup` and confirm `~/.asylum/projects/<cname>/tmp` is gone.

## 5. Read-only cleanup

- [x] 5.1 Add a `removeTree` helper in `cmd/asylum/main.go` that retries `os.RemoveAll` once after granting owner write permission on the tree's directories, if the first attempt fails with a permission error. Verify with a test that removes a tree containing a 0555 directory with 0444 files.
- [x] 5.2 Use `removeTree` for the project data dir, the cache dir, and each entry in `removeProjectsDir`. Verify with `go test ./...` and `go vet ./...`.
