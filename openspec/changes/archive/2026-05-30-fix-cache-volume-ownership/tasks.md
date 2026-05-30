## 1. Implementation

- [x] 1.1 In `cmd/asylum/main.go`, extend the post-`RunDetached` chown block (currently at lines 302-308) to also iterate `cacheDirs` and run `docker.Exec(cname, "root", "chown", uid, dst)` for each cache mount point. Reuse the existing `uid := fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid())` value.
- [x] 1.2 Verify the chown remains inside the `freshContainer` block so it does not run on container reuse.

## 2. Verification

- [x] 2.1 Manually verify on a project with the `java/gradle` and `java/maven` kits: after `asylum --cleanup` and a fresh start, `ls -la ~/` inside the container shows `~/.gradle` and `~/.m2` owned by the host user (not `root`).
- [x] 2.2 Verify the agent (Claude) can write to `~/.gradle` (e.g., run `touch ~/.gradle/test-write` inside the container as the agent user).
- [x] 2.3 Run `go test ./...` to confirm no regressions.
- [x] 2.4 Run `go vet ./...` to confirm no vet warnings.

## 3. Documentation

- [x] 3.1 Add a CHANGELOG entry under **Unreleased → Fixed**: "Cache directory volumes (`~/.gradle`, `~/.m2`, `~/.npm`, `~/.cache/pip`) are now correctly owned by the container user, fixing agent write failures introduced when caches switched to named Docker volumes."
